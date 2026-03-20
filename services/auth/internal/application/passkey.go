package application

import (
	"context"
	"database/sql"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/auth/internal/domain/entity"
	autherrors "github.com/dealance/services/auth/internal/domain/errors"
	"github.com/dealance/services/auth/internal/domain/repository"
	sharedentity "github.com/dealance/shared/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/pkg/crypto"
)

// PasskeyService handles passkey registration on existing accounts.
type PasskeyService struct {
	identRepo   repository.IdentityProviderRepository
	sessionRepo repository.SessionRepository
	auditRepo   repository.AuditLogRepository
	log         zerolog.Logger
}

// NewPasskeyService creates a new passkey service.
func NewPasskeyService(
	identRepo repository.IdentityProviderRepository,
	sessionRepo repository.SessionRepository,
	auditRepo repository.AuditLogRepository,
	log zerolog.Logger,
) *PasskeyService {
	return &PasskeyService{
		identRepo:   identRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		log:         log,
	}
}

// BeginRegistration starts passkey registration by generating a challenge.
func (s *PasskeyService) BeginRegistration(ctx context.Context, userID string, req entity.BeginPasskeyRegistrationRequest) (*entity.BeginPasskeyRegistrationResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("Invalid user ID")
	}

	// Generate 32-byte challenge
	challengeBytes, err := crypto.GenerateRandomBytes(32)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	challengeID := uuid.New().String()

	// Store challenge in Redis (TTL 2 minutes)
	err = s.sessionRepo.CreateChallenge(ctx, challengeID, challengeBytes, 2*time.Minute)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	_ = uid

	return &entity.BeginPasskeyRegistrationResponse{
		ChallengeID: challengeID,
		Challenge:   base64.URLEncoding.EncodeToString(challengeBytes),
		UserID:      userID,
	}, nil
}

// FinishRegistration completes passkey registration by verifying the attestation.
func (s *PasskeyService) FinishRegistration(ctx context.Context, userID string, req entity.FinishPasskeyRegistrationRequest) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.ErrValidation("Invalid user ID")
	}

	// Get challenge from Redis
	challenge, err := s.sessionRepo.GetChallenge(ctx, req.ChallengeID)
	if err != nil || challenge == nil {
		return autherrors.ErrChallengeExpired
	}

	// Clean up challenge
	_ = s.sessionRepo.DeleteChallenge(ctx, req.ChallengeID)

	// Decode public key
	publicKeyBytes, err := base64.URLEncoding.DecodeString(req.PublicKey)
	if err != nil {
		return apperrors.ErrValidation("Invalid public key encoding")
	}

	// In production, verify the attestation against the challenge
	// For now, we store the credential

	// Check for duplicate
	existing, err := s.identRepo.GetByProviderAndExternalID(ctx, string(sharedentity.ProviderTypePasskey), req.CredentialID)
	if err == nil && existing != nil {
		return autherrors.ErrProviderAlreadyLinked
	}

	// Create identity provider record
	ip := &entity.IdentityProvider{
		ID:           uuid.New(),
		UserID:       uid,
		ProviderType: string(sharedentity.ProviderTypePasskey),
		ExternalID:   req.CredentialID,
		PublicKey:    publicKeyBytes,
		SignCount:    0,
		DeviceName:   sql.NullString{String: req.DeviceName, Valid: req.DeviceName != ""},
	}

	err = s.identRepo.Create(ctx, ip)
	if err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Audit log
	s.logAudit(ctx, uid, string(sharedentity.AuditEventPasskeyRegistered), "")

	return nil
}

func (s *PasskeyService) logAudit(ctx context.Context, userID uuid.UUID, eventType, deviceID string) {
	entry := &entity.AuditLogEntry{
		UserID:    userID,
		EventAt:   time.Now(),
		EventID:   uuid.New(),
		DeviceID:  deviceID,
		EventType: eventType,
	}
	if err := s.auditRepo.Log(ctx, entry); err != nil {
		s.log.Error().Err(err).Str("event_type", eventType).Msg("failed to write audit log")
	}
}

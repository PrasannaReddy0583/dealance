package application

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/auth/internal/domain/entity"
	autherrors "github.com/dealance/services/auth/internal/domain/errors"
	"github.com/dealance/services/auth/internal/domain/repository"
	sharedentity "github.com/dealance/shared/domain/entity"
	apperrors "github.com/dealance/shared/domain/errors"
	"github.com/dealance/shared/pkg/crypto"
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

// TokenService handles token refresh and logout.
type TokenService struct {
	userRepo    repository.UserRepository
	roleRepo    repository.UserRoleRepository
	refreshRepo repository.RefreshTokenRepository
	sessionRepo repository.SessionRepository
	auditRepo   repository.AuditLogRepository
	jwtIssuer   *dealjwt.Issuer
	jwtVerifier *dealjwt.Verifier
	log         zerolog.Logger
}

// NewTokenService creates a new token service.
func NewTokenService(
	userRepo repository.UserRepository,
	roleRepo repository.UserRoleRepository,
	refreshRepo repository.RefreshTokenRepository,
	sessionRepo repository.SessionRepository,
	auditRepo repository.AuditLogRepository,
	jwtIssuer *dealjwt.Issuer,
	jwtVerifier *dealjwt.Verifier,
	log zerolog.Logger,
) *TokenService {
	return &TokenService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		refreshRepo: refreshRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		jwtIssuer:   jwtIssuer,
		jwtVerifier: jwtVerifier,
		log:         log,
	}
}

// RefreshToken validates the refresh token, rotates it, and issues a new pair.
func (s *TokenService) RefreshToken(ctx context.Context, req entity.RefreshTokenRequest) (*entity.RefreshTokenResponse, error) {
	// Verify refresh token JWT
	claims, err := s.jwtVerifier.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, apperrors.ErrTokenExpired
	}

	// Get stored refresh token by hash
	tokenHash := crypto.HashSHA256([]byte(req.RefreshToken))
	storedToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, autherrors.ErrRefreshTokenNotFound
	}

	// Check if revoked
	if storedToken.IsRevoked {
		// Token reuse detected — revoke the entire family (security measure)
		s.log.Warn().
			Str("family_id", storedToken.FamilyID.String()).
			Str("user_id", storedToken.UserID.String()).
			Msg("refresh token reuse detected — revoking family")

		_ = s.refreshRepo.RevokeByFamilyID(ctx, storedToken.FamilyID, "TOKEN_REUSE")
		return nil, apperrors.ErrRefreshRevoked
	}

	// Check expiration
	if storedToken.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.ErrTokenExpired
	}

	// Revoke old token
	err = s.refreshRepo.RevokeByFamilyID(ctx, storedToken.FamilyID, "ROTATED")
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	userID := storedToken.UserID

	// Get user and roles for new access token
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	roles, err := s.roleRepo.GetActiveRoles(ctx, userID)
	if err != nil {
		roles = []string{}
	}

	activeRole := ""
	if len(roles) > 0 {
		activeRole = roles[0]
	}

	deviceID := ""
	if storedToken.DeviceID.Valid {
		deviceID = storedToken.DeviceID.String
	}

	// Issue new access token
	accessToken, accessClaims, err := s.jwtIssuer.IssueAccessToken(
		userID.String(),
		user.Email,
		roles,
		activeRole,
		"PENDING", // Re-fetch KYC status in production
		deviceID,
		user.EmailVerified,
	)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Issue new refresh token (same family)
	newRefreshToken, _, err := s.jwtIssuer.IssueRefreshToken(userID.String(), deviceID, storedToken.FamilyID.String())
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Store new refresh token
	newRTHash := crypto.HashSHA256([]byte(newRefreshToken))
	rt := &entity.RefreshTokenFamily{
		ID:        uuid.New(),
		UserID:    userID,
		FamilyID:  storedToken.FamilyID,
		TokenHash: newRTHash,
		DeviceID:  storedToken.DeviceID,
		IsRevoked: false,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(sharedentity.RefreshTokenTTLDays) * 24 * time.Hour),
	}

	err = s.refreshRepo.Create(ctx, rt)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Audit log
	s.logAudit(ctx, userID, string(sharedentity.AuditEventTokenRefreshed), deviceID)

	_ = claims // Validated above

	return &entity.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    accessClaims.ExpiresAt.Unix(),
		TokenType:    "Bearer",
	}, nil
}

// Logout revokes the refresh token family and blacklists the current JTI.
func (s *TokenService) Logout(ctx context.Context, userID uuid.UUID, jti string, req entity.LogoutRequest) error {
	// Blacklist current access token JTI
	if jti != "" {
		err := s.sessionRepo.BlacklistJTI(ctx, jti, 15*time.Minute)
		if err != nil {
			s.log.Error().Err(err).Msg("failed to blacklist JTI")
		}
	}

	// Revoke refresh token family if provided
	if req.RefreshToken != "" {
		tokenHash := crypto.HashSHA256([]byte(req.RefreshToken))
		storedToken, err := s.refreshRepo.GetByTokenHash(ctx, tokenHash)
		if err == nil && storedToken != nil {
			_ = s.refreshRepo.RevokeByFamilyID(ctx, storedToken.FamilyID, "LOGOUT")
		}
	} else {
		// Revoke all refresh tokens for this user
		_ = s.refreshRepo.RevokeByUserID(ctx, userID, "LOGOUT")
	}

	// Audit log
	s.logAudit(ctx, userID, string(sharedentity.AuditEventLogout), "")

	return nil
}

func (s *TokenService) logAudit(ctx context.Context, userID uuid.UUID, eventType, deviceID string) {
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

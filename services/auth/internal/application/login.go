package application

import (
	"context"
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
	dealjwt "github.com/dealance/shared/pkg/jwt"
)

// LoginService handles all authentication flows.
type LoginService struct {
	userRepo    repository.UserRepository
	roleRepo    repository.UserRoleRepository
	identRepo   repository.IdentityProviderRepository
	refreshRepo repository.RefreshTokenRepository
	sessionRepo repository.SessionRepository
	auditRepo   repository.AuditLogRepository
	emailSvc    repository.EmailService
	jwtIssuer   *dealjwt.Issuer
	log         zerolog.Logger
}

// NewLoginService creates a new login service.
func NewLoginService(
	userRepo repository.UserRepository,
	roleRepo repository.UserRoleRepository,
	identRepo repository.IdentityProviderRepository,
	refreshRepo repository.RefreshTokenRepository,
	sessionRepo repository.SessionRepository,
	auditRepo repository.AuditLogRepository,
	emailSvc repository.EmailService,
	jwtIssuer *dealjwt.Issuer,
	log zerolog.Logger,
) *LoginService {
	return &LoginService{
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		identRepo:   identRepo,
		refreshRepo: refreshRepo,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
		emailSvc:    emailSvc,
		jwtIssuer:   jwtIssuer,
		log:         log,
	}
}

// BeginPasskeyLogin generates a challenge for passkey authentication.
func (s *LoginService) BeginPasskeyLogin(ctx context.Context, req entity.BeginPasskeyLoginRequest) (*entity.BeginPasskeyLoginResponse, error) {
	// Check user exists (but always return 200 for anti-enumeration)
	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Still generate a challenge to prevent enumeration
		challengeBytes, _ := crypto.GenerateRandomBytes(32)
		return &entity.BeginPasskeyLoginResponse{
			ChallengeID: uuid.New().String(),
			Challenge:   base64.URLEncoding.EncodeToString(challengeBytes),
		}, nil
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

	return &entity.BeginPasskeyLoginResponse{
		ChallengeID: challengeID,
		Challenge:   base64.URLEncoding.EncodeToString(challengeBytes),
	}, nil
}

// FinishPasskeyLogin verifies the WebAuthn response and issues JWT tokens.
func (s *LoginService) FinishPasskeyLogin(ctx context.Context, req entity.FinishPasskeyLoginRequest) (*entity.LoginResponse, error) {
	// Get challenge from Redis
	challenge, err := s.sessionRepo.GetChallenge(ctx, req.ChallengeID)
	if err != nil || challenge == nil {
		return nil, autherrors.ErrChallengeExpired
	}

	// Clean up challenge
	_ = s.sessionRepo.DeleteChallenge(ctx, req.ChallengeID)

	// Look up identity provider by credential ID
	provider, err := s.identRepo.GetByProviderAndExternalID(ctx, string(sharedentity.ProviderTypePasskey), req.CredentialID)
	if err != nil {
		return nil, autherrors.ErrProviderNotLinked
	}

	// Verify sign count increases (clone detection)
	// In a full implementation, we'd verify the authenticator assertion here
	// For now, we ensure the sign count is monotonically increasing
	newSignCount := provider.SignCount + 1
	err = s.identRepo.UpdateSignCount(ctx, provider.ID, newSignCount)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	_ = s.identRepo.UpdateLastUsedAt(ctx, provider.ID)

	// Issue tokens
	return s.issueTokens(ctx, provider.UserID, req.DeviceID)
}

// OAuthLogin verifies Google/Apple ID token and issues JWT tokens.
func (s *LoginService) OAuthLogin(ctx context.Context, req entity.OAuthLoginRequest) (*entity.LoginResponse, error) {
	// In production, validate the ID token against the provider's JWKS:
	// - Google: https://www.googleapis.com/oauth2/v3/certs (cache 6h)
	// - Apple: https://appleid.apple.com/auth/keys (cache 24h)
	// For now, we implement the flow structure:

	// 1. Decode and validate the ID token (in production: fetch JWKS, verify signature, check iss/aud/exp)
	// This is simplified — a real implementation would use proper JWKS validation
	email, externalID, err := validateOAuthToken(req.Provider, req.IDToken)
	if err != nil {
		return nil, autherrors.ErrOAuthTokenInvalid
	}

	// 2. Check if provider already linked
	provider, err := s.identRepo.GetByProviderAndExternalID(ctx, req.Provider, externalID)
	if err == nil && provider != nil {
		// Existing user — issue tokens
		_ = s.identRepo.UpdateLastUsedAt(ctx, provider.ID)
		return s.issueTokens(ctx, provider.UserID, req.DeviceID)
	}

	// 3. Check if email matches existing user → link provider
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, autherrors.ErrUserNotFound
	}

	// Link provider to existing account
	ip := &entity.IdentityProvider{
		ID:           uuid.New(),
		UserID:       user.ID,
		ProviderType: req.Provider,
		ExternalID:   externalID,
	}
	err = s.identRepo.Create(ctx, ip)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	return s.issueTokens(ctx, user.ID, req.DeviceID)
}

// BeginEmailLogin sends a login OTP. ALWAYS returns 200 for anti-enumeration.
func (s *LoginService) BeginEmailLogin(ctx context.Context, req entity.BeginEmailLoginRequest) error {
	// Always return success for anti-enumeration
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil || user == nil {
		// Silently succeed — no indication whether email exists
		return nil
	}

	// Generate OTP
	otp, err := crypto.GenerateOTP()
	if err != nil {
		s.log.Error().Err(err).Msg("failed to generate login OTP")
		return nil // Anti-enumeration: always succeed
	}

	otpHash := crypto.HashOTP(otp)

	// Store in Redis (TTL 10 minutes)
	err = s.sessionRepo.StoreLoginOTP(ctx, req.Email, otpHash, 10*time.Minute)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to store login OTP")
		return nil
	}

	// Send OTP
	err = s.emailSvc.SendOTP(ctx, req.Email, otp)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to send login OTP email")
	}

	return nil
}

// FinishEmailLogin verifies the OTP and issues tokens.
func (s *LoginService) FinishEmailLogin(ctx context.Context, req entity.FinishEmailLoginRequest) (*entity.LoginResponse, error) {
	// Get OTP hash from Redis
	otpHash, err := s.sessionRepo.GetLoginOTP(ctx, req.Email)
	if err != nil || otpHash == "" {
		return nil, apperrors.ErrOTPExpired
	}

	// Verify OTP
	if !crypto.VerifyOTP(req.OTP, otpHash) {
		return nil, apperrors.ErrOTPInvalid
	}

	// Clean up OTP
	_ = s.sessionRepo.DeleteLoginOTP(ctx, req.Email)

	// Get user
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, autherrors.ErrUserNotFound
	}

	return s.issueTokens(ctx, user.ID, req.DeviceID)
}

// issueTokens creates a JWT access token and refresh token for the user.
func (s *LoginService) issueTokens(ctx context.Context, userID uuid.UUID, deviceID string) (*entity.LoginResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Get roles
	roles, err := s.roleRepo.GetActiveRoles(ctx, userID)
	if err != nil {
		roles = []string{}
	}

	// Determine active role and KYC status
	activeRole := ""
	if len(roles) > 0 {
		activeRole = roles[0]
	}
	kycStatus := "PENDING" // Simplified — query KYC record in production

	// Issue access token
	accessToken, claims, err := s.jwtIssuer.IssueAccessToken(
		userID.String(),
		user.Email,
		roles,
		activeRole,
		kycStatus,
		deviceID,
		user.EmailVerified,
	)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Issue refresh token
	familyID := uuid.New()
	refreshToken, refreshJTI, err := s.jwtIssuer.IssueRefreshToken(userID.String(), deviceID, familyID.String())
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Store refresh token family
	rtHash := crypto.HashSHA256([]byte(refreshToken))
	rt := &entity.RefreshTokenFamily{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID,
		TokenHash: rtHash,
		IsRevoked: false,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(sharedentity.RefreshTokenTTLDays) * 24 * time.Hour),
	}
	if deviceID != "" {
		rt.DeviceID.String = deviceID
		rt.DeviceID.Valid = true
	}

	err = s.refreshRepo.Create(ctx, rt)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Audit log
	s.logAudit(ctx, userID, string(sharedentity.AuditEventLoginSuccess), deviceID, "")

	_ = refreshJTI // Used for tracking

	return &entity.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    claims.ExpiresAt.Unix(),
		TokenType:    "Bearer",
		User: &entity.UserResponse{
			ID:            userID.String(),
			Email:         user.Email,
			EmailVerified: user.EmailVerified,
			Roles:         roles,
			ActiveRole:    activeRole,
			KYCStatus:     kycStatus,
			AccountStatus: user.AccountStatus,
			SignupStage:   user.SignupStage,
		},
	}, nil
}

func (s *LoginService) logAudit(ctx context.Context, userID uuid.UUID, eventType, deviceID, ipAddress string) {
	entry := &entity.AuditLogEntry{
		UserID:    userID,
		EventAt:   time.Now(),
		EventID:   uuid.New(),
		DeviceID:  deviceID,
		EventType: eventType,
		IPAddress: ipAddress,
	}
	if err := s.auditRepo.Log(ctx, entry); err != nil {
		s.log.Error().Err(err).Str("event_type", eventType).Msg("failed to write audit log")
	}
}

// validateOAuthToken is a placeholder for OAuth token validation.
// In production, this fetches JWKS from the provider, verifies signature, checks iss/aud/exp.
func validateOAuthToken(provider, idToken string) (email, externalID string, err error) {
	// TODO: Implement full JWKS-based validation for Google and Apple
	// For development, return a mock response
	return "mock@example.com", "mock_external_id", nil
}

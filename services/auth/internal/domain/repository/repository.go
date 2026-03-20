package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dealance/services/auth/internal/domain/entity"
)

// UserRepository handles user CRUD operations.
type UserRepository interface {
	Create(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateSignupStage(ctx context.Context, id uuid.UUID, stage string) error
	UpdateAccountStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateEmailVerified(ctx context.Context, id uuid.UUID) error
	UpdateCountryCode(ctx context.Context, id uuid.UUID, countryCode string) error
}

// UserRoleRepository handles user role operations.
type UserRoleRepository interface {
	Create(ctx context.Context, userID uuid.UUID, role string) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.UserRole, error)
	GetActiveRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// IdentityProviderRepository handles identity provider operations.
type IdentityProviderRepository interface {
	Create(ctx context.Context, ip *entity.IdentityProvider) error
	GetByProviderAndExternalID(ctx context.Context, providerType, externalID string) (*entity.IdentityProvider, error)
	GetByUserIDAndProvider(ctx context.Context, userID uuid.UUID, providerType string) (*entity.IdentityProvider, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.IdentityProvider, error)
	UpdateSignCount(ctx context.Context, id uuid.UUID, signCount int64) error
	UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error
}

// DeviceAttestationRepository handles device attestation records.
type DeviceAttestationRepository interface {
	Create(ctx context.Context, da *entity.DeviceAttestation) error
	GetByDeviceID(ctx context.Context, deviceID string) (*entity.DeviceAttestation, error)
	GetSigningKey(ctx context.Context, deviceID string) (string, error)
}

// KYCRepository handles KYC verification records.
type KYCRepository interface {
	Create(ctx context.Context, record *entity.KYCRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.KYCRecord, error)
	GetByUserIDAndType(ctx context.Context, userID uuid.UUID, kycType string) ([]entity.KYCRecord, error)
	GetLatestByUserIDAndType(ctx context.Context, userID uuid.UUID, kycType string) (*entity.KYCRecord, error)
	CountAttempts(ctx context.Context, userID uuid.UUID, kycType string) (int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, scores *KYCScores) error
	UpdateVendorSession(ctx context.Context, id uuid.UUID, vendorSessionID string) error
}

// KYCScores holds the verification scores from the KYC vendor.
type KYCScores struct {
	FaceMatchScore  *float64
	LivenessScore   *float64
	DeepfakeScore   *float64
	RejectionReason *string
}

// InvestorVerificationRepository handles investor accreditation.
type InvestorVerificationRepository interface {
	Create(ctx context.Context, iv *entity.InvestorVerification) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.InvestorVerification, error)
	Update(ctx context.Context, iv *entity.InvestorVerification) error
}

// RefreshTokenRepository handles refresh token families.
type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *entity.RefreshTokenFamily) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshTokenFamily, error)
	RevokeByFamilyID(ctx context.Context, familyID uuid.UUID, reason string) error
	RevokeByUserID(ctx context.Context, userID uuid.UUID, reason string) error
	IsRevoked(ctx context.Context, tokenHash string) (bool, error)
}

// AuditLogRepository handles security audit log (ScyllaDB).
type AuditLogRepository interface {
	Log(ctx context.Context, entry *entity.AuditLogEntry) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]entity.AuditLogEntry, error)
}

// SessionRepository handles Redis session data (signup sessions, challenges).
type SessionRepository interface {
	CreateSignupSession(ctx context.Context, sessionID, email, otpHash string, ttl time.Duration) error
	GetSignupSession(ctx context.Context, sessionID string) (email string, otpHash string, err error)
	DeleteSignupSession(ctx context.Context, sessionID string) error
	UpdateSignupSessionOTP(ctx context.Context, sessionID, otpHash string) error

	// Challenge sessions for passkey login
	CreateChallenge(ctx context.Context, challengeID string, challenge []byte, ttl time.Duration) error
	GetChallenge(ctx context.Context, challengeID string) ([]byte, error)
	DeleteChallenge(ctx context.Context, challengeID string) error

	// OTP for email login
	StoreLoginOTP(ctx context.Context, email, otpHash string, ttl time.Duration) error
	GetLoginOTP(ctx context.Context, email string) (string, error)
	DeleteLoginOTP(ctx context.Context, email string) error

	// JTI blacklist for logout
	BlacklistJTI(ctx context.Context, jti string, ttl time.Duration) error
	IsJTIBlacklisted(ctx context.Context, jti string) (bool, error)

	// JWKS cache for OAuth
	CacheJWKS(ctx context.Context, provider string, jwksData []byte, ttl time.Duration) error
	GetCachedJWKS(ctx context.Context, provider string) ([]byte, error)
}

// EmailService handles sending emails.
type EmailService interface {
	SendOTP(ctx context.Context, email, otp string) error
	SendWelcome(ctx context.Context, email string) error
	SendKYCApproved(ctx context.Context, email string) error
	SendKYCRejected(ctx context.Context, email, reason string) error
}

// KYCVendorService handles KYC vendor API calls.
type KYCVendorService interface {
	InitiateSession(ctx context.Context, userID string, kycType string) (sessionID, sdkToken string, err error)
	VerifyWebhookSignature(payload []byte, signature string) bool
}

package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the auth service.
type User struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	Email         string         `db:"email" json:"email"`
	Phone         sql.NullString `db:"phone" json:"phone,omitempty"`
	EmailVerified bool           `db:"email_verified" json:"email_verified"`
	PhoneVerified bool           `db:"phone_verified" json:"phone_verified"`
	CountryCode   sql.NullString `db:"country_code" json:"country_code,omitempty"`
	AccountStatus string         `db:"account_status" json:"account_status"`
	SignupStage   string         `db:"signup_stage" json:"signup_stage"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

// UserRole represents a role assigned to a user.
type UserRole struct {
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Role      string    `db:"role" json:"role"`
	Verified  bool      `db:"verified" json:"verified"`
	Active    bool      `db:"active" json:"active"`
	GrantedAt time.Time `db:"granted_at" json:"granted_at"`
}

// IdentityProvider represents a linked authentication provider.
type IdentityProvider struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	UserID       uuid.UUID      `db:"user_id" json:"user_id"`
	ProviderType string         `db:"provider_type" json:"provider_type"`
	ExternalID   string         `db:"external_id" json:"external_id"`
	PublicKey    []byte         `db:"public_key" json:"-"`
	SignCount    int64          `db:"sign_count" json:"sign_count"`
	DeviceName   sql.NullString `db:"device_name" json:"device_name,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	LastUsedAt   sql.NullTime   `db:"last_used_at" json:"last_used_at,omitempty"`
}

// DeviceAttestation represents a verified device.
type DeviceAttestation struct {
	ID               uuid.UUID      `db:"id" json:"id"`
	UserID           uuid.UUID      `db:"user_id" json:"user_id"`
	DeviceID         string         `db:"device_id" json:"device_id"`
	Platform         string         `db:"platform" json:"platform"`
	SigningKeyPublic sql.NullString `db:"signing_key_public" json:"-"`
	AttestationHash  sql.NullString `db:"attestation_hash" json:"-"`
	RiskLevel        string         `db:"risk_level" json:"risk_level"`
	IsValid          bool           `db:"is_valid" json:"is_valid"`
	VerifiedAt       time.Time      `db:"verified_at" json:"verified_at"`
	ExpiresAt        sql.NullTime   `db:"expires_at" json:"expires_at,omitempty"`
}

// KYCRecord represents a KYC verification attempt.
type KYCRecord struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	UserID          uuid.UUID       `db:"user_id" json:"user_id"`
	KYCType         string          `db:"kyc_type" json:"kyc_type"`
	Vendor          sql.NullString  `db:"vendor" json:"vendor,omitempty"`
	VendorSessionID sql.NullString  `db:"vendor_session_id" json:"vendor_session_id,omitempty"`
	DocumentType    sql.NullString  `db:"document_type" json:"document_type,omitempty"`
	DocumentHash    sql.NullString  `db:"document_hash" json:"-"`
	Status          string          `db:"status" json:"status"`
	FaceMatchScore  sql.NullFloat64 `db:"face_match_score" json:"face_match_score,omitempty"`
	LivenessScore   sql.NullFloat64 `db:"liveness_score" json:"liveness_score,omitempty"`
	DeepfakeScore   sql.NullFloat64 `db:"deepfake_score" json:"deepfake_score,omitempty"`
	RejectionReason sql.NullString  `db:"rejection_reason" json:"rejection_reason,omitempty"`
	AttemptNumber   int             `db:"attempt_number" json:"attempt_number"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	CompletedAt     sql.NullTime    `db:"completed_at" json:"completed_at,omitempty"`
}

// InvestorVerification represents full investor accreditation records.
type InvestorVerification struct {
	ID                    uuid.UUID      `db:"id" json:"id"`
	UserID                uuid.UUID      `db:"user_id" json:"user_id"`
	BankAccountVerified   bool           `db:"bank_account_verified" json:"bank_account_verified"`
	BankAccountHash       sql.NullString `db:"bank_account_hash" json:"-"`
	BankAccountLast4      sql.NullString `db:"bank_account_last4" json:"bank_account_last4,omitempty"`
	BankIFSC              sql.NullString `db:"bank_ifsc" json:"bank_ifsc,omitempty"`
	BankName              sql.NullString `db:"bank_name" json:"bank_name,omitempty"`
	PennyDropVerified     bool           `db:"penny_drop_verified" json:"penny_drop_verified"`
	PANVerified           bool           `db:"pan_verified" json:"pan_verified"`
	NetWorthPaise         sql.NullInt64  `db:"net_worth_paise" json:"net_worth_paise,omitempty"`
	AnnualIncomePaise     sql.NullInt64  `db:"annual_income_paise" json:"annual_income_paise,omitempty"`
	AccreditationDocType  sql.NullString `db:"accreditation_doc_type" json:"accreditation_doc_type,omitempty"`
	AccreditationVerified bool           `db:"accreditation_verified" json:"accreditation_verified"`
	DematAccountID        sql.NullString `db:"demat_account_id" json:"demat_account_id,omitempty"`
	Depository            sql.NullString `db:"depository" json:"depository,omitempty"`
	BrokerageVerified     bool           `db:"brokerage_verified" json:"brokerage_verified"`
	VerificationStatus    string         `db:"verification_status" json:"verification_status"`
	VerifiedAt            sql.NullTime   `db:"verified_at" json:"verified_at,omitempty"`
	ExpiresAt             sql.NullTime   `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt             time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time      `db:"updated_at" json:"updated_at"`
}

// RefreshTokenFamily tracks refresh token rotation chains.
type RefreshTokenFamily struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	UserID      uuid.UUID      `db:"user_id" json:"user_id"`
	FamilyID    uuid.UUID      `db:"family_id" json:"family_id"`
	TokenHash   string         `db:"token_hash" json:"-"`
	DeviceID    sql.NullString `db:"device_id" json:"device_id,omitempty"`
	IsRevoked   bool           `db:"is_revoked" json:"is_revoked"`
	RevokeReason sql.NullString `db:"revoke_reason" json:"revoke_reason,omitempty"`
	IssuedAt    time.Time      `db:"issued_at" json:"issued_at"`
	ExpiresAt   time.Time      `db:"expires_at" json:"expires_at"`
}

// AuditLogEntry represents a security audit event (ScyllaDB).
type AuditLogEntry struct {
	UserID    uuid.UUID `json:"user_id"`
	EventAt   time.Time `json:"event_at"`
	EventID   uuid.UUID `json:"event_id"`
	DeviceID  string    `json:"device_id"`
	EventType string    `json:"event_type"`
	EventData string    `json:"event_data"`
	IPAddress string    `json:"ip_address"`
	RiskScore float64   `json:"risk_score"`
}

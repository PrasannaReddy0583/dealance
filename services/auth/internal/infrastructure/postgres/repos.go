package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dealance/services/auth/internal/domain/entity"
	"github.com/dealance/services/auth/internal/domain/repository"
)

// UserRepo implements UserRepository using PostgreSQL.
type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	err := r.db.QueryRowxContext(ctx,
		`INSERT INTO users (email) VALUES ($1) RETURNING id, email, phone, email_verified, phone_verified, country_code, account_status, signup_stage, created_at, updated_at`,
		email,
	).StructScan(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user := &entity.User{}
	err := r.db.GetContext(ctx, user,
		`SELECT id, email, phone, email_verified, phone_verified, country_code, account_status, signup_stage, created_at, updated_at FROM users WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	user := &entity.User{}
	err := r.db.GetContext(ctx, user,
		`SELECT id, email, phone, email_verified, phone_verified, country_code, account_status, signup_stage, created_at, updated_at FROM users WHERE email = $1`,
		email,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`,
		email,
	)
	return exists, err
}

func (r *UserRepo) UpdateSignupStage(ctx context.Context, id uuid.UUID, stage string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET signup_stage = $1 WHERE id = $2`,
		stage, id,
	)
	return err
}

func (r *UserRepo) UpdateAccountStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET account_status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *UserRepo) UpdateEmailVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET email_verified = TRUE WHERE id = $1`,
		id,
	)
	return err
}

func (r *UserRepo) UpdateCountryCode(ctx context.Context, id uuid.UUID, countryCode string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET country_code = $1 WHERE id = $2`,
		countryCode, id,
	)
	return err
}

// --- UserRoleRepo ---

type UserRoleRepo struct {
	db *sqlx.DB
}

func NewUserRoleRepo(db *sqlx.DB) *UserRoleRepo {
	return &UserRoleRepo{db: db}
}

func (r *UserRoleRepo) Create(ctx context.Context, userID uuid.UUID, role string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_roles (user_id, role) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, role,
	)
	return err
}

func (r *UserRoleRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.UserRole, error) {
	var roles []entity.UserRole
	err := r.db.SelectContext(ctx, &roles,
		`SELECT user_id, role, verified, active, granted_at FROM user_roles WHERE user_id = $1`,
		userID,
	)
	return roles, err
}

func (r *UserRoleRepo) GetActiveRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var roles []string
	err := r.db.SelectContext(ctx, &roles,
		`SELECT role FROM user_roles WHERE user_id = $1 AND active = TRUE`,
		userID,
	)
	return roles, err
}

// --- IdentityProviderRepo ---

type IdentityProviderRepo struct {
	db *sqlx.DB
}

func NewIdentityProviderRepo(db *sqlx.DB) *IdentityProviderRepo {
	return &IdentityProviderRepo{db: db}
}

func (r *IdentityProviderRepo) Create(ctx context.Context, ip *entity.IdentityProvider) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO identity_providers (id, user_id, provider_type, external_id, public_key, sign_count, device_name)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		ip.ID, ip.UserID, ip.ProviderType, ip.ExternalID, ip.PublicKey, ip.SignCount, ip.DeviceName,
	)
	return err
}

func (r *IdentityProviderRepo) GetByProviderAndExternalID(ctx context.Context, providerType, externalID string) (*entity.IdentityProvider, error) {
	ip := &entity.IdentityProvider{}
	err := r.db.GetContext(ctx, ip,
		`SELECT id, user_id, provider_type, external_id, public_key, sign_count, device_name, created_at, last_used_at
		 FROM identity_providers WHERE provider_type = $1 AND external_id = $2`,
		providerType, externalID,
	)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *IdentityProviderRepo) GetByUserIDAndProvider(ctx context.Context, userID uuid.UUID, providerType string) (*entity.IdentityProvider, error) {
	ip := &entity.IdentityProvider{}
	err := r.db.GetContext(ctx, ip,
		`SELECT id, user_id, provider_type, external_id, public_key, sign_count, device_name, created_at, last_used_at
		 FROM identity_providers WHERE user_id = $1 AND provider_type = $2 LIMIT 1`,
		userID, providerType,
	)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (r *IdentityProviderRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.IdentityProvider, error) {
	var providers []entity.IdentityProvider
	err := r.db.SelectContext(ctx, &providers,
		`SELECT id, user_id, provider_type, external_id, public_key, sign_count, device_name, created_at, last_used_at
		 FROM identity_providers WHERE user_id = $1`,
		userID,
	)
	return providers, err
}

func (r *IdentityProviderRepo) UpdateSignCount(ctx context.Context, id uuid.UUID, signCount int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE identity_providers SET sign_count = $1 WHERE id = $2`,
		signCount, id,
	)
	return err
}

func (r *IdentityProviderRepo) UpdateLastUsedAt(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE identity_providers SET last_used_at = NOW() WHERE id = $1`,
		id,
	)
	return err
}

// --- DeviceAttestationRepo ---

type DeviceAttestationRepo struct {
	db *sqlx.DB
}

func NewDeviceAttestationRepo(db *sqlx.DB) *DeviceAttestationRepo {
	return &DeviceAttestationRepo{db: db}
}

func (r *DeviceAttestationRepo) Create(ctx context.Context, da *entity.DeviceAttestation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO device_attestations (id, user_id, device_id, platform, signing_key_public, attestation_hash, risk_level, is_valid, verified_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		da.ID, da.UserID, da.DeviceID, da.Platform, da.SigningKeyPublic, da.AttestationHash,
		da.RiskLevel, da.IsValid, da.VerifiedAt, da.ExpiresAt,
	)
	return err
}

func (r *DeviceAttestationRepo) GetByDeviceID(ctx context.Context, deviceID string) (*entity.DeviceAttestation, error) {
	da := &entity.DeviceAttestation{}
	err := r.db.GetContext(ctx, da,
		`SELECT id, user_id, device_id, platform, signing_key_public, attestation_hash, risk_level, is_valid, verified_at, expires_at
		 FROM device_attestations WHERE device_id = $1 AND is_valid = TRUE`,
		deviceID,
	)
	if err != nil {
		return nil, err
	}
	return da, nil
}

func (r *DeviceAttestationRepo) GetSigningKey(ctx context.Context, deviceID string) (string, error) {
	var key string
	err := r.db.GetContext(ctx, &key,
		`SELECT signing_key_public FROM device_attestations WHERE device_id = $1 AND is_valid = TRUE`,
		deviceID,
	)
	return key, err
}

// --- KYCRepo ---

type KYCRepo struct {
	db *sqlx.DB
}

func NewKYCRepo(db *sqlx.DB) *KYCRepo {
	return &KYCRepo{db: db}
}

func (r *KYCRepo) Create(ctx context.Context, record *entity.KYCRecord) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO kyc_records (id, user_id, kyc_type, vendor, vendor_session_id, document_type, status, attempt_number, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		record.ID, record.UserID, record.KYCType, record.Vendor, record.VendorSessionID,
		record.DocumentType, record.Status, record.AttemptNumber, record.CreatedAt,
	)
	return err
}

func (r *KYCRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.KYCRecord, error) {
	record := &entity.KYCRecord{}
	err := r.db.GetContext(ctx, record,
		`SELECT id, user_id, kyc_type, vendor, vendor_session_id, document_type, document_hash, status,
		        face_match_score, liveness_score, deepfake_score, rejection_reason, attempt_number, created_at, completed_at
		 FROM kyc_records WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *KYCRepo) GetByUserIDAndType(ctx context.Context, userID uuid.UUID, kycType string) ([]entity.KYCRecord, error) {
	var records []entity.KYCRecord
	err := r.db.SelectContext(ctx, &records,
		`SELECT id, user_id, kyc_type, vendor, vendor_session_id, document_type, document_hash, status,
		        face_match_score, liveness_score, deepfake_score, rejection_reason, attempt_number, created_at, completed_at
		 FROM kyc_records WHERE user_id = $1 AND kyc_type = $2 ORDER BY created_at DESC`,
		userID, kycType,
	)
	return records, err
}

func (r *KYCRepo) GetLatestByUserIDAndType(ctx context.Context, userID uuid.UUID, kycType string) (*entity.KYCRecord, error) {
	record := &entity.KYCRecord{}
	err := r.db.GetContext(ctx, record,
		`SELECT id, user_id, kyc_type, vendor, vendor_session_id, document_type, document_hash, status,
		        face_match_score, liveness_score, deepfake_score, rejection_reason, attempt_number, created_at, completed_at
		 FROM kyc_records WHERE user_id = $1 AND kyc_type = $2 ORDER BY created_at DESC LIMIT 1`,
		userID, kycType,
	)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *KYCRepo) CountAttempts(ctx context.Context, userID uuid.UUID, kycType string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM kyc_records WHERE user_id = $1 AND kyc_type = $2`,
		userID, kycType,
	)
	return count, err
}

func (r *KYCRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, scores *repository.KYCScores) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE kyc_records SET status = $1, face_match_score = $2, liveness_score = $3, deepfake_score = $4,
		        rejection_reason = $5, completed_at = $6 WHERE id = $7`,
		status, scores.FaceMatchScore, scores.LivenessScore, scores.DeepfakeScore,
		scores.RejectionReason, now, id,
	)
	return err
}

func (r *KYCRepo) UpdateVendorSession(ctx context.Context, id uuid.UUID, vendorSessionID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE kyc_records SET vendor_session_id = $1 WHERE id = $2`,
		vendorSessionID, id,
	)
	return err
}

// --- InvestorVerificationRepo ---

type InvestorVerificationRepo struct {
	db *sqlx.DB
}

func NewInvestorVerificationRepo(db *sqlx.DB) *InvestorVerificationRepo {
	return &InvestorVerificationRepo{db: db}
}

func (r *InvestorVerificationRepo) Create(ctx context.Context, iv *entity.InvestorVerification) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO investor_verifications (id, user_id) VALUES ($1, $2)`,
		iv.ID, iv.UserID,
	)
	return err
}

func (r *InvestorVerificationRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.InvestorVerification, error) {
	iv := &entity.InvestorVerification{}
	err := r.db.GetContext(ctx, iv,
		`SELECT * FROM investor_verifications WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return iv, nil
}

func (r *InvestorVerificationRepo) Update(ctx context.Context, iv *entity.InvestorVerification) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE investor_verifications SET
			bank_account_verified = $1, bank_account_hash = $2, bank_account_last4 = $3,
			bank_ifsc = $4, bank_name = $5, penny_drop_verified = $6, pan_verified = $7,
			net_worth_paise = $8, annual_income_paise = $9, accreditation_doc_type = $10,
			accreditation_verified = $11, demat_account_id = $12, depository = $13,
			brokerage_verified = $14, verification_status = $15, verified_at = $16, expires_at = $17
		 WHERE user_id = $18`,
		iv.BankAccountVerified, iv.BankAccountHash, iv.BankAccountLast4,
		iv.BankIFSC, iv.BankName, iv.PennyDropVerified, iv.PANVerified,
		iv.NetWorthPaise, iv.AnnualIncomePaise, iv.AccreditationDocType,
		iv.AccreditationVerified, iv.DematAccountID, iv.Depository,
		iv.BrokerageVerified, iv.VerificationStatus, iv.VerifiedAt, iv.ExpiresAt,
		iv.UserID,
	)
	return err
}

// --- RefreshTokenRepo ---

type RefreshTokenRepo struct {
	db *sqlx.DB
}

func NewRefreshTokenRepo(db *sqlx.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (r *RefreshTokenRepo) Create(ctx context.Context, rt *entity.RefreshTokenFamily) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refresh_token_families (id, user_id, family_id, token_hash, device_id, is_revoked, issued_at, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		rt.ID, rt.UserID, rt.FamilyID, rt.TokenHash, rt.DeviceID, rt.IsRevoked, rt.IssuedAt, rt.ExpiresAt,
	)
	return err
}

func (r *RefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshTokenFamily, error) {
	rt := &entity.RefreshTokenFamily{}
	err := r.db.GetContext(ctx, rt,
		`SELECT id, user_id, family_id, token_hash, device_id, is_revoked, revoke_reason, issued_at, expires_at
		 FROM refresh_token_families WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return nil, err
	}
	return rt, nil
}

func (r *RefreshTokenRepo) RevokeByFamilyID(ctx context.Context, familyID uuid.UUID, reason string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_token_families SET is_revoked = TRUE, revoke_reason = $1 WHERE family_id = $2`,
		reason, familyID,
	)
	return err
}

func (r *RefreshTokenRepo) RevokeByUserID(ctx context.Context, userID uuid.UUID, reason string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refresh_token_families SET is_revoked = TRUE, revoke_reason = $1 WHERE user_id = $2 AND is_revoked = FALSE`,
		reason, userID,
	)
	return err
}

func (r *RefreshTokenRepo) IsRevoked(ctx context.Context, tokenHash string) (bool, error) {
	var isRevoked bool
	err := r.db.GetContext(ctx, &isRevoked,
		`SELECT is_revoked FROM refresh_token_families WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil // Token not found = treat as revoked
		}
		return false, err
	}
	return isRevoked, nil
}

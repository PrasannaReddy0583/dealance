-- Auth Service Schema
-- PostgreSQL 16

CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          VARCHAR(255) UNIQUE NOT NULL,
    phone          VARCHAR(20) UNIQUE,
    email_verified BOOLEAN DEFAULT FALSE,
    phone_verified BOOLEAN DEFAULT FALSE,
    country_code   VARCHAR(2),
    account_status VARCHAR(20) DEFAULT 'PENDING',
    signup_stage   VARCHAR(30) DEFAULT 'EMAIL_VERIFY',
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_account_status ON users(account_status);
CREATE INDEX idx_users_signup_stage ON users(signup_stage);

CREATE TABLE user_roles (
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(20) NOT NULL,
    verified   BOOLEAN DEFAULT FALSE,
    active     BOOLEAN DEFAULT TRUE,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role)
);

CREATE INDEX idx_user_roles_role ON user_roles(role);

CREATE TABLE identity_providers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID REFERENCES users(id) ON DELETE CASCADE,
    provider_type VARCHAR(20) NOT NULL,
    external_id   TEXT NOT NULL,
    public_key    BYTEA,
    sign_count    BIGINT DEFAULT 0,
    device_name   VARCHAR(255),
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    last_used_at  TIMESTAMPTZ,
    UNIQUE(provider_type, external_id)
);

CREATE INDEX idx_identity_providers_user_id ON identity_providers(user_id);
CREATE INDEX idx_identity_providers_provider_external ON identity_providers(provider_type, external_id);

CREATE TABLE device_attestations (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID REFERENCES users(id),
    device_id          TEXT UNIQUE NOT NULL,
    platform           VARCHAR(10) NOT NULL,
    signing_key_public TEXT,
    attestation_hash   TEXT,
    risk_level         VARCHAR(10) DEFAULT 'LOW',
    is_valid           BOOLEAN DEFAULT TRUE,
    verified_at        TIMESTAMPTZ DEFAULT NOW(),
    expires_at         TIMESTAMPTZ
);

CREATE INDEX idx_device_attestations_user_id ON device_attestations(user_id);
CREATE INDEX idx_device_attestations_device_id ON device_attestations(device_id);

CREATE TABLE kyc_records (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID REFERENCES users(id) ON DELETE CASCADE,
    kyc_type          VARCHAR(30) NOT NULL,
    vendor            VARCHAR(30),
    vendor_session_id TEXT,
    document_type     VARCHAR(30),
    document_hash     TEXT,
    status            VARCHAR(20) DEFAULT 'PENDING',
    face_match_score  DECIMAL(5,4),
    liveness_score    DECIMAL(5,4),
    deepfake_score    DECIMAL(5,4),
    rejection_reason  TEXT,
    attempt_number    INT DEFAULT 1,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    completed_at      TIMESTAMPTZ
);

CREATE INDEX idx_kyc_records_user_id ON kyc_records(user_id);
CREATE INDEX idx_kyc_records_status ON kyc_records(status);
CREATE INDEX idx_kyc_records_user_type ON kyc_records(user_id, kyc_type);

CREATE TABLE investor_verifications (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                UUID UNIQUE REFERENCES users(id),
    bank_account_verified  BOOLEAN DEFAULT FALSE,
    bank_account_hash      TEXT,
    bank_account_last4     VARCHAR(4),
    bank_ifsc              VARCHAR(20),
    bank_name              VARCHAR(100),
    penny_drop_verified    BOOLEAN DEFAULT FALSE,
    pan_verified           BOOLEAN DEFAULT FALSE,
    net_worth_paise        BIGINT,
    annual_income_paise    BIGINT,
    accreditation_doc_type VARCHAR(50),
    accreditation_verified BOOLEAN DEFAULT FALSE,
    demat_account_id       VARCHAR(50),
    depository             VARCHAR(10),
    brokerage_verified     BOOLEAN DEFAULT FALSE,
    verification_status    VARCHAR(20) DEFAULT 'PENDING',
    verified_at            TIMESTAMPTZ,
    expires_at             DATE,
    created_at             TIMESTAMPTZ DEFAULT NOW(),
    updated_at             TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_investor_verifications_user_id ON investor_verifications(user_id);
CREATE INDEX idx_investor_verifications_status ON investor_verifications(verification_status);

CREATE TABLE refresh_token_families (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID REFERENCES users(id) ON DELETE CASCADE,
    family_id     UUID NOT NULL,
    token_hash    TEXT NOT NULL UNIQUE,
    device_id     TEXT,
    is_revoked    BOOLEAN DEFAULT FALSE,
    revoke_reason VARCHAR(50),
    issued_at     TIMESTAMPTZ DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_refresh_token_families_user_id ON refresh_token_families(user_id);
CREATE INDEX idx_refresh_token_families_family_id ON refresh_token_families(family_id);
CREATE INDEX idx_refresh_token_families_token_hash ON refresh_token_families(token_hash);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_investor_verifications_updated_at BEFORE UPDATE ON investor_verifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

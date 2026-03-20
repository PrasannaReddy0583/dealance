-- Auth Service Schema Rollback

DROP TRIGGER IF EXISTS update_investor_verifications_updated_at ON investor_verifications;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS refresh_token_families;
DROP TABLE IF EXISTS investor_verifications;
DROP TABLE IF EXISTS kyc_records;
DROP TABLE IF EXISTS device_attestations;
DROP TABLE IF EXISTS identity_providers;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;

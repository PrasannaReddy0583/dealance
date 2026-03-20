-- User Service Schema Rollback

DROP TRIGGER IF EXISTS update_investor_profiles_updated_at ON investor_profiles;
DROP TRIGGER IF EXISTS update_entrepreneur_profiles_updated_at ON entrepreneur_profiles;
DROP TRIGGER IF EXISTS update_profile_media_updated_at ON profile_media;
DROP TRIGGER IF EXISTS update_user_settings_updated_at ON user_settings;
DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;

DROP TABLE IF EXISTS investor_profiles;
DROP TABLE IF EXISTS entrepreneur_profiles;
DROP TABLE IF EXISTS profile_media;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS blocked_users;
DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS profiles;

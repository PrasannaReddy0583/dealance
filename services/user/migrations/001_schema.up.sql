-- User Service Schema
-- PostgreSQL 16

-- User profiles (the social identity)
CREATE TABLE profiles (
    id              UUID PRIMARY KEY,  -- same as auth users.id
    username        VARCHAR(30) UNIQUE NOT NULL,
    display_name    VARCHAR(100) NOT NULL,
    bio             TEXT DEFAULT '',
    avatar_url      TEXT DEFAULT '',
    cover_url       TEXT DEFAULT '',
    location        VARCHAR(100) DEFAULT '',
    website         VARCHAR(255) DEFAULT '',
    linkedin_url    VARCHAR(255) DEFAULT '',
    twitter_url     VARCHAR(255) DEFAULT '',
    date_of_birth   DATE,
    gender          VARCHAR(20),
    profession      VARCHAR(100) DEFAULT '',
    company         VARCHAR(100) DEFAULT '',
    experience_years INT DEFAULT 0,
    is_public       BOOLEAN DEFAULT TRUE,
    is_verified     BOOLEAN DEFAULT FALSE,
    follower_count  INT DEFAULT 0,
    following_count INT DEFAULT 0,
    post_count      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_profiles_username ON profiles(username);
CREATE INDEX idx_profiles_display_name ON profiles(display_name);
CREATE INDEX idx_profiles_location ON profiles(location);
CREATE INDEX idx_profiles_profession ON profiles(profession);
CREATE INDEX idx_profiles_is_public ON profiles(is_public);

-- Follow graph
CREATE TABLE follows (
    follower_id  UUID NOT NULL,
    following_id UUID NOT NULL,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id)
);

CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);
CREATE INDEX idx_follows_created ON follows(created_at DESC);

-- Blocked users
CREATE TABLE blocked_users (
    blocker_id UUID NOT NULL,
    blocked_id UUID NOT NULL,
    reason     VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id)
);

CREATE INDEX idx_blocked_blocker ON blocked_users(blocker_id);
CREATE INDEX idx_blocked_blocked ON blocked_users(blocked_id);

-- User settings
CREATE TABLE user_settings (
    user_id                  UUID PRIMARY KEY,
    notification_push        BOOLEAN DEFAULT TRUE,
    notification_email       BOOLEAN DEFAULT TRUE,
    notification_sms         BOOLEAN DEFAULT FALSE,
    notification_deal_updates BOOLEAN DEFAULT TRUE,
    notification_new_followers BOOLEAN DEFAULT TRUE,
    notification_messages    BOOLEAN DEFAULT TRUE,
    privacy_show_email       BOOLEAN DEFAULT FALSE,
    privacy_show_phone       BOOLEAN DEFAULT FALSE,
    privacy_show_location    BOOLEAN DEFAULT TRUE,
    privacy_allow_messages   VARCHAR(20) DEFAULT 'EVERYONE',  -- EVERYONE, FOLLOWERS, NOBODY
    privacy_show_investments BOOLEAN DEFAULT FALSE,
    feed_content_language    VARCHAR(10) DEFAULT 'en',
    feed_sort_preference     VARCHAR(20) DEFAULT 'ALGORITHMIC', -- ALGORITHMIC, CHRONOLOGICAL
    theme                    VARCHAR(10) DEFAULT 'SYSTEM',      -- LIGHT, DARK, SYSTEM
    updated_at               TIMESTAMPTZ DEFAULT NOW()
);

-- Profile media (portfolio items, certifications, etc.)
CREATE TABLE profile_media (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    media_type  VARCHAR(20) NOT NULL,  -- PORTFOLIO, CERTIFICATION, AWARD, PRESS
    title       VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    media_url   TEXT NOT NULL,
    thumbnail_url TEXT,
    display_order INT DEFAULT 0,
    is_visible  BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_profile_media_user ON profile_media(user_id);
CREATE INDEX idx_profile_media_type ON profile_media(user_id, media_type);

-- Entrepreneur-specific profile data
CREATE TABLE entrepreneur_profiles (
    user_id          UUID PRIMARY KEY,
    startup_count    INT DEFAULT 0,
    total_raised_paise BIGINT DEFAULT 0,
    sectors          TEXT[] DEFAULT '{}',
    skills           TEXT[] DEFAULT '{}',
    education        JSONB DEFAULT '[]',
    work_history     JSONB DEFAULT '[]',
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

-- Investor-specific profile data
CREATE TABLE investor_profiles (
    user_id              UUID PRIMARY KEY,
    investor_type        VARCHAR(30),  -- ANGEL, VC, FAMILY_OFFICE, HNI, SYNDICATE
    investment_range_min_paise BIGINT DEFAULT 0,
    investment_range_max_paise BIGINT DEFAULT 0,
    preferred_sectors    TEXT[] DEFAULT '{}',
    preferred_stages     TEXT[] DEFAULT '{}',  -- PRE_SEED, SEED, SERIES_A, etc.
    portfolio_count      INT DEFAULT 0,
    total_invested_paise BIGINT DEFAULT 0,
    investment_thesis    TEXT DEFAULT '',
    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_settings_updated_at BEFORE UPDATE ON user_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_profile_media_updated_at BEFORE UPDATE ON profile_media
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_entrepreneur_profiles_updated_at BEFORE UPDATE ON entrepreneur_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_investor_profiles_updated_at BEFORE UPDATE ON investor_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

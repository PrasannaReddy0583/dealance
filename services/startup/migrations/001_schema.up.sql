-- Startup Service Schema
-- PostgreSQL 16

-- Startups (company profiles)
CREATE TABLE startups (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    founder_id      UUID NOT NULL,
    name            VARCHAR(200) NOT NULL,
    slug            VARCHAR(200) UNIQUE NOT NULL,
    tagline         VARCHAR(300),
    description     TEXT DEFAULT '',
    logo_url        TEXT DEFAULT '',
    cover_url       TEXT DEFAULT '',
    website         VARCHAR(255),
    founded_year    INT,
    headquarters    VARCHAR(100),
    country         VARCHAR(50),
    sector          VARCHAR(50) NOT NULL,          -- FINTECH, HEALTHTECH, EDTECH, etc.
    stage           VARCHAR(30) DEFAULT 'IDEA',    -- IDEA, MVP, EARLY_TRACTION, GROWTH, SCALE
    business_model  VARCHAR(30),                   -- B2B, B2C, B2B2C, MARKETPLACE, SAAS
    team_size       INT DEFAULT 1,
    incorporation_type VARCHAR(30),                -- PRIVATE_LTD, LLP, PARTNERSHIP, SOLE_PROP
    cin_number      VARCHAR(50),                   -- Corporate Identification Number (MCA India)
    gstin           VARCHAR(20),                   -- GSTIN
    status          VARCHAR(20) DEFAULT 'ACTIVE',  -- ACTIVE, PAUSED, ACQUIRED, CLOSED
    is_verified     BOOLEAN DEFAULT FALSE,
    is_featured     BOOLEAN DEFAULT FALSE,
    view_count      INT DEFAULT 0,
    follower_count  INT DEFAULT 0,
    tags            TEXT[] DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_startups_founder ON startups(founder_id);
CREATE INDEX idx_startups_slug ON startups(slug);
CREATE INDEX idx_startups_sector ON startups(sector);
CREATE INDEX idx_startups_stage ON startups(stage);
CREATE INDEX idx_startups_country ON startups(country);
CREATE INDEX idx_startups_status ON startups(status);
CREATE INDEX idx_startups_tags ON startups USING GIN(tags);

-- Funding rounds
CREATE TABLE funding_rounds (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    startup_id      UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    round_type      VARCHAR(30) NOT NULL,     -- PRE_SEED, SEED, SERIES_A, SERIES_B, BRIDGE, DEBT
    amount_paise    BIGINT NOT NULL,           -- Amount in paise (INR × 100)
    valuation_paise BIGINT,                    -- Pre-money valuation in paise
    currency        VARCHAR(3) DEFAULT 'INR',
    status          VARCHAR(20) DEFAULT 'OPEN', -- OPEN, CLOSED, CANCELLED
    target_paise    BIGINT,                    -- Target raise amount
    min_ticket_paise BIGINT,                   -- Minimum investment
    equity_offered  DECIMAL(5,2),              -- Percentage equity offered
    instrument_type VARCHAR(30),               -- EQUITY, CCPS, CCD, SAFE, CONVERTIBLE_NOTE
    open_date       DATE,
    close_date      DATE,
    description     TEXT,
    terms_url       TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_funding_startup ON funding_rounds(startup_id);
CREATE INDEX idx_funding_type ON funding_rounds(round_type);
CREATE INDEX idx_funding_status ON funding_rounds(status);

-- Team members
CREATE TABLE team_members (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    startup_id  UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    user_id     UUID,            -- NULL if not on platform
    name        VARCHAR(100) NOT NULL,
    role        VARCHAR(100) NOT NULL,
    title       VARCHAR(100),
    bio         TEXT DEFAULT '',
    avatar_url  TEXT,
    linkedin_url VARCHAR(255),
    is_founder  BOOLEAN DEFAULT FALSE,
    equity_pct  DECIMAL(5,2),
    joined_date DATE,
    display_order INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_team_startup ON team_members(startup_id);
CREATE INDEX idx_team_user ON team_members(user_id);

-- Startup media (pitch deck, product screenshots, demo videos)
CREATE TABLE startup_media (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    startup_id  UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    media_type  VARCHAR(30) NOT NULL,  -- PITCH_DECK, SCREENSHOT, VIDEO, DOCUMENT, LOGO
    title       VARCHAR(200),
    media_url   TEXT NOT NULL,
    thumbnail_url TEXT,
    file_size   BIGINT,
    is_public   BOOLEAN DEFAULT TRUE,  -- Some media may require NDA
    display_order INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_startup_media ON startup_media(startup_id);

-- Key metrics (MRR, users, growth)
CREATE TABLE startup_metrics (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    startup_id  UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    metric_type VARCHAR(30) NOT NULL,  -- MRR, ARR, USERS, DAU, MAU, REVENUE, BURN_RATE, RUNWAY_MONTHS
    value       DECIMAL(20,2) NOT NULL,
    currency    VARCHAR(3),
    period      VARCHAR(10),           -- 2024-Q1, 2024-01, etc.
    recorded_at TIMESTAMPTZ DEFAULT NOW(),
    is_verified BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_metrics_startup ON startup_metrics(startup_id);
CREATE INDEX idx_metrics_type ON startup_metrics(startup_id, metric_type);

-- Startup followers
CREATE TABLE startup_follows (
    user_id     UUID NOT NULL,
    startup_id  UUID NOT NULL REFERENCES startups(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, startup_id)
);

CREATE INDEX idx_startup_follows_startup ON startup_follows(startup_id);

-- Triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_startups_updated_at BEFORE UPDATE ON startups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_funding_updated_at BEFORE UPDATE ON funding_rounds
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

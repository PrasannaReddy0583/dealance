CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    role VARCHAR(20) NOT NULL DEFAULT 'MODERATOR',
    permissions TEXT[] DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES admin_users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(30),
    entity_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_audit_admin ON admin_audit_log(admin_id, created_at DESC);

CREATE TABLE platform_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stat_date DATE NOT NULL UNIQUE,
    total_users INT DEFAULT 0,
    active_users INT DEFAULT 0,
    total_startups INT DEFAULT 0,
    total_deals INT DEFAULT 0,
    total_invested_paise BIGINT DEFAULT 0,
    total_posts INT DEFAULT 0,
    new_users_today INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_stats_date ON platform_stats(stat_date DESC);

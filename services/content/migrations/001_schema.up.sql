-- Content Service Schema
-- PostgreSQL 16

-- Posts (text, shorts, long videos)
CREATE TABLE posts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id       UUID NOT NULL,
    post_type       VARCHAR(20) NOT NULL DEFAULT 'TEXT',  -- TEXT, SHORT, VIDEO, ARTICLE
    title           VARCHAR(300),
    body            TEXT DEFAULT '',
    visibility      VARCHAR(20) DEFAULT 'PUBLIC',  -- PUBLIC, FOLLOWERS, PRIVATE
    is_published    BOOLEAN DEFAULT TRUE,
    is_pinned       BOOLEAN DEFAULT FALSE,
    allow_comments  BOOLEAN DEFAULT TRUE,
    hashtags        TEXT[] DEFAULT '{}',
    mention_ids     UUID[] DEFAULT '{}',
    view_count      INT DEFAULT 0,
    like_count      INT DEFAULT 0,
    comment_count   INT DEFAULT 0,
    share_count     INT DEFAULT 0,
    save_count      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_type ON posts(post_type);
CREATE INDEX idx_posts_created ON posts(created_at DESC);
CREATE INDEX idx_posts_author_created ON posts(author_id, created_at DESC);
CREATE INDEX idx_posts_visibility ON posts(visibility);
CREATE INDEX idx_posts_hashtags ON posts USING GIN(hashtags);

-- Post media attachments (images, videos, thumbnails)
CREATE TABLE post_media (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    media_type  VARCHAR(20) NOT NULL,  -- IMAGE, VIDEO, THUMBNAIL
    media_url   TEXT NOT NULL,
    thumbnail_url TEXT,
    width       INT,
    height      INT,
    duration_ms INT,        -- video duration in milliseconds
    file_size   BIGINT,     -- bytes
    mime_type   VARCHAR(50),
    display_order INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_post_media_post ON post_media(post_id);

-- Comments
CREATE TABLE comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id     UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL,
    parent_id   UUID REFERENCES comments(id) ON DELETE CASCADE,  -- NULL = top-level
    body        TEXT NOT NULL,
    like_count  INT DEFAULT 0,
    reply_count INT DEFAULT 0,
    is_edited   BOOLEAN DEFAULT FALSE,
    is_deleted  BOOLEAN DEFAULT FALSE,  -- soft delete for thread integrity
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON comments(post_id, created_at);
CREATE INDEX idx_comments_author ON comments(author_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);

-- Reactions (likes, celebrates, insightful, etc.)
CREATE TABLE reactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL,
    target_id   UUID NOT NULL,              -- post_id OR comment_id
    target_type VARCHAR(10) NOT NULL,       -- POST, COMMENT
    reaction_type VARCHAR(20) DEFAULT 'LIKE',  -- LIKE, CELEBRATE, INSIGHTFUL, LOVE, CURIOUS
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, target_id, target_type)
);

CREATE INDEX idx_reactions_target ON reactions(target_id, target_type);
CREATE INDEX idx_reactions_user ON reactions(user_id);

-- Saved posts (bookmarks)
CREATE TABLE saved_posts (
    user_id     UUID NOT NULL,
    post_id     UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    collection  VARCHAR(50) DEFAULT 'default',
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, post_id)
);

CREATE INDEX idx_saved_posts_user ON saved_posts(user_id, created_at DESC);

-- Hashtags (for trending and discovery)
CREATE TABLE hashtags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) UNIQUE NOT NULL,
    post_count  INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_hashtags_name ON hashtags(name);
CREATE INDEX idx_hashtags_count ON hashtags(post_count DESC);

-- Reports (content moderation)
CREATE TABLE reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL,
    target_id   UUID NOT NULL,
    target_type VARCHAR(10) NOT NULL,  -- POST, COMMENT
    reason      VARCHAR(50) NOT NULL,  -- SPAM, HARASSMENT, MISINFORMATION, NSFW, OTHER
    description TEXT,
    status      VARCHAR(20) DEFAULT 'PENDING',  -- PENDING, REVIEWED, ACTIONED, DISMISSED
    reviewed_by UUID,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_reports_target ON reports(target_id, target_type);
CREATE INDEX idx_reports_status ON reports(status);

-- Triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_reports_updated_at BEFORE UPDATE ON reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

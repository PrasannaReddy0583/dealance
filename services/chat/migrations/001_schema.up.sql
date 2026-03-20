-- Chat Service Schema
-- PostgreSQL 16

-- Conversations (DM or group)
CREATE TABLE conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conv_type       VARCHAR(20) DEFAULT 'DM',     -- DM, GROUP, DEAL_ROOM
    title           VARCHAR(200),
    deal_id         UUID,                          -- For deal room conversations
    created_by      UUID NOT NULL,
    last_message_at TIMESTAMPTZ DEFAULT NOW(),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_conv_type ON conversations(conv_type);
CREATE INDEX idx_conv_deal ON conversations(deal_id);

-- Conversation participants
CREATE TABLE conversation_participants (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    role            VARCHAR(20) DEFAULT 'MEMBER',  -- ADMIN, MEMBER, OBSERVER
    unread_count    INT DEFAULT 0,
    muted           BOOLEAN DEFAULT FALSE,
    last_read_at    TIMESTAMPTZ DEFAULT NOW(),
    joined_at       TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX idx_participants_user ON conversation_participants(user_id);

-- Messages
CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL,
    message_type    VARCHAR(20) DEFAULT 'TEXT',   -- TEXT, IMAGE, FILE, AUDIO, VIDEO, SYSTEM
    body            TEXT NOT NULL,
    media_url       TEXT,
    media_type      VARCHAR(50),
    file_name       VARCHAR(200),
    file_size       BIGINT,
    reply_to_id     UUID REFERENCES messages(id),
    is_edited       BOOLEAN DEFAULT FALSE,
    is_deleted      BOOLEAN DEFAULT FALSE,
    metadata        JSONB DEFAULT '{}',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_conv ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_sender ON messages(sender_id);

-- Read receipts
CREATE TABLE read_receipts (
    message_id  UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    read_at     TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (message_id, user_id)
);

-- Message reactions (emoji)
CREATE TABLE message_reactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id  UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id     UUID NOT NULL,
    emoji       VARCHAR(10) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);

CREATE INDEX idx_msg_reactions ON message_reactions(message_id);

-- Triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ language 'plpgsql';

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

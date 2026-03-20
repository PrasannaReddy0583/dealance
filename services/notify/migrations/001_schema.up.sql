CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    notif_type VARCHAR(30) NOT NULL,
    title VARCHAR(200) NOT NULL,
    body TEXT NOT NULL,
    channel VARCHAR(20) NOT NULL DEFAULT 'PUSH',
    entity_type VARCHAR(30),
    entity_id UUID,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'PENDING',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_notif_user ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notif_status ON notifications(status);

CREATE TABLE device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    device_token TEXT NOT NULL UNIQUE,
    platform VARCHAR(10) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_device_user ON device_tokens(user_id);

CREATE TABLE notification_preferences (
    user_id UUID PRIMARY KEY,
    push_enabled BOOLEAN DEFAULT TRUE,
    email_enabled BOOLEAN DEFAULT TRUE,
    sms_enabled BOOLEAN DEFAULT FALSE,
    deal_updates BOOLEAN DEFAULT TRUE,
    chat_messages BOOLEAN DEFAULT TRUE,
    content_reactions BOOLEAN DEFAULT TRUE,
    new_followers BOOLEAN DEFAULT TRUE,
    marketing BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

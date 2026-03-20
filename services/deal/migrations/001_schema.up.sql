-- Deal Service Schema
-- PostgreSQL 16

-- Deals (investment deal rooms)
CREATE TABLE deals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    startup_id      UUID NOT NULL,
    funding_round_id UUID,
    title           VARCHAR(300) NOT NULL,
    description     TEXT DEFAULT '',
    deal_type       VARCHAR(30) NOT NULL,     -- EQUITY, CCPS, CCD, SAFE, CONVERTIBLE_NOTE, DEBT
    status          VARCHAR(30) DEFAULT 'DRAFT',  -- DRAFT, OPEN, IN_PROGRESS, DUE_DILIGENCE, CLOSING, CLOSED, CANCELLED
    amount_paise    BIGINT NOT NULL,          -- Total deal size in paise
    min_ticket_paise BIGINT,                  -- Minimum investment per participant
    max_participants INT DEFAULT 50,
    equity_pct      DECIMAL(5,2),
    valuation_paise BIGINT,                   -- Pre-money valuation
    currency        VARCHAR(3) DEFAULT 'INR',
    terms_summary   TEXT,
    requires_nda    BOOLEAN DEFAULT TRUE,
    requires_kyc    BOOLEAN DEFAULT TRUE,
    created_by      UUID NOT NULL,
    open_date       TIMESTAMPTZ,
    close_date      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_deals_startup ON deals(startup_id);
CREATE INDEX idx_deals_status ON deals(status);
CREATE INDEX idx_deals_type ON deals(deal_type);
CREATE INDEX idx_deals_created_by ON deals(created_by);

-- Deal participants (investors in the deal)
CREATE TABLE deal_participants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id         UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    role            VARCHAR(20) DEFAULT 'INVESTOR', -- INVESTOR, LEAD_INVESTOR, ADVISOR, OBSERVER
    status          VARCHAR(20) DEFAULT 'INTERESTED', -- INTERESTED, NDA_SIGNED, DUE_DILIGENCE, COMMITTED, INVESTED, DECLINED, REMOVED
    commitment_paise BIGINT,                -- Committed investment amount
    invested_paise  BIGINT DEFAULT 0,       -- Actual invested amount
    equity_pct      DECIMAL(5,2),
    nda_signed_at   TIMESTAMPTZ,
    committed_at    TIMESTAMPTZ,
    invested_at     TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(deal_id, user_id)
);

CREATE INDEX idx_participants_deal ON deal_participants(deal_id);
CREATE INDEX idx_participants_user ON deal_participants(user_id);
CREATE INDEX idx_participants_status ON deal_participants(status);

-- Deal documents (term sheets, SHA, due diligence docs)
CREATE TABLE deal_documents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id         UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    uploaded_by     UUID NOT NULL,
    doc_type        VARCHAR(30) NOT NULL,   -- TERM_SHEET, SHA, PITCH_DECK, FINANCIAL_MODEL, DUE_DILIGENCE, NDA, OTHER
    title           VARCHAR(200) NOT NULL,
    file_url        TEXT NOT NULL,
    file_size       BIGINT,
    mime_type       VARCHAR(50),
    is_confidential BOOLEAN DEFAULT TRUE,
    access_level    VARCHAR(20) DEFAULT 'NDA_SIGNED', -- PUBLIC, NDA_SIGNED, COMMITTED, ADMIN_ONLY
    version         INT DEFAULT 1,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_documents_deal ON deal_documents(deal_id);

-- Deal milestones (tracking deal progress)
CREATE TABLE deal_milestones (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id         UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    title           VARCHAR(200) NOT NULL,
    description     TEXT,
    milestone_type  VARCHAR(30) NOT NULL,   -- NDA_PHASE, DUE_DILIGENCE, TERM_SHEET, COMMITMENT, CLOSING, DISBURSEMENT
    status          VARCHAR(20) DEFAULT 'PENDING', -- PENDING, IN_PROGRESS, COMPLETED, SKIPPED
    due_date        DATE,
    completed_at    TIMESTAMPTZ,
    display_order   INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_milestones_deal ON deal_milestones(deal_id);

-- NDAs (Non-Disclosure Agreements)
CREATE TABLE deal_ndas (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id         UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL,
    nda_template_id UUID,
    status          VARCHAR(20) DEFAULT 'PENDING', -- PENDING, SIGNED, REJECTED, EXPIRED
    signed_at       TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    signature_hash  VARCHAR(64),             -- SHA-256 of signature data
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(deal_id, user_id)
);

CREATE INDEX idx_ndas_deal ON deal_ndas(deal_id);
CREATE INDEX idx_ndas_user ON deal_ndas(user_id);

-- Deal negotiations (messages/offers within a deal room)
CREATE TABLE deal_negotiations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id         UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL,
    message_type    VARCHAR(20) NOT NULL,    -- MESSAGE, OFFER, COUNTER_OFFER, ACCEPTANCE, REJECTION, SYSTEM
    body            TEXT NOT NULL,
    amount_paise    BIGINT,                  -- For offers/counter-offers
    equity_pct      DECIMAL(5,2),
    parent_id       UUID REFERENCES deal_negotiations(id),  -- For threaded negotiation
    status          VARCHAR(20) DEFAULT 'ACTIVE', -- ACTIVE, WITHDRAWN, ACCEPTED, REJECTED, EXPIRED
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_negotiations_deal ON deal_negotiations(deal_id, created_at);
CREATE INDEX idx_negotiations_sender ON deal_negotiations(sender_id);

-- Escrow records
CREATE TABLE deal_escrow (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id         UUID NOT NULL REFERENCES deals(id) ON DELETE CASCADE,
    participant_id  UUID NOT NULL REFERENCES deal_participants(id),
    amount_paise    BIGINT NOT NULL,
    status          VARCHAR(20) DEFAULT 'HELD', -- HELD, RELEASED, REFUNDED
    escrow_ref      VARCHAR(100),            -- External escrow reference
    held_at         TIMESTAMPTZ DEFAULT NOW(),
    released_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_escrow_deal ON deal_escrow(deal_id);

-- Triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ language 'plpgsql';

CREATE TRIGGER update_deals_updated_at BEFORE UPDATE ON deals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_participants_updated_at BEFORE UPDATE ON deal_participants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

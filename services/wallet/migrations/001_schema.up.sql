-- Wallet Service Schema
-- PostgreSQL 16
-- ALL monetary values in int64 paise (INR × 100)

-- Wallets (each user has one wallet)
CREATE TABLE wallets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE,
    balance_paise   BIGINT DEFAULT 0,
    locked_paise    BIGINT DEFAULT 0,       -- Funds in escrow
    currency        VARCHAR(3) DEFAULT 'INR',
    status          VARCHAR(20) DEFAULT 'ACTIVE',  -- ACTIVE, FROZEN, CLOSED
    kyc_verified    BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_wallets_user ON wallets(user_id);

-- Ledger entries (double-entry bookkeeping)
CREATE TABLE ledger_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets(id),
    entry_type      VARCHAR(30) NOT NULL,     -- CREDIT, DEBIT
    amount_paise    BIGINT NOT NULL CHECK (amount_paise > 0),
    balance_after   BIGINT NOT NULL,          -- Running balance after this entry
    category        VARCHAR(30) NOT NULL,     -- DEPOSIT, WITHDRAWAL, INVESTMENT, RETURN, ESCROW_LOCK, ESCROW_RELEASE, FEE, REFUND
    reference_type  VARCHAR(30),              -- DEAL, ESCROW, BANK_TRANSFER, UPI, etc.
    reference_id    UUID,                     -- ID of the related deal/escrow/transfer
    description     TEXT,
    metadata        JSONB DEFAULT '{}',       -- Additional context
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ledger_wallet ON ledger_entries(wallet_id, created_at DESC);
CREATE INDEX idx_ledger_category ON ledger_entries(category);
CREATE INDEX idx_ledger_reference ON ledger_entries(reference_type, reference_id);

-- Transactions (deposits, withdrawals, transfers)
CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets(id),
    tx_type         VARCHAR(20) NOT NULL,    -- DEPOSIT, WITHDRAWAL, TRANSFER, INVESTMENT, RETURN
    amount_paise    BIGINT NOT NULL CHECK (amount_paise > 0),
    fee_paise       BIGINT DEFAULT 0,
    net_paise       BIGINT NOT NULL,         -- amount - fee
    status          VARCHAR(20) DEFAULT 'PENDING',  -- PENDING, PROCESSING, COMPLETED, FAILED, REVERSED
    payment_method  VARCHAR(30),             -- UPI, NEFT, RTGS, IMPS, NET_BANKING
    payment_ref     VARCHAR(100),            -- External payment gateway reference
    bank_ref        VARCHAR(100),            -- Bank reference number
    counterparty_id UUID,                    -- Other wallet for transfers
    deal_id         UUID,                    -- Related deal (for investments)
    description     TEXT,
    failure_reason  TEXT,
    completed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tx_wallet ON transactions(wallet_id, created_at DESC);
CREATE INDEX idx_tx_status ON transactions(status);
CREATE INDEX idx_tx_type ON transactions(tx_type);
CREATE INDEX idx_tx_payment_ref ON transactions(payment_ref);

-- Bank accounts (linked for withdrawals)
CREATE TABLE bank_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL,
    account_holder  VARCHAR(200) NOT NULL,
    account_number  VARCHAR(30) NOT NULL,    -- Encrypted at app level
    ifsc_code       VARCHAR(11) NOT NULL,
    bank_name       VARCHAR(100) NOT NULL,
    account_type    VARCHAR(20) DEFAULT 'SAVINGS',  -- SAVINGS, CURRENT
    is_primary      BOOLEAN DEFAULT FALSE,
    is_verified     BOOLEAN DEFAULT FALSE,
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_bank_user ON bank_accounts(user_id);

-- Payment webhooks log (idempotency)
CREATE TABLE payment_webhooks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider        VARCHAR(30) NOT NULL,    -- RAZORPAY, CASHFREE, PAYU
    event_type      VARCHAR(50) NOT NULL,
    event_id        VARCHAR(100) UNIQUE NOT NULL,  -- Idempotency key
    payload         JSONB NOT NULL,
    status          VARCHAR(20) DEFAULT 'RECEIVED',  -- RECEIVED, PROCESSED, FAILED
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhook_event ON payment_webhooks(event_id);

-- Triggers
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = NOW(); RETURN NEW; END;
$$ language 'plpgsql';

CREATE TRIGGER update_wallets_updated_at BEFORE UPDATE ON wallets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	BalancePaise int64     `db:"balance_paise" json:"balance_paise"`
	LockedPaise  int64     `db:"locked_paise" json:"locked_paise"`
	Currency     string    `db:"currency" json:"currency"`
	Status       string    `db:"status" json:"status"`
	KYCVerified  bool      `db:"kyc_verified" json:"kyc_verified"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type LedgerEntry struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	WalletID      uuid.UUID      `db:"wallet_id" json:"wallet_id"`
	EntryType     string         `db:"entry_type" json:"entry_type"`
	AmountPaise   int64          `db:"amount_paise" json:"amount_paise"`
	BalanceAfter  int64          `db:"balance_after" json:"balance_after"`
	Category      string         `db:"category" json:"category"`
	ReferenceType sql.NullString `db:"reference_type" json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID     `db:"reference_id" json:"reference_id,omitempty"`
	Description   sql.NullString `db:"description" json:"description,omitempty"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

type Transaction struct {
	ID             uuid.UUID      `db:"id" json:"id"`
	WalletID       uuid.UUID      `db:"wallet_id" json:"wallet_id"`
	TxType         string         `db:"tx_type" json:"tx_type"`
	AmountPaise    int64          `db:"amount_paise" json:"amount_paise"`
	FeePaise       int64          `db:"fee_paise" json:"fee_paise"`
	NetPaise       int64          `db:"net_paise" json:"net_paise"`
	Status         string         `db:"status" json:"status"`
	PaymentMethod  sql.NullString `db:"payment_method" json:"payment_method,omitempty"`
	PaymentRef     sql.NullString `db:"payment_ref" json:"payment_ref,omitempty"`
	BankRef        sql.NullString `db:"bank_ref" json:"bank_ref,omitempty"`
	CounterpartyID *uuid.UUID     `db:"counterparty_id" json:"counterparty_id,omitempty"`
	DealID         *uuid.UUID     `db:"deal_id" json:"deal_id,omitempty"`
	Description    sql.NullString `db:"description" json:"description,omitempty"`
	FailureReason  sql.NullString `db:"failure_reason" json:"failure_reason,omitempty"`
	CompletedAt    sql.NullTime   `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

type BankAccount struct {
	ID            uuid.UUID    `db:"id" json:"id"`
	UserID        uuid.UUID    `db:"user_id" json:"user_id"`
	AccountHolder string       `db:"account_holder" json:"account_holder"`
	AccountNumber string       `db:"account_number" json:"account_number"`
	IFSCCode      string       `db:"ifsc_code" json:"ifsc_code"`
	BankName      string       `db:"bank_name" json:"bank_name"`
	AccountType   string       `db:"account_type" json:"account_type"`
	IsPrimary     bool         `db:"is_primary" json:"is_primary"`
	IsVerified    bool         `db:"is_verified" json:"is_verified"`
	VerifiedAt    sql.NullTime `db:"verified_at" json:"verified_at,omitempty"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
}

type PaymentWebhook struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	Provider    string       `db:"provider" json:"provider"`
	EventType   string       `db:"event_type" json:"event_type"`
	EventID     string       `db:"event_id" json:"event_id"`
	Status      string       `db:"status" json:"status"`
	ProcessedAt sql.NullTime `db:"processed_at" json:"processed_at,omitempty"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
}

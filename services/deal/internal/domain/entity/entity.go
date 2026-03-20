package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Deal statuses
const (
	DealStatusDraft        = "DRAFT"
	DealStatusOpen         = "OPEN"
	DealStatusInProgress   = "IN_PROGRESS"
	DealStatusDueDiligence = "DUE_DILIGENCE"
	DealStatusClosing      = "CLOSING"
	DealStatusClosed       = "CLOSED"
	DealStatusCancelled    = "CANCELLED"
)

// Participant statuses
const (
	ParticipantInterested   = "INTERESTED"
	ParticipantNDASigned    = "NDA_SIGNED"
	ParticipantDueDiligence = "DUE_DILIGENCE"
	ParticipantCommitted    = "COMMITTED"
	ParticipantInvested     = "INVESTED"
	ParticipantDeclined     = "DECLINED"
)

type Deal struct {
	ID              uuid.UUID      `db:"id" json:"id"`
	StartupID       uuid.UUID      `db:"startup_id" json:"startup_id"`
	FundingRoundID  *uuid.UUID     `db:"funding_round_id" json:"funding_round_id,omitempty"`
	Title           string         `db:"title" json:"title"`
	Description     string         `db:"description" json:"description"`
	DealType        string         `db:"deal_type" json:"deal_type"`
	Status          string         `db:"status" json:"status"`
	AmountPaise     int64          `db:"amount_paise" json:"amount_paise"`
	MinTicketPaise  sql.NullInt64  `db:"min_ticket_paise" json:"min_ticket_paise,omitempty"`
	MaxParticipants int            `db:"max_participants" json:"max_participants"`
	EquityPct       sql.NullFloat64 `db:"equity_pct" json:"equity_pct,omitempty"`
	ValuationPaise  sql.NullInt64  `db:"valuation_paise" json:"valuation_paise,omitempty"`
	Currency        string         `db:"currency" json:"currency"`
	TermsSummary    sql.NullString `db:"terms_summary" json:"terms_summary,omitempty"`
	RequiresNDA     bool           `db:"requires_nda" json:"requires_nda"`
	RequiresKYC     bool           `db:"requires_kyc" json:"requires_kyc"`
	CreatedBy       uuid.UUID      `db:"created_by" json:"created_by"`
	OpenDate        sql.NullTime   `db:"open_date" json:"open_date,omitempty"`
	CloseDate       sql.NullTime   `db:"close_date" json:"close_date,omitempty"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

type DealParticipant struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	DealID          uuid.UUID       `db:"deal_id" json:"deal_id"`
	UserID          uuid.UUID       `db:"user_id" json:"user_id"`
	Role            string          `db:"role" json:"role"`
	Status          string          `db:"status" json:"status"`
	CommitmentPaise sql.NullInt64   `db:"commitment_paise" json:"commitment_paise,omitempty"`
	InvestedPaise   int64           `db:"invested_paise" json:"invested_paise"`
	EquityPct       sql.NullFloat64 `db:"equity_pct" json:"equity_pct,omitempty"`
	NDASignedAt     sql.NullTime    `db:"nda_signed_at" json:"nda_signed_at,omitempty"`
	CommittedAt     sql.NullTime    `db:"committed_at" json:"committed_at,omitempty"`
	InvestedAt      sql.NullTime    `db:"invested_at" json:"invested_at,omitempty"`
	Notes           sql.NullString  `db:"notes" json:"notes,omitempty"`
	CreatedAt       time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at" json:"updated_at"`
}

type DealDocument struct {
	ID             uuid.UUID      `db:"id" json:"id"`
	DealID         uuid.UUID      `db:"deal_id" json:"deal_id"`
	UploadedBy     uuid.UUID      `db:"uploaded_by" json:"uploaded_by"`
	DocType        string         `db:"doc_type" json:"doc_type"`
	Title          string         `db:"title" json:"title"`
	FileURL        string         `db:"file_url" json:"file_url"`
	FileSize       sql.NullInt64  `db:"file_size" json:"file_size,omitempty"`
	MimeType       sql.NullString `db:"mime_type" json:"mime_type,omitempty"`
	IsConfidential bool           `db:"is_confidential" json:"is_confidential"`
	AccessLevel    string         `db:"access_level" json:"access_level"`
	Version        int            `db:"version" json:"version"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
}

type DealMilestone struct {
	ID            uuid.UUID    `db:"id" json:"id"`
	DealID        uuid.UUID    `db:"deal_id" json:"deal_id"`
	Title         string       `db:"title" json:"title"`
	Description   sql.NullString `db:"description" json:"description,omitempty"`
	MilestoneType string       `db:"milestone_type" json:"milestone_type"`
	Status        string       `db:"status" json:"status"`
	DueDate       sql.NullTime `db:"due_date" json:"due_date,omitempty"`
	CompletedAt   sql.NullTime `db:"completed_at" json:"completed_at,omitempty"`
	DisplayOrder  int          `db:"display_order" json:"display_order"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
}

type DealNDA struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	DealID        uuid.UUID      `db:"deal_id" json:"deal_id"`
	UserID        uuid.UUID      `db:"user_id" json:"user_id"`
	NDATemplateID *uuid.UUID     `db:"nda_template_id" json:"nda_template_id,omitempty"`
	Status        string         `db:"status" json:"status"`
	SignedAt      sql.NullTime   `db:"signed_at" json:"signed_at,omitempty"`
	ExpiresAt     sql.NullTime   `db:"expires_at" json:"expires_at,omitempty"`
	IPAddress     sql.NullString `db:"ip_address" json:"ip_address,omitempty"`
	UserAgent     sql.NullString `db:"user_agent" json:"user_agent,omitempty"`
	SignatureHash sql.NullString `db:"signature_hash" json:"signature_hash,omitempty"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

type DealNegotiation struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	DealID       uuid.UUID       `db:"deal_id" json:"deal_id"`
	SenderID     uuid.UUID       `db:"sender_id" json:"sender_id"`
	MessageType  string          `db:"message_type" json:"message_type"`
	Body         string          `db:"body" json:"body"`
	AmountPaise  sql.NullInt64   `db:"amount_paise" json:"amount_paise,omitempty"`
	EquityPct    sql.NullFloat64 `db:"equity_pct" json:"equity_pct,omitempty"`
	ParentID     *uuid.UUID      `db:"parent_id" json:"parent_id,omitempty"`
	Status       string          `db:"status" json:"status"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
}

type DealEscrow struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	DealID        uuid.UUID      `db:"deal_id" json:"deal_id"`
	ParticipantID uuid.UUID      `db:"participant_id" json:"participant_id"`
	AmountPaise   int64          `db:"amount_paise" json:"amount_paise"`
	Status        string         `db:"status" json:"status"`
	EscrowRef     sql.NullString `db:"escrow_ref" json:"escrow_ref,omitempty"`
	HeldAt        time.Time      `db:"held_at" json:"held_at"`
	ReleasedAt    sql.NullTime   `db:"released_at" json:"released_at,omitempty"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

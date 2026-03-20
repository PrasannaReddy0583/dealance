package entity

type CreateDealRequest struct {
	StartupID      string  `json:"startup_id" validate:"required,uuid"`
	FundingRoundID string  `json:"funding_round_id,omitempty" validate:"omitempty,uuid"`
	Title          string  `json:"title" validate:"required,max=300"`
	Description    string  `json:"description,omitempty" validate:"omitempty,max=5000"`
	DealType       string  `json:"deal_type" validate:"required,oneof=EQUITY CCPS CCD SAFE CONVERTIBLE_NOTE DEBT"`
	AmountPaise    int64   `json:"amount_paise" validate:"required,gt=0"`
	MinTicketPaise int64   `json:"min_ticket_paise,omitempty"`
	EquityPct      float64 `json:"equity_pct,omitempty"`
	ValuationPaise int64   `json:"valuation_paise,omitempty"`
	TermsSummary   string  `json:"terms_summary,omitempty"`
	RequiresNDA    *bool   `json:"requires_nda,omitempty"`
}

type UpdateDealRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,max=300"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=DRAFT OPEN IN_PROGRESS DUE_DILIGENCE CLOSING CLOSED CANCELLED"`
}

type DealResponse struct {
	ID              string  `json:"id"`
	StartupID       string  `json:"startup_id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	DealType        string  `json:"deal_type"`
	Status          string  `json:"status"`
	AmountPaise     int64   `json:"amount_paise"`
	MinTicketPaise  int64   `json:"min_ticket_paise,omitempty"`
	MaxParticipants int     `json:"max_participants"`
	EquityPct       float64 `json:"equity_pct,omitempty"`
	ValuationPaise  int64   `json:"valuation_paise,omitempty"`
	Currency        string  `json:"currency"`
	RequiresNDA     bool    `json:"requires_nda"`
	RequiresKYC     bool    `json:"requires_kyc"`
	CreatedAt       string  `json:"created_at"`
}

type DealListItem struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	DealType    string `json:"deal_type"`
	Status      string `json:"status"`
	AmountPaise int64  `json:"amount_paise"`
	Currency    string `json:"currency"`
	CreatedAt   string `json:"created_at"`
}

type JoinDealRequest struct {
	Role string `json:"role,omitempty" validate:"omitempty,oneof=INVESTOR LEAD_INVESTOR ADVISOR OBSERVER"`
}

type CommitRequest struct {
	AmountPaise int64 `json:"amount_paise" validate:"required,gt=0"`
}

type SignNDARequest struct {
	SignatureHash string `json:"signature_hash" validate:"required"`
}

type NegotiationMessageRequest struct {
	MessageType string  `json:"message_type" validate:"required,oneof=MESSAGE OFFER COUNTER_OFFER ACCEPTANCE REJECTION"`
	Body        string  `json:"body" validate:"required,max=2000"`
	AmountPaise int64   `json:"amount_paise,omitempty"`
	EquityPct   float64 `json:"equity_pct,omitempty"`
	ParentID    string  `json:"parent_id,omitempty" validate:"omitempty,uuid"`
}

type UploadDocumentRequest struct {
	DocType        string `json:"doc_type" validate:"required,oneof=TERM_SHEET SHA PITCH_DECK FINANCIAL_MODEL DUE_DILIGENCE NDA OTHER"`
	Title          string `json:"title" validate:"required,max=200"`
	FileURL        string `json:"file_url" validate:"required"`
	IsConfidential *bool  `json:"is_confidential,omitempty"`
	AccessLevel    string `json:"access_level,omitempty" validate:"omitempty,oneof=PUBLIC NDA_SIGNED COMMITTED ADMIN_ONLY"`
}

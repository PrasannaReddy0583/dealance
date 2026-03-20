package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Startup stages
const (
	StageIdea           = "IDEA"
	StageMVP            = "MVP"
	StageEarlyTraction  = "EARLY_TRACTION"
	StageGrowth         = "GROWTH"
	StageScale          = "SCALE"
)

// Startup represents a company on the platform.
type Startup struct {
	ID                uuid.UUID      `db:"id" json:"id"`
	FounderID         uuid.UUID      `db:"founder_id" json:"founder_id"`
	Name              string         `db:"name" json:"name"`
	Slug              string         `db:"slug" json:"slug"`
	Tagline           sql.NullString `db:"tagline" json:"tagline,omitempty"`
	Description       string         `db:"description" json:"description"`
	LogoURL           string         `db:"logo_url" json:"logo_url"`
	CoverURL          string         `db:"cover_url" json:"cover_url"`
	Website           sql.NullString `db:"website" json:"website,omitempty"`
	FoundedYear       sql.NullInt32  `db:"founded_year" json:"founded_year,omitempty"`
	Headquarters      sql.NullString `db:"headquarters" json:"headquarters,omitempty"`
	Country           sql.NullString `db:"country" json:"country,omitempty"`
	Sector            string         `db:"sector" json:"sector"`
	Stage             string         `db:"stage" json:"stage"`
	BusinessModel     sql.NullString `db:"business_model" json:"business_model,omitempty"`
	TeamSize          int            `db:"team_size" json:"team_size"`
	IncorporationType sql.NullString `db:"incorporation_type" json:"incorporation_type,omitempty"`
	CINNumber         sql.NullString `db:"cin_number" json:"cin_number,omitempty"`
	GSTIN             sql.NullString `db:"gstin" json:"gstin,omitempty"`
	Status            string         `db:"status" json:"status"`
	IsVerified        bool           `db:"is_verified" json:"is_verified"`
	IsFeatured        bool           `db:"is_featured" json:"is_featured"`
	ViewCount         int            `db:"view_count" json:"view_count"`
	FollowerCount     int            `db:"follower_count" json:"follower_count"`
	Tags              pq.StringArray `db:"tags" json:"tags"`
	CreatedAt         time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at" json:"updated_at"`
}

// FundingRound represents a fundraising round.
type FundingRound struct {
	ID              uuid.UUID        `db:"id" json:"id"`
	StartupID       uuid.UUID        `db:"startup_id" json:"startup_id"`
	RoundType       string           `db:"round_type" json:"round_type"`
	AmountPaise     int64            `db:"amount_paise" json:"amount_paise"`
	ValuationPaise  sql.NullInt64    `db:"valuation_paise" json:"valuation_paise,omitempty"`
	Currency        string           `db:"currency" json:"currency"`
	Status          string           `db:"status" json:"status"`
	TargetPaise     sql.NullInt64    `db:"target_paise" json:"target_paise,omitempty"`
	MinTicketPaise  sql.NullInt64    `db:"min_ticket_paise" json:"min_ticket_paise,omitempty"`
	EquityOffered   sql.NullFloat64  `db:"equity_offered" json:"equity_offered,omitempty"`
	InstrumentType  sql.NullString   `db:"instrument_type" json:"instrument_type,omitempty"`
	OpenDate        sql.NullTime     `db:"open_date" json:"open_date,omitempty"`
	CloseDate       sql.NullTime     `db:"close_date" json:"close_date,omitempty"`
	Description     sql.NullString   `db:"description" json:"description,omitempty"`
	TermsURL        sql.NullString   `db:"terms_url" json:"terms_url,omitempty"`
	CreatedAt       time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at" json:"updated_at"`
}

// TeamMember represents a startup team member.
type TeamMember struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	StartupID    uuid.UUID      `db:"startup_id" json:"startup_id"`
	UserID       *uuid.UUID     `db:"user_id" json:"user_id,omitempty"`
	Name         string         `db:"name" json:"name"`
	Role         string         `db:"role" json:"role"`
	Title        sql.NullString `db:"title" json:"title,omitempty"`
	Bio          string         `db:"bio" json:"bio"`
	AvatarURL    sql.NullString `db:"avatar_url" json:"avatar_url,omitempty"`
	LinkedInURL  sql.NullString `db:"linkedin_url" json:"linkedin_url,omitempty"`
	IsFounder    bool           `db:"is_founder" json:"is_founder"`
	EquityPct    sql.NullFloat64 `db:"equity_pct" json:"equity_pct,omitempty"`
	JoinedDate   sql.NullTime   `db:"joined_date" json:"joined_date,omitempty"`
	DisplayOrder int            `db:"display_order" json:"display_order"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// StartupMedia represents startup attachments (pitch deck, screenshots, etc.).
type StartupMedia struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	StartupID    uuid.UUID      `db:"startup_id" json:"startup_id"`
	MediaType    string         `db:"media_type" json:"media_type"`
	Title        sql.NullString `db:"title" json:"title,omitempty"`
	MediaURL     string         `db:"media_url" json:"media_url"`
	ThumbnailURL sql.NullString `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	FileSize     sql.NullInt64  `db:"file_size" json:"file_size,omitempty"`
	IsPublic     bool           `db:"is_public" json:"is_public"`
	DisplayOrder int            `db:"display_order" json:"display_order"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// StartupMetric represents a key metric data point.
type StartupMetric struct {
	ID         uuid.UUID        `db:"id" json:"id"`
	StartupID  uuid.UUID        `db:"startup_id" json:"startup_id"`
	MetricType string           `db:"metric_type" json:"metric_type"`
	Value      float64          `db:"value" json:"value"`
	Currency   sql.NullString   `db:"currency" json:"currency,omitempty"`
	Period     sql.NullString   `db:"period" json:"period,omitempty"`
	RecordedAt time.Time        `db:"recorded_at" json:"recorded_at"`
	IsVerified bool             `db:"is_verified" json:"is_verified"`
	CreatedAt  time.Time        `db:"created_at" json:"created_at"`
}

// StartupFollow represents a user following a startup.
type StartupFollow struct {
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	StartupID uuid.UUID `db:"startup_id" json:"startup_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

package entity

// --- Startup DTOs ---

type CreateStartupRequest struct {
	Name              string   `json:"name" validate:"required,min=2,max=200"`
	Tagline           string   `json:"tagline,omitempty" validate:"omitempty,max=300"`
	Description       string   `json:"description,omitempty" validate:"omitempty,max=5000"`
	Sector            string   `json:"sector" validate:"required,max=50"`
	Stage             string   `json:"stage,omitempty" validate:"omitempty,oneof=IDEA MVP EARLY_TRACTION GROWTH SCALE"`
	BusinessModel     string   `json:"business_model,omitempty"`
	Website           string   `json:"website,omitempty" validate:"omitempty,url"`
	FoundedYear       int      `json:"founded_year,omitempty"`
	Headquarters      string   `json:"headquarters,omitempty"`
	Country           string   `json:"country,omitempty"`
	IncorporationType string   `json:"incorporation_type,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

type UpdateStartupRequest struct {
	Name              *string  `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	Tagline           *string  `json:"tagline,omitempty" validate:"omitempty,max=300"`
	Description       *string  `json:"description,omitempty" validate:"omitempty,max=5000"`
	LogoURL           *string  `json:"logo_url,omitempty"`
	CoverURL          *string  `json:"cover_url,omitempty"`
	Sector            *string  `json:"sector,omitempty"`
	Stage             *string  `json:"stage,omitempty" validate:"omitempty,oneof=IDEA MVP EARLY_TRACTION GROWTH SCALE"`
	BusinessModel     *string  `json:"business_model,omitempty"`
	Website           *string  `json:"website,omitempty"`
	TeamSize          *int     `json:"team_size,omitempty"`
	IncorporationType *string  `json:"incorporation_type,omitempty"`
	CINNumber         *string  `json:"cin_number,omitempty"`
	GSTIN             *string  `json:"gstin,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

type StartupResponse struct {
	ID            string `json:"id"`
	FounderID     string `json:"founder_id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Tagline       string `json:"tagline,omitempty"`
	Description   string `json:"description"`
	LogoURL       string `json:"logo_url"`
	CoverURL      string `json:"cover_url"`
	Website       string `json:"website,omitempty"`
	Sector        string `json:"sector"`
	Stage         string `json:"stage"`
	BusinessModel string `json:"business_model,omitempty"`
	TeamSize      int    `json:"team_size"`
	Status        string `json:"status"`
	IsVerified    bool   `json:"is_verified"`
	ViewCount     int    `json:"view_count"`
	FollowerCount int    `json:"follower_count"`
	IsFollowing   bool   `json:"is_following,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type StartupListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Tagline    string `json:"tagline,omitempty"`
	LogoURL    string `json:"logo_url"`
	Sector     string `json:"sector"`
	Stage      string `json:"stage"`
	IsVerified bool   `json:"is_verified"`
}

// --- Funding Round DTOs ---

type CreateFundingRoundRequest struct {
	RoundType      string  `json:"round_type" validate:"required,oneof=PRE_SEED SEED SERIES_A SERIES_B BRIDGE DEBT"`
	AmountPaise    int64   `json:"amount_paise" validate:"required,gt=0"`
	ValuationPaise int64   `json:"valuation_paise,omitempty"`
	TargetPaise    int64   `json:"target_paise,omitempty"`
	MinTicketPaise int64   `json:"min_ticket_paise,omitempty"`
	EquityOffered  float64 `json:"equity_offered,omitempty"`
	InstrumentType string  `json:"instrument_type,omitempty"`
	Description    string  `json:"description,omitempty"`
}

// --- Team Member DTOs ---

type AddTeamMemberRequest struct {
	Name        string `json:"name" validate:"required,max=100"`
	Role        string `json:"role" validate:"required,max=100"`
	Title       string `json:"title,omitempty"`
	Bio         string `json:"bio,omitempty"`
	LinkedInURL string `json:"linkedin_url,omitempty"`
	IsFounder   bool   `json:"is_founder,omitempty"`
}

// --- Search DTOs ---

type SearchStartupsRequest struct {
	Query   string `json:"query,omitempty"`
	Sector  string `json:"sector,omitempty"`
	Stage   string `json:"stage,omitempty"`
	Country string `json:"country,omitempty"`
	Limit   int    `json:"limit,omitempty"`
}

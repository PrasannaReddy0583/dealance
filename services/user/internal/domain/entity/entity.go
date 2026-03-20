package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Profile is the social identity of a user.
type Profile struct {
	ID              uuid.UUID      `db:"id" json:"id"`
	Username        string         `db:"username" json:"username"`
	DisplayName     string         `db:"display_name" json:"display_name"`
	Bio             string         `db:"bio" json:"bio"`
	AvatarURL       string         `db:"avatar_url" json:"avatar_url"`
	CoverURL        string         `db:"cover_url" json:"cover_url"`
	Location        string         `db:"location" json:"location"`
	Website         string         `db:"website" json:"website"`
	LinkedInURL     string         `db:"linkedin_url" json:"linkedin_url"`
	TwitterURL      string         `db:"twitter_url" json:"twitter_url"`
	DateOfBirth     sql.NullTime   `db:"date_of_birth" json:"date_of_birth,omitempty"`
	Gender          sql.NullString `db:"gender" json:"gender,omitempty"`
	Profession      string         `db:"profession" json:"profession"`
	Company         string         `db:"company" json:"company"`
	ExperienceYears int            `db:"experience_years" json:"experience_years"`
	IsPublic        bool           `db:"is_public" json:"is_public"`
	IsVerified      bool           `db:"is_verified" json:"is_verified"`
	FollowerCount   int            `db:"follower_count" json:"follower_count"`
	FollowingCount  int            `db:"following_count" json:"following_count"`
	PostCount       int            `db:"post_count" json:"post_count"`
	CreatedAt       time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at" json:"updated_at"`
}

// Follow represents a follower→following edge.
type Follow struct {
	FollowerID  uuid.UUID `db:"follower_id" json:"follower_id"`
	FollowingID uuid.UUID `db:"following_id" json:"following_id"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// BlockedUser represents a blocked relationship.
type BlockedUser struct {
	BlockerID uuid.UUID      `db:"blocker_id" json:"blocker_id"`
	BlockedID uuid.UUID      `db:"blocked_id" json:"blocked_id"`
	Reason    sql.NullString `db:"reason" json:"reason,omitempty"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}

// UserSettings holds all user preferences.
type UserSettings struct {
	UserID                    uuid.UUID `db:"user_id" json:"user_id"`
	NotificationPush          bool      `db:"notification_push" json:"notification_push"`
	NotificationEmail         bool      `db:"notification_email" json:"notification_email"`
	NotificationSMS           bool      `db:"notification_sms" json:"notification_sms"`
	NotificationDealUpdates   bool      `db:"notification_deal_updates" json:"notification_deal_updates"`
	NotificationNewFollowers  bool      `db:"notification_new_followers" json:"notification_new_followers"`
	NotificationMessages      bool      `db:"notification_messages" json:"notification_messages"`
	PrivacyShowEmail          bool      `db:"privacy_show_email" json:"privacy_show_email"`
	PrivacyShowPhone          bool      `db:"privacy_show_phone" json:"privacy_show_phone"`
	PrivacyShowLocation       bool      `db:"privacy_show_location" json:"privacy_show_location"`
	PrivacyAllowMessages      string    `db:"privacy_allow_messages" json:"privacy_allow_messages"`
	PrivacyShowInvestments    bool      `db:"privacy_show_investments" json:"privacy_show_investments"`
	FeedContentLanguage       string    `db:"feed_content_language" json:"feed_content_language"`
	FeedSortPreference        string    `db:"feed_sort_preference" json:"feed_sort_preference"`
	Theme                     string    `db:"theme" json:"theme"`
	UpdatedAt                 time.Time `db:"updated_at" json:"updated_at"`
}

// ProfileMedia represents portfolio items, certifications, etc.
type ProfileMedia struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	UserID       uuid.UUID      `db:"user_id" json:"user_id"`
	MediaType    string         `db:"media_type" json:"media_type"`
	Title        string         `db:"title" json:"title"`
	Description  string         `db:"description" json:"description"`
	MediaURL     string         `db:"media_url" json:"media_url"`
	ThumbnailURL sql.NullString `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	DisplayOrder int            `db:"display_order" json:"display_order"`
	IsVisible    bool           `db:"is_visible" json:"is_visible"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

// EntrepreneurProfile holds entrepreneur-specific data.
type EntrepreneurProfile struct {
	UserID           uuid.UUID      `db:"user_id" json:"user_id"`
	StartupCount     int            `db:"startup_count" json:"startup_count"`
	TotalRaisedPaise int64          `db:"total_raised_paise" json:"total_raised_paise"`
	Sectors          pq.StringArray `db:"sectors" json:"sectors"`
	Skills           pq.StringArray `db:"skills" json:"skills"`
	Education        []byte         `db:"education" json:"education"`       // JSONB
	WorkHistory      []byte         `db:"work_history" json:"work_history"` // JSONB
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
}

// InvestorProfile holds investor-specific data.
type InvestorProfile struct {
	UserID                 uuid.UUID      `db:"user_id" json:"user_id"`
	InvestorType           sql.NullString `db:"investor_type" json:"investor_type,omitempty"`
	InvestmentRangeMinPaise int64         `db:"investment_range_min_paise" json:"investment_range_min_paise"`
	InvestmentRangeMaxPaise int64         `db:"investment_range_max_paise" json:"investment_range_max_paise"`
	PreferredSectors       pq.StringArray `db:"preferred_sectors" json:"preferred_sectors"`
	PreferredStages        pq.StringArray `db:"preferred_stages" json:"preferred_stages"`
	PortfolioCount         int            `db:"portfolio_count" json:"portfolio_count"`
	TotalInvestedPaise     int64          `db:"total_invested_paise" json:"total_invested_paise"`
	InvestmentThesis       string         `db:"investment_thesis" json:"investment_thesis"`
	CreatedAt              time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time      `db:"updated_at" json:"updated_at"`
}

package entity

// --- Profile DTOs ---

type CreateProfileRequest struct {
	UserID      string `json:"user_id" validate:"required,uuid"`
	Username    string `json:"username" validate:"required,min=3,max=30,alphanum"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
}

type UpdateProfileRequest struct {
	DisplayName     *string `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`
	Bio             *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	CoverURL        *string `json:"cover_url,omitempty"`
	Location        *string `json:"location,omitempty" validate:"omitempty,max=100"`
	Website         *string `json:"website,omitempty" validate:"omitempty,url,max=255"`
	LinkedInURL     *string `json:"linkedin_url,omitempty" validate:"omitempty,url,max=255"`
	TwitterURL      *string `json:"twitter_url,omitempty" validate:"omitempty,url,max=255"`
	Profession      *string `json:"profession,omitempty" validate:"omitempty,max=100"`
	Company         *string `json:"company,omitempty" validate:"omitempty,max=100"`
	ExperienceYears *int    `json:"experience_years,omitempty" validate:"omitempty,min=0,max=60"`
	IsPublic        *bool   `json:"is_public,omitempty"`
}

type ProfileResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	Bio             string `json:"bio"`
	AvatarURL       string `json:"avatar_url"`
	CoverURL        string `json:"cover_url"`
	Location        string `json:"location"`
	Website         string `json:"website"`
	LinkedInURL     string `json:"linkedin_url"`
	TwitterURL      string `json:"twitter_url"`
	Profession      string `json:"profession"`
	Company         string `json:"company"`
	ExperienceYears int    `json:"experience_years"`
	IsPublic        bool   `json:"is_public"`
	IsVerified      bool   `json:"is_verified"`
	FollowerCount   int    `json:"follower_count"`
	FollowingCount  int    `json:"following_count"`
	PostCount       int    `json:"post_count"`
	IsFollowing     bool   `json:"is_following,omitempty"`
	IsFollowedBy    bool   `json:"is_followed_by,omitempty"`
	CreatedAt       string `json:"created_at"`
}

type ProfileListItem struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	Profession  string `json:"profession"`
	IsVerified  bool   `json:"is_verified"`
	IsFollowing bool   `json:"is_following,omitempty"`
}

// --- Follow DTOs ---

type FollowRequest struct {
	TargetUserID string `json:"target_user_id" validate:"required,uuid"`
}

type FollowCountsResponse struct {
	FollowerCount  int `json:"follower_count"`
	FollowingCount int `json:"following_count"`
}

// --- Settings DTOs ---

type UpdateSettingsRequest struct {
	NotificationPush          *bool   `json:"notification_push,omitempty"`
	NotificationEmail         *bool   `json:"notification_email,omitempty"`
	NotificationSMS           *bool   `json:"notification_sms,omitempty"`
	NotificationDealUpdates   *bool   `json:"notification_deal_updates,omitempty"`
	NotificationNewFollowers  *bool   `json:"notification_new_followers,omitempty"`
	NotificationMessages      *bool   `json:"notification_messages,omitempty"`
	PrivacyShowEmail          *bool   `json:"privacy_show_email,omitempty"`
	PrivacyShowPhone          *bool   `json:"privacy_show_phone,omitempty"`
	PrivacyShowLocation       *bool   `json:"privacy_show_location,omitempty"`
	PrivacyAllowMessages      *string `json:"privacy_allow_messages,omitempty" validate:"omitempty,oneof=EVERYONE FOLLOWERS NOBODY"`
	PrivacyShowInvestments    *bool   `json:"privacy_show_investments,omitempty"`
	FeedContentLanguage       *string `json:"feed_content_language,omitempty" validate:"omitempty,len=2"`
	FeedSortPreference        *string `json:"feed_sort_preference,omitempty" validate:"omitempty,oneof=ALGORITHMIC CHRONOLOGICAL"`
	Theme                     *string `json:"theme,omitempty" validate:"omitempty,oneof=LIGHT DARK SYSTEM"`
}

// --- ProfileMedia DTOs ---

type CreateProfileMediaRequest struct {
	MediaType    string `json:"media_type" validate:"required,oneof=PORTFOLIO CERTIFICATION AWARD PRESS"`
	Title        string `json:"title" validate:"required,min=1,max=200"`
	Description  string `json:"description,omitempty" validate:"omitempty,max=1000"`
	MediaURL     string `json:"media_url" validate:"required,url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty" validate:"omitempty,url"`
	DisplayOrder int    `json:"display_order,omitempty"`
}

type UpdateProfileMediaRequest struct {
	Title        *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description  *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	MediaURL     *string `json:"media_url,omitempty" validate:"omitempty,url"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty" validate:"omitempty,url"`
	DisplayOrder *int    `json:"display_order,omitempty"`
	IsVisible    *bool   `json:"is_visible,omitempty"`
}

// --- Entrepreneur/Investor Profile DTOs ---

type UpdateEntrepreneurProfileRequest struct {
	Sectors     []string `json:"sectors,omitempty"`
	Skills      []string `json:"skills,omitempty"`
	Education   []byte   `json:"education,omitempty"`
	WorkHistory []byte   `json:"work_history,omitempty"`
}

type UpdateInvestorProfileRequest struct {
	InvestorType            *string  `json:"investor_type,omitempty" validate:"omitempty,oneof=ANGEL VC FAMILY_OFFICE HNI SYNDICATE"`
	InvestmentRangeMinPaise *int64   `json:"investment_range_min_paise,omitempty"`
	InvestmentRangeMaxPaise *int64   `json:"investment_range_max_paise,omitempty"`
	PreferredSectors        []string `json:"preferred_sectors,omitempty"`
	PreferredStages         []string `json:"preferred_stages,omitempty"`
	InvestmentThesis        *string  `json:"investment_thesis,omitempty" validate:"omitempty,max=2000"`
}

// --- Block DTOs ---

type BlockUserRequest struct {
	TargetUserID string `json:"target_user_id" validate:"required,uuid"`
	Reason       string `json:"reason,omitempty" validate:"omitempty,max=100"`
}

// --- Search DTOs ---

type SearchUsersRequest struct {
	Query      string `json:"query" validate:"required,min=2,max=100"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

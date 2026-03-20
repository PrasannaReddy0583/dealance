package entity

// --- Post DTOs ---

type CreatePostRequest struct {
	PostType      string   `json:"post_type" validate:"required,oneof=TEXT SHORT VIDEO ARTICLE"`
	Title         string   `json:"title,omitempty" validate:"omitempty,max=300"`
	Body          string   `json:"body" validate:"required,max=5000"`
	Visibility    string   `json:"visibility,omitempty" validate:"omitempty,oneof=PUBLIC FOLLOWERS PRIVATE"`
	AllowComments *bool    `json:"allow_comments,omitempty"`
	Hashtags      []string `json:"hashtags,omitempty"`
	MentionIDs    []string `json:"mention_ids,omitempty"`
	Media         []CreatePostMediaRequest `json:"media,omitempty"`
}

type UpdatePostRequest struct {
	Title         *string  `json:"title,omitempty" validate:"omitempty,max=300"`
	Body          *string  `json:"body,omitempty" validate:"omitempty,max=5000"`
	Visibility    *string  `json:"visibility,omitempty" validate:"omitempty,oneof=PUBLIC FOLLOWERS PRIVATE"`
	AllowComments *bool    `json:"allow_comments,omitempty"`
	Hashtags      []string `json:"hashtags,omitempty"`
	IsPinned      *bool    `json:"is_pinned,omitempty"`
}

type CreatePostMediaRequest struct {
	MediaType    string `json:"media_type" validate:"required,oneof=IMAGE VIDEO THUMBNAIL"`
	MediaURL     string `json:"media_url" validate:"required,url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	DurationMs   int    `json:"duration_ms,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	DisplayOrder int    `json:"display_order,omitempty"`
}

type PostResponse struct {
	ID            string              `json:"id"`
	AuthorID      string              `json:"author_id"`
	PostType      string              `json:"post_type"`
	Title         string              `json:"title,omitempty"`
	Body          string              `json:"body"`
	Visibility    string              `json:"visibility"`
	IsPublished   bool                `json:"is_published"`
	IsPinned      bool                `json:"is_pinned"`
	AllowComments bool                `json:"allow_comments"`
	Hashtags      []string            `json:"hashtags"`
	ViewCount     int                 `json:"view_count"`
	LikeCount     int                 `json:"like_count"`
	CommentCount  int                 `json:"comment_count"`
	ShareCount    int                 `json:"share_count"`
	SaveCount     int                 `json:"save_count"`
	Media         []PostMediaResponse `json:"media,omitempty"`
	HasLiked      bool                `json:"has_liked"`
	HasSaved      bool                `json:"has_saved"`
	CreatedAt     string              `json:"created_at"`
}

type PostMediaResponse struct {
	ID           string `json:"id"`
	MediaType    string `json:"media_type"`
	MediaURL     string `json:"media_url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	DurationMs   int    `json:"duration_ms,omitempty"`
	DisplayOrder int    `json:"display_order"`
}

type PostListItem struct {
	ID           string `json:"id"`
	AuthorID     string `json:"author_id"`
	PostType     string `json:"post_type"`
	Title        string `json:"title,omitempty"`
	Body         string `json:"body"`
	LikeCount    int    `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	CreatedAt    string `json:"created_at"`
}

// --- Comment DTOs ---

type CreateCommentRequest struct {
	PostID   string `json:"post_id" validate:"required,uuid"`
	ParentID string `json:"parent_id,omitempty" validate:"omitempty,uuid"`
	Body     string `json:"body" validate:"required,min=1,max=2000"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" validate:"required,min=1,max=2000"`
}

type CommentResponse struct {
	ID         string `json:"id"`
	PostID     string `json:"post_id"`
	AuthorID   string `json:"author_id"`
	ParentID   string `json:"parent_id,omitempty"`
	Body       string `json:"body"`
	LikeCount  int    `json:"like_count"`
	ReplyCount int    `json:"reply_count"`
	IsEdited   bool   `json:"is_edited"`
	CreatedAt  string `json:"created_at"`
}

// --- Reaction DTOs ---

type ReactRequest struct {
	TargetID     string `json:"target_id" validate:"required,uuid"`
	TargetType   string `json:"target_type" validate:"required,oneof=POST COMMENT"`
	ReactionType string `json:"reaction_type" validate:"required,oneof=LIKE CELEBRATE INSIGHTFUL LOVE CURIOUS"`
}

type UnreactRequest struct {
	TargetID   string `json:"target_id" validate:"required,uuid"`
	TargetType string `json:"target_type" validate:"required,oneof=POST COMMENT"`
}

// --- Save DTOs ---

type SavePostRequest struct {
	PostID     string `json:"post_id" validate:"required,uuid"`
	Collection string `json:"collection,omitempty" validate:"omitempty,max=50"`
}

// --- Report DTOs ---

type ReportRequest struct {
	TargetID    string `json:"target_id" validate:"required,uuid"`
	TargetType  string `json:"target_type" validate:"required,oneof=POST COMMENT"`
	Reason      string `json:"reason" validate:"required,oneof=SPAM HARASSMENT MISINFORMATION NSFW OTHER"`
	Description string `json:"description,omitempty" validate:"omitempty,max=500"`
}

// --- Feed DTOs ---

type FeedRequest struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

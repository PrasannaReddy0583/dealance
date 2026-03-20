package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Post types
const (
	PostTypeText    = "TEXT"
	PostTypeShort   = "SHORT"
	PostTypeVideo   = "VIDEO"
	PostTypeArticle = "ARTICLE"
)

// Visibility
const (
	VisibilityPublic    = "PUBLIC"
	VisibilityFollowers = "FOLLOWERS"
	VisibilityPrivate   = "PRIVATE"
)

// Reaction types
const (
	ReactionLike       = "LIKE"
	ReactionCelebrate  = "CELEBRATE"
	ReactionInsightful = "INSIGHTFUL"
	ReactionLove       = "LOVE"
	ReactionCurious    = "CURIOUS"
)

// Post is the core content entity.
type Post struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	AuthorID      uuid.UUID      `db:"author_id" json:"author_id"`
	PostType      string         `db:"post_type" json:"post_type"`
	Title         sql.NullString `db:"title" json:"title,omitempty"`
	Body          string         `db:"body" json:"body"`
	Visibility    string         `db:"visibility" json:"visibility"`
	IsPublished   bool           `db:"is_published" json:"is_published"`
	IsPinned      bool           `db:"is_pinned" json:"is_pinned"`
	AllowComments bool           `db:"allow_comments" json:"allow_comments"`
	Hashtags      pq.StringArray `db:"hashtags" json:"hashtags"`
	MentionIDs    pq.StringArray `db:"mention_ids" json:"mention_ids"`
	ViewCount     int            `db:"view_count" json:"view_count"`
	LikeCount     int            `db:"like_count" json:"like_count"`
	CommentCount  int            `db:"comment_count" json:"comment_count"`
	ShareCount    int            `db:"share_count" json:"share_count"`
	SaveCount     int            `db:"save_count" json:"save_count"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

// PostMedia is an attachment to a post.
type PostMedia struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	PostID       uuid.UUID      `db:"post_id" json:"post_id"`
	MediaType    string         `db:"media_type" json:"media_type"`
	MediaURL     string         `db:"media_url" json:"media_url"`
	ThumbnailURL sql.NullString `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	Width        sql.NullInt32  `db:"width" json:"width,omitempty"`
	Height       sql.NullInt32  `db:"height" json:"height,omitempty"`
	DurationMs   sql.NullInt32  `db:"duration_ms" json:"duration_ms,omitempty"`
	FileSize     sql.NullInt64  `db:"file_size" json:"file_size,omitempty"`
	MimeType     sql.NullString `db:"mime_type" json:"mime_type,omitempty"`
	DisplayOrder int            `db:"display_order" json:"display_order"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// Comment supports threaded comments.
type Comment struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	PostID     uuid.UUID  `db:"post_id" json:"post_id"`
	AuthorID   uuid.UUID  `db:"author_id" json:"author_id"`
	ParentID   *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	Body       string     `db:"body" json:"body"`
	LikeCount  int        `db:"like_count" json:"like_count"`
	ReplyCount int        `db:"reply_count" json:"reply_count"`
	IsEdited   bool       `db:"is_edited" json:"is_edited"`
	IsDeleted  bool       `db:"is_deleted" json:"is_deleted"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}

// Reaction is a like/celebrate/etc on a post or comment.
type Reaction struct {
	ID           uuid.UUID `db:"id" json:"id"`
	UserID       uuid.UUID `db:"user_id" json:"user_id"`
	TargetID     uuid.UUID `db:"target_id" json:"target_id"`
	TargetType   string    `db:"target_type" json:"target_type"` // POST, COMMENT
	ReactionType string    `db:"reaction_type" json:"reaction_type"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// SavedPost is a user bookmark.
type SavedPost struct {
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	PostID     uuid.UUID `db:"post_id" json:"post_id"`
	Collection string    `db:"collection" json:"collection"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// Hashtag for trending/discovery.
type Hashtag struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	PostCount int       `db:"post_count" json:"post_count"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Report for content moderation.
type Report struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	ReporterID  uuid.UUID      `db:"reporter_id" json:"reporter_id"`
	TargetID    uuid.UUID      `db:"target_id" json:"target_id"`
	TargetType  string         `db:"target_type" json:"target_type"`
	Reason      string         `db:"reason" json:"reason"`
	Description sql.NullString `db:"description" json:"description,omitempty"`
	Status      string         `db:"status" json:"status"`
	ReviewedBy  *uuid.UUID     `db:"reviewed_by" json:"reviewed_by,omitempty"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}

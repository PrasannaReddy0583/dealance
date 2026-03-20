package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dealance/services/content/internal/domain/entity"
)

// PostRepository handles post CRUD.
type PostRepository interface {
	Create(ctx context.Context, post *entity.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByAuthor(ctx context.Context, authorID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Post, error)
	GetFeed(ctx context.Context, limit int, cursorTime *time.Time) ([]entity.Post, error)
	GetByHashtag(ctx context.Context, hashtag string, limit int, cursorTime *time.Time) ([]entity.Post, error)
	IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error
}

// PostMediaRepository handles post media attachments.
type PostMediaRepository interface {
	CreateBatch(ctx context.Context, media []entity.PostMedia) error
	GetByPostID(ctx context.Context, postID uuid.UUID) ([]entity.PostMedia, error)
	DeleteByPostID(ctx context.Context, postID uuid.UUID) error
}

// CommentRepository handles comments.
type CommentRepository interface {
	Create(ctx context.Context, comment *entity.Comment) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error)
	GetByPostID(ctx context.Context, postID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Comment, error)
	GetReplies(ctx context.Context, parentID uuid.UUID, limit int) ([]entity.Comment, error)
	Update(ctx context.Context, id uuid.UUID, body string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error
}

// ReactionRepository handles reactions (likes, etc.).
type ReactionRepository interface {
	Create(ctx context.Context, reaction *entity.Reaction) error
	Delete(ctx context.Context, userID, targetID uuid.UUID, targetType string) error
	GetByUserAndTarget(ctx context.Context, userID, targetID uuid.UUID, targetType string) (*entity.Reaction, error)
	HasReacted(ctx context.Context, userID, targetID uuid.UUID, targetType string) (bool, error)
	GetReactionCounts(ctx context.Context, targetID uuid.UUID, targetType string) (map[string]int, error)
}

// SavedPostRepository handles bookmarks.
type SavedPostRepository interface {
	Save(ctx context.Context, userID, postID uuid.UUID, collection string) error
	Unsave(ctx context.Context, userID, postID uuid.UUID) error
	IsSaved(ctx context.Context, userID, postID uuid.UUID) (bool, error)
	GetSavedPosts(ctx context.Context, userID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.SavedPost, error)
}

// HashtagRepository handles hashtag tracking.
type HashtagRepository interface {
	UpsertBatch(ctx context.Context, names []string) error
	IncrementCount(ctx context.Context, name string, delta int) error
	GetTrending(ctx context.Context, limit int) ([]entity.Hashtag, error)
}

// ReportRepository handles content reports.
type ReportRepository interface {
	Create(ctx context.Context, report *entity.Report) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Report, error)
	GetPending(ctx context.Context, limit int) ([]entity.Report, error)
}

// CacheRepository handles Redis caching for content data.
type CacheRepository interface {
	CacheReactionCount(ctx context.Context, targetID, targetType string, count int) error
	GetReactionCount(ctx context.Context, targetID, targetType string) (int, error)
	IncrReactionCount(ctx context.Context, targetID, targetType string, delta int) error
	CacheTrendingHashtags(ctx context.Context, hashtags []entity.Hashtag) error
	GetTrendingHashtags(ctx context.Context) ([]entity.Hashtag, error)
}

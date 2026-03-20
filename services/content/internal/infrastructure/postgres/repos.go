package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/dealance/services/content/internal/domain/entity"
)

// --- PostRepo ---

type PostRepo struct{ db *sqlx.DB }

func NewPostRepo(db *sqlx.DB) *PostRepo { return &PostRepo{db: db} }

func (r *PostRepo) Create(ctx context.Context, p *entity.Post) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO posts (id, author_id, post_type, title, body, visibility, is_published, allow_comments, hashtags, mention_ids)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		p.ID, p.AuthorID, p.PostType, p.Title, p.Body, p.Visibility, p.IsPublished, p.AllowComments, p.Hashtags, p.MentionIDs,
	)
	return err
}

func (r *PostRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error) {
	var p entity.Post
	err := r.db.GetContext(ctx, &p, `SELECT * FROM posts WHERE id = $1 AND is_published = true`, id)
	return &p, err
}

func (r *PostRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	args = append(args, id)
	q := fmt.Sprintf(`UPDATE posts SET %s WHERE id = $%d`, strings.Join(setClauses, ", "), i)
	_, err := r.db.ExecContext(ctx, q, args...)
	return err
}

func (r *PostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	return err
}

func (r *PostRepo) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Post, error) {
	var posts []entity.Post
	if cursorTime != nil {
		err := r.db.SelectContext(ctx, &posts,
			`SELECT * FROM posts WHERE author_id = $1 AND is_published = true AND created_at < $3
			ORDER BY created_at DESC LIMIT $2`, authorID, limit, cursorTime)
		return posts, err
	}
	err := r.db.SelectContext(ctx, &posts,
		`SELECT * FROM posts WHERE author_id = $1 AND is_published = true
		ORDER BY created_at DESC LIMIT $2`, authorID, limit)
	return posts, err
}

func (r *PostRepo) GetFeed(ctx context.Context, limit int, cursorTime *time.Time) ([]entity.Post, error) {
	var posts []entity.Post
	if cursorTime != nil {
		err := r.db.SelectContext(ctx, &posts,
			`SELECT * FROM posts WHERE is_published = true AND visibility = 'PUBLIC' AND created_at < $2
			ORDER BY created_at DESC LIMIT $1`, limit, cursorTime)
		return posts, err
	}
	err := r.db.SelectContext(ctx, &posts,
		`SELECT * FROM posts WHERE is_published = true AND visibility = 'PUBLIC'
		ORDER BY created_at DESC LIMIT $1`, limit)
	return posts, err
}

func (r *PostRepo) GetByHashtag(ctx context.Context, hashtag string, limit int, cursorTime *time.Time) ([]entity.Post, error) {
	var posts []entity.Post
	if cursorTime != nil {
		err := r.db.SelectContext(ctx, &posts,
			`SELECT * FROM posts WHERE $1 = ANY(hashtags) AND is_published = true AND created_at < $3
			ORDER BY created_at DESC LIMIT $2`, hashtag, limit, cursorTime)
		return posts, err
	}
	err := r.db.SelectContext(ctx, &posts,
		`SELECT * FROM posts WHERE $1 = ANY(hashtags) AND is_published = true
		ORDER BY created_at DESC LIMIT $2`, hashtag, limit)
	return posts, err
}

func (r *PostRepo) IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error {
	q := fmt.Sprintf(`UPDATE posts SET %s = %s + $1 WHERE id = $2`, column, column)
	_, err := r.db.ExecContext(ctx, q, delta, id)
	return err
}

// --- PostMediaRepo ---

type PostMediaRepo struct{ db *sqlx.DB }

func NewPostMediaRepo(db *sqlx.DB) *PostMediaRepo { return &PostMediaRepo{db: db} }

func (r *PostMediaRepo) CreateBatch(ctx context.Context, media []entity.PostMedia) error {
	if len(media) == 0 {
		return nil
	}
	q := `INSERT INTO post_media (id, post_id, media_type, media_url, thumbnail_url, width, height, duration_ms, file_size, mime_type, display_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	for _, m := range media {
		_, err := r.db.ExecContext(ctx, q,
			m.ID, m.PostID, m.MediaType, m.MediaURL, m.ThumbnailURL,
			m.Width, m.Height, m.DurationMs, m.FileSize, m.MimeType, m.DisplayOrder)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostMediaRepo) GetByPostID(ctx context.Context, postID uuid.UUID) ([]entity.PostMedia, error) {
	var media []entity.PostMedia
	err := r.db.SelectContext(ctx, &media,
		`SELECT * FROM post_media WHERE post_id = $1 ORDER BY display_order`, postID)
	return media, err
}

func (r *PostMediaRepo) DeleteByPostID(ctx context.Context, postID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM post_media WHERE post_id = $1`, postID)
	return err
}

// --- CommentRepo ---

type CommentRepo struct{ db *sqlx.DB }

func NewCommentRepo(db *sqlx.DB) *CommentRepo { return &CommentRepo{db: db} }

func (r *CommentRepo) Create(ctx context.Context, c *entity.Comment) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO comments (id, post_id, author_id, parent_id, body) VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.PostID, c.AuthorID, c.ParentID, c.Body)
	return err
}

func (r *CommentRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error) {
	var c entity.Comment
	err := r.db.GetContext(ctx, &c, `SELECT * FROM comments WHERE id = $1 AND is_deleted = false`, id)
	return &c, err
}

func (r *CommentRepo) GetByPostID(ctx context.Context, postID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Comment, error) {
	var comments []entity.Comment
	if cursorTime != nil {
		err := r.db.SelectContext(ctx, &comments,
			`SELECT * FROM comments WHERE post_id = $1 AND parent_id IS NULL AND is_deleted = false AND created_at < $3
			ORDER BY created_at DESC LIMIT $2`, postID, limit, cursorTime)
		return comments, err
	}
	err := r.db.SelectContext(ctx, &comments,
		`SELECT * FROM comments WHERE post_id = $1 AND parent_id IS NULL AND is_deleted = false
		ORDER BY created_at DESC LIMIT $2`, postID, limit)
	return comments, err
}

func (r *CommentRepo) GetReplies(ctx context.Context, parentID uuid.UUID, limit int) ([]entity.Comment, error) {
	var replies []entity.Comment
	err := r.db.SelectContext(ctx, &replies,
		`SELECT * FROM comments WHERE parent_id = $1 AND is_deleted = false ORDER BY created_at LIMIT $2`, parentID, limit)
	return replies, err
}

func (r *CommentRepo) Update(ctx context.Context, id uuid.UUID, body string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE comments SET body = $1, is_edited = true WHERE id = $2`, body, id)
	return err
}

func (r *CommentRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE comments SET is_deleted = true, body = '[deleted]' WHERE id = $1`, id)
	return err
}

func (r *CommentRepo) IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error {
	q := fmt.Sprintf(`UPDATE comments SET %s = %s + $1 WHERE id = $2`, column, column)
	_, err := r.db.ExecContext(ctx, q, delta, id)
	return err
}

// --- ReactionRepo ---

type ReactionRepo struct{ db *sqlx.DB }

func NewReactionRepo(db *sqlx.DB) *ReactionRepo { return &ReactionRepo{db: db} }

func (r *ReactionRepo) Create(ctx context.Context, rx *entity.Reaction) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO reactions (id, user_id, target_id, target_type, reaction_type)
		VALUES ($1,$2,$3,$4,$5) ON CONFLICT (user_id, target_id, target_type) DO UPDATE SET reaction_type = $5`,
		rx.ID, rx.UserID, rx.TargetID, rx.TargetType, rx.ReactionType)
	return err
}

func (r *ReactionRepo) Delete(ctx context.Context, userID, targetID uuid.UUID, targetType string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM reactions WHERE user_id = $1 AND target_id = $2 AND target_type = $3`,
		userID, targetID, targetType)
	return err
}

func (r *ReactionRepo) GetByUserAndTarget(ctx context.Context, userID, targetID uuid.UUID, targetType string) (*entity.Reaction, error) {
	var rx entity.Reaction
	err := r.db.GetContext(ctx, &rx,
		`SELECT * FROM reactions WHERE user_id = $1 AND target_id = $2 AND target_type = $3`,
		userID, targetID, targetType)
	return &rx, err
}

func (r *ReactionRepo) HasReacted(ctx context.Context, userID, targetID uuid.UUID, targetType string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM reactions WHERE user_id = $1 AND target_id = $2 AND target_type = $3)`,
		userID, targetID, targetType)
	return exists, err
}

func (r *ReactionRepo) GetReactionCounts(ctx context.Context, targetID uuid.UUID, targetType string) (map[string]int, error) {
	type row struct {
		ReactionType string `db:"reaction_type"`
		Count        int    `db:"count"`
	}
	var rows []row
	err := r.db.SelectContext(ctx, &rows,
		`SELECT reaction_type, COUNT(*) as count FROM reactions WHERE target_id = $1 AND target_type = $2 GROUP BY reaction_type`,
		targetID, targetType)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, r := range rows {
		counts[r.ReactionType] = r.Count
	}
	return counts, nil
}

// --- SavedPostRepo ---

type SavedPostRepo struct{ db *sqlx.DB }

func NewSavedPostRepo(db *sqlx.DB) *SavedPostRepo { return &SavedPostRepo{db: db} }

func (r *SavedPostRepo) Save(ctx context.Context, userID, postID uuid.UUID, collection string) error {
	if collection == "" {
		collection = "default"
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO saved_posts (user_id, post_id, collection) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		userID, postID, collection)
	return err
}

func (r *SavedPostRepo) Unsave(ctx context.Context, userID, postID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM saved_posts WHERE user_id = $1 AND post_id = $2`, userID, postID)
	return err
}

func (r *SavedPostRepo) IsSaved(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM saved_posts WHERE user_id = $1 AND post_id = $2)`, userID, postID)
	return exists, err
}

func (r *SavedPostRepo) GetSavedPosts(ctx context.Context, userID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.SavedPost, error) {
	var saved []entity.SavedPost
	if cursorTime != nil {
		err := r.db.SelectContext(ctx, &saved,
			`SELECT * FROM saved_posts WHERE user_id = $1 AND created_at < $3 ORDER BY created_at DESC LIMIT $2`,
			userID, limit, cursorTime)
		return saved, err
	}
	err := r.db.SelectContext(ctx, &saved,
		`SELECT * FROM saved_posts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	return saved, err
}

// --- HashtagRepo ---

type HashtagRepo struct{ db *sqlx.DB }

func NewHashtagRepo(db *sqlx.DB) *HashtagRepo { return &HashtagRepo{db: db} }

func (r *HashtagRepo) UpsertBatch(ctx context.Context, names []string) error {
	for _, name := range names {
		_, err := r.db.ExecContext(ctx,
			`INSERT INTO hashtags (id, name, post_count) VALUES ($1, $2, 1)
			ON CONFLICT (name) DO UPDATE SET post_count = hashtags.post_count + 1`,
			uuid.New(), name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *HashtagRepo) IncrementCount(ctx context.Context, name string, delta int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE hashtags SET post_count = post_count + $1 WHERE name = $2`, delta, name)
	return err
}

func (r *HashtagRepo) GetTrending(ctx context.Context, limit int) ([]entity.Hashtag, error) {
	var tags []entity.Hashtag
	err := r.db.SelectContext(ctx, &tags,
		`SELECT * FROM hashtags ORDER BY post_count DESC LIMIT $1`, limit)
	return tags, err
}

// --- ReportRepo ---

type ReportRepo struct{ db *sqlx.DB }

func NewReportRepo(db *sqlx.DB) *ReportRepo { return &ReportRepo{db: db} }

func (r *ReportRepo) Create(ctx context.Context, report *entity.Report) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO reports (id, reporter_id, target_id, target_type, reason, description)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		report.ID, report.ReporterID, report.TargetID, report.TargetType, report.Reason, report.Description)
	return err
}

func (r *ReportRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Report, error) {
	var report entity.Report
	err := r.db.GetContext(ctx, &report, `SELECT * FROM reports WHERE id = $1`, id)
	return &report, err
}

func (r *ReportRepo) GetPending(ctx context.Context, limit int) ([]entity.Report, error) {
	var reports []entity.Report
	err := r.db.SelectContext(ctx, &reports,
		`SELECT * FROM reports WHERE status = 'PENDING' ORDER BY created_at LIMIT $1`, limit)
	return reports, err
}

// Ensure unused imports are satisfied
var (
	_ = sql.NullString{}
	_ = pq.StringArray{}
)

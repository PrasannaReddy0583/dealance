package application

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/rs/zerolog"

	"github.com/dealance/services/content/internal/domain/entity"
	"github.com/dealance/services/content/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

// PostService handles post CRUD operations.
type PostService struct {
	postRepo      repository.PostRepository
	mediaRepo     repository.PostMediaRepository
	hashtagRepo   repository.HashtagRepository
	savedPostRepo repository.SavedPostRepository
	reactionRepo  repository.ReactionRepository
	cacheRepo     repository.CacheRepository
	log           zerolog.Logger
}

func NewPostService(
	postRepo repository.PostRepository,
	mediaRepo repository.PostMediaRepository,
	hashtagRepo repository.HashtagRepository,
	savedPostRepo repository.SavedPostRepository,
	reactionRepo repository.ReactionRepository,
	cacheRepo repository.CacheRepository,
	log zerolog.Logger,
) *PostService {
	return &PostService{
		postRepo: postRepo, mediaRepo: mediaRepo, hashtagRepo: hashtagRepo,
		savedPostRepo: savedPostRepo, reactionRepo: reactionRepo, cacheRepo: cacheRepo, log: log,
	}
}

// CreatePost creates a new post with optional media.
func (s *PostService) CreatePost(ctx context.Context, authorID string, req entity.CreatePostRequest) (*entity.PostResponse, error) {
	aID, err := uuid.Parse(authorID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid author ID")
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = entity.VisibilityPublic
	}
	allowComments := true
	if req.AllowComments != nil {
		allowComments = *req.AllowComments
	}

	postID := uuid.New()
	post := &entity.Post{
		ID:            postID,
		AuthorID:      aID,
		PostType:      req.PostType,
		Title:         sql.NullString{String: req.Title, Valid: req.Title != ""},
		Body:          req.Body,
		Visibility:    visibility,
		IsPublished:   true,
		AllowComments: allowComments,
		Hashtags:      pq.StringArray(req.Hashtags),
		MentionIDs:    pq.StringArray(req.MentionIDs),
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Create media attachments
	if len(req.Media) > 0 {
		mediaItems := make([]entity.PostMedia, len(req.Media))
		for i, m := range req.Media {
			mediaItems[i] = entity.PostMedia{
				ID:           uuid.New(),
				PostID:       postID,
				MediaType:    m.MediaType,
				MediaURL:     m.MediaURL,
				ThumbnailURL: sql.NullString{String: m.ThumbnailURL, Valid: m.ThumbnailURL != ""},
				Width:        sql.NullInt32{Int32: int32(m.Width), Valid: m.Width > 0},
				Height:       sql.NullInt32{Int32: int32(m.Height), Valid: m.Height > 0},
				DurationMs:   sql.NullInt32{Int32: int32(m.DurationMs), Valid: m.DurationMs > 0},
				FileSize:     sql.NullInt64{Int64: m.FileSize, Valid: m.FileSize > 0},
				MimeType:     sql.NullString{String: m.MimeType, Valid: m.MimeType != ""},
				DisplayOrder: m.DisplayOrder,
			}
		}
		if err := s.mediaRepo.CreateBatch(ctx, mediaItems); err != nil {
			s.log.Error().Err(err).Msg("failed to create post media")
		}
	}

	// Track hashtags
	if len(req.Hashtags) > 0 {
		_ = s.hashtagRepo.UpsertBatch(ctx, req.Hashtags)
	}

	return s.toPostResponse(post, nil, false, false), nil
}

// GetPost returns a post with media.
func (s *PostService) GetPost(ctx context.Context, postID, viewerID string) (*entity.PostResponse, error) {
	pID, err := uuid.Parse(postID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid post ID")
	}

	post, err := s.postRepo.GetByID(ctx, pID)
	if err != nil {
		return nil, apperrors.ErrNotFound("Post")
	}

	media, _ := s.mediaRepo.GetByPostID(ctx, pID)

	var hasLiked, hasSaved bool
	if viewerID != "" {
		vID, _ := uuid.Parse(viewerID)
		hasLiked, _ = s.reactionRepo.HasReacted(ctx, vID, pID, "POST")
		hasSaved, _ = s.savedPostRepo.IsSaved(ctx, vID, pID)
	}

	// Increment view count async (fire and forget)
	_ = s.postRepo.IncrementCounter(ctx, pID, "view_count", 1)

	return s.toPostResponse(post, media, hasLiked, hasSaved), nil
}

// UpdatePost updates a post (author-only).
func (s *PostService) UpdatePost(ctx context.Context, postID, authorID string, req entity.UpdatePostRequest) (*entity.PostResponse, error) {
	pID, err := uuid.Parse(postID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid post ID")
	}

	post, err := s.postRepo.GetByID(ctx, pID)
	if err != nil {
		return nil, apperrors.ErrNotFound("Post")
	}
	if post.AuthorID.String() != authorID {
		return nil, apperrors.ErrForbidden("NOT_AUTHOR", "Only the author can edit this post")
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}
	if req.AllowComments != nil {
		updates["allow_comments"] = *req.AllowComments
	}
	if req.Hashtags != nil {
		updates["hashtags"] = pq.StringArray(req.Hashtags)
	}
	if req.IsPinned != nil {
		updates["is_pinned"] = *req.IsPinned
	}

	if len(updates) == 0 {
		return nil, apperrors.ErrValidation("no fields to update")
	}

	if err := s.postRepo.Update(ctx, pID, updates); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	updated, _ := s.postRepo.GetByID(ctx, pID)
	media, _ := s.mediaRepo.GetByPostID(ctx, pID)
	return s.toPostResponse(updated, media, false, false), nil
}

// DeletePost deletes a post (author-only).
func (s *PostService) DeletePost(ctx context.Context, postID, authorID string) error {
	pID, err := uuid.Parse(postID)
	if err != nil {
		return apperrors.ErrValidation("invalid post ID")
	}

	post, err := s.postRepo.GetByID(ctx, pID)
	if err != nil {
		return apperrors.ErrNotFound("Post")
	}
	if post.AuthorID.String() != authorID {
		return apperrors.ErrForbidden("NOT_AUTHOR", "Only the author can delete this post")
	}

	// Decrement hashtag counts
	if len(post.Hashtags) > 0 {
		for _, tag := range post.Hashtags {
			_ = s.hashtagRepo.IncrementCount(ctx, tag, -1)
		}
	}

	return s.postRepo.Delete(ctx, pID)
}

// GetUserPosts returns posts by a specific user.
func (s *PostService) GetUserPosts(ctx context.Context, authorID string, limit int) ([]entity.PostListItem, error) {
	aID, err := uuid.Parse(authorID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user ID")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	posts, err := s.postRepo.GetByAuthor(ctx, aID, limit, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	return s.toPostList(posts), nil
}

// GetFeed returns the global public feed.
func (s *PostService) GetFeed(ctx context.Context, limit int) ([]entity.PostListItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	posts, err := s.postRepo.GetFeed(ctx, limit, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	return s.toPostList(posts), nil
}

// GetByHashtag returns posts with a given hashtag.
func (s *PostService) GetByHashtag(ctx context.Context, hashtag string, limit int) ([]entity.PostListItem, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	posts, err := s.postRepo.GetByHashtag(ctx, hashtag, limit, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	return s.toPostList(posts), nil
}

// SavePost bookmarks a post.
func (s *PostService) SavePost(ctx context.Context, userID, postID, collection string) error {
	uID, _ := uuid.Parse(userID)
	pID, _ := uuid.Parse(postID)
	if err := s.savedPostRepo.Save(ctx, uID, pID, collection); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}
	_ = s.postRepo.IncrementCounter(ctx, pID, "save_count", 1)
	return nil
}

// UnsavePost removes a bookmark.
func (s *PostService) UnsavePost(ctx context.Context, userID, postID string) error {
	uID, _ := uuid.Parse(userID)
	pID, _ := uuid.Parse(postID)
	if err := s.savedPostRepo.Unsave(ctx, uID, pID); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}
	_ = s.postRepo.IncrementCounter(ctx, pID, "save_count", -1)
	return nil
}

// GetTrendingHashtags returns trending hashtags.
func (s *PostService) GetTrendingHashtags(ctx context.Context, limit int) ([]entity.Hashtag, error) {
	// Try cache
	if cached, err := s.cacheRepo.GetTrendingHashtags(ctx); err == nil && len(cached) > 0 {
		return cached, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	tags, err := s.hashtagRepo.GetTrending(ctx, limit)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}
	_ = s.cacheRepo.CacheTrendingHashtags(ctx, tags)
	return tags, nil
}

func (s *PostService) toPostResponse(p *entity.Post, media []entity.PostMedia, hasLiked, hasSaved bool) *entity.PostResponse {
	resp := &entity.PostResponse{
		ID:            p.ID.String(),
		AuthorID:      p.AuthorID.String(),
		PostType:      p.PostType,
		Body:          p.Body,
		Visibility:    p.Visibility,
		IsPublished:   p.IsPublished,
		IsPinned:      p.IsPinned,
		AllowComments: p.AllowComments,
		ViewCount:     p.ViewCount,
		LikeCount:     p.LikeCount,
		CommentCount:  p.CommentCount,
		ShareCount:    p.ShareCount,
		SaveCount:     p.SaveCount,
		HasLiked:      hasLiked,
		HasSaved:      hasSaved,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.Title.Valid {
		resp.Title = p.Title.String
	}
	if p.Hashtags != nil {
		resp.Hashtags = p.Hashtags
	} else {
		resp.Hashtags = []string{}
	}

	resp.Media = make([]entity.PostMediaResponse, len(media))
	for i, m := range media {
		resp.Media[i] = entity.PostMediaResponse{
			ID:           m.ID.String(),
			MediaType:    m.MediaType,
			MediaURL:     m.MediaURL,
			ThumbnailURL: m.ThumbnailURL.String,
			DisplayOrder: m.DisplayOrder,
		}
		if m.Width.Valid {
			resp.Media[i].Width = int(m.Width.Int32)
		}
		if m.Height.Valid {
			resp.Media[i].Height = int(m.Height.Int32)
		}
		if m.DurationMs.Valid {
			resp.Media[i].DurationMs = int(m.DurationMs.Int32)
		}
	}
	return resp
}

func (s *PostService) toPostList(posts []entity.Post) []entity.PostListItem {
	items := make([]entity.PostListItem, len(posts))
	for i, p := range posts {
		items[i] = entity.PostListItem{
			ID:           p.ID.String(),
			AuthorID:     p.AuthorID.String(),
			PostType:     p.PostType,
			Body:         p.Body,
			LikeCount:    p.LikeCount,
			CommentCount: p.CommentCount,
			CreatedAt:    p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if p.Title.Valid {
			items[i].Title = p.Title.String
		}
	}
	return items
}

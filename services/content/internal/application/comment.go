package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/content/internal/domain/entity"
	"github.com/dealance/services/content/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

// CommentService handles comment operations.
type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	log         zerolog.Logger
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	log zerolog.Logger,
) *CommentService {
	return &CommentService{commentRepo: commentRepo, postRepo: postRepo, log: log}
}

// CreateComment creates a new comment or reply.
func (s *CommentService) CreateComment(ctx context.Context, authorID string, req entity.CreateCommentRequest) (*entity.CommentResponse, error) {
	aID, err := uuid.Parse(authorID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid author ID")
	}
	postID, err := uuid.Parse(req.PostID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid post ID")
	}

	// Verify post exists and allows comments
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, apperrors.ErrNotFound("Post")
	}
	if !post.AllowComments {
		return nil, apperrors.ErrForbidden("COMMENTS_DISABLED", "Comments are disabled on this post")
	}

	comment := &entity.Comment{
		ID:       uuid.New(),
		PostID:   postID,
		AuthorID: aID,
		Body:     req.Body,
	}

	// Handle reply
	if req.ParentID != "" {
		parentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, apperrors.ErrValidation("invalid parent comment ID")
		}
		comment.ParentID = &parentID

		// Increment parent's reply count
		_ = s.commentRepo.IncrementCounter(ctx, parentID, "reply_count", 1)
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Increment post comment count
	_ = s.postRepo.IncrementCounter(ctx, postID, "comment_count", 1)

	return s.toCommentResponse(comment), nil
}

// GetComments returns top-level comments for a post.
func (s *CommentService) GetComments(ctx context.Context, postID string, limit int) ([]entity.CommentResponse, error) {
	pID, err := uuid.Parse(postID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid post ID")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	comments, err := s.commentRepo.GetByPostID(ctx, pID, limit, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	items := make([]entity.CommentResponse, len(comments))
	for i, c := range comments {
		items[i] = *s.toCommentResponse(&c)
	}
	return items, nil
}

// GetReplies returns replies to a comment.
func (s *CommentService) GetReplies(ctx context.Context, commentID string, limit int) ([]entity.CommentResponse, error) {
	cID, err := uuid.Parse(commentID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid comment ID")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	replies, err := s.commentRepo.GetReplies(ctx, cID, limit)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	items := make([]entity.CommentResponse, len(replies))
	for i, r := range replies {
		items[i] = *s.toCommentResponse(&r)
	}
	return items, nil
}

// UpdateComment edits a comment (author-only).
func (s *CommentService) UpdateComment(ctx context.Context, commentID, authorID string, req entity.UpdateCommentRequest) (*entity.CommentResponse, error) {
	cID, err := uuid.Parse(commentID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid comment ID")
	}

	comment, err := s.commentRepo.GetByID(ctx, cID)
	if err != nil {
		return nil, apperrors.ErrNotFound("Comment")
	}
	if comment.AuthorID.String() != authorID {
		return nil, apperrors.ErrForbidden("NOT_AUTHOR", "Only the author can edit this comment")
	}

	if err := s.commentRepo.Update(ctx, cID, req.Body); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	comment.Body = req.Body
	comment.IsEdited = true
	return s.toCommentResponse(comment), nil
}

// DeleteComment soft-deletes a comment (author-only, preserves thread).
func (s *CommentService) DeleteComment(ctx context.Context, commentID, authorID string) error {
	cID, err := uuid.Parse(commentID)
	if err != nil {
		return apperrors.ErrValidation("invalid comment ID")
	}

	comment, err := s.commentRepo.GetByID(ctx, cID)
	if err != nil {
		return apperrors.ErrNotFound("Comment")
	}
	if comment.AuthorID.String() != authorID {
		return apperrors.ErrForbidden("NOT_AUTHOR", "Only the author can delete this comment")
	}

	if err := s.commentRepo.SoftDelete(ctx, cID); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Decrement post comment count
	_ = s.postRepo.IncrementCounter(ctx, comment.PostID, "comment_count", -1)

	return nil
}

func (s *CommentService) toCommentResponse(c *entity.Comment) *entity.CommentResponse {
	resp := &entity.CommentResponse{
		ID:         c.ID.String(),
		PostID:     c.PostID.String(),
		AuthorID:   c.AuthorID.String(),
		Body:       c.Body,
		LikeCount:  c.LikeCount,
		ReplyCount: c.ReplyCount,
		IsEdited:   c.IsEdited,
		CreatedAt:  c.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if c.ParentID != nil {
		resp.ParentID = c.ParentID.String()
	}
	return resp
}

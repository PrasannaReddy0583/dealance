package application

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/content/internal/domain/entity"
	"github.com/dealance/services/content/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

// ReactionService handles reactions and reports.
type ReactionService struct {
	reactionRepo repository.ReactionRepository
	postRepo     repository.PostRepository
	commentRepo  repository.CommentRepository
	reportRepo   repository.ReportRepository
	cacheRepo    repository.CacheRepository
	log          zerolog.Logger
}

func NewReactionService(
	reactionRepo repository.ReactionRepository,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
	reportRepo repository.ReportRepository,
	cacheRepo repository.CacheRepository,
	log zerolog.Logger,
) *ReactionService {
	return &ReactionService{
		reactionRepo: reactionRepo, postRepo: postRepo,
		commentRepo: commentRepo, reportRepo: reportRepo,
		cacheRepo: cacheRepo, log: log,
	}
}

// React adds or updates a reaction.
func (s *ReactionService) React(ctx context.Context, userID string, req entity.ReactRequest) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.ErrValidation("invalid user ID")
	}
	tID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return apperrors.ErrValidation("invalid target ID")
	}

	// Check if already reacted (for counter tracking)
	alreadyReacted, _ := s.reactionRepo.HasReacted(ctx, uID, tID, req.TargetType)

	rx := &entity.Reaction{
		ID:           uuid.New(),
		UserID:       uID,
		TargetID:     tID,
		TargetType:   req.TargetType,
		ReactionType: req.ReactionType,
	}

	if err := s.reactionRepo.Create(ctx, rx); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Update counter only if new reaction
	if !alreadyReacted {
		if req.TargetType == "POST" {
			_ = s.postRepo.IncrementCounter(ctx, tID, "like_count", 1)
		} else if req.TargetType == "COMMENT" {
			_ = s.commentRepo.IncrementCounter(ctx, tID, "like_count", 1)
		}
		_ = s.cacheRepo.IncrReactionCount(ctx, req.TargetID, req.TargetType, 1)
	}

	return nil
}

// Unreact removes a reaction.
func (s *ReactionService) Unreact(ctx context.Context, userID string, req entity.UnreactRequest) error {
	uID, err := uuid.Parse(userID)
	if err != nil {
		return apperrors.ErrValidation("invalid user ID")
	}
	tID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return apperrors.ErrValidation("invalid target ID")
	}

	// Check if reaction exists
	exists, _ := s.reactionRepo.HasReacted(ctx, uID, tID, req.TargetType)
	if !exists {
		return nil // Idempotent
	}

	if err := s.reactionRepo.Delete(ctx, uID, tID, req.TargetType); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Decrement counter
	if req.TargetType == "POST" {
		_ = s.postRepo.IncrementCounter(ctx, tID, "like_count", -1)
	} else if req.TargetType == "COMMENT" {
		_ = s.commentRepo.IncrementCounter(ctx, tID, "like_count", -1)
	}
	_ = s.cacheRepo.IncrReactionCount(ctx, req.TargetID, req.TargetType, -1)

	return nil
}

// ReportContent creates a content moderation report.
func (s *ReactionService) ReportContent(ctx context.Context, reporterID string, req entity.ReportRequest) error {
	rID, err := uuid.Parse(reporterID)
	if err != nil {
		return apperrors.ErrValidation("invalid reporter ID")
	}
	tID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return apperrors.ErrValidation("invalid target ID")
	}

	report := &entity.Report{
		ID:          uuid.New(),
		ReporterID:  rID,
		TargetID:    tID,
		TargetType:  req.TargetType,
		Reason:      req.Reason,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Status:      "PENDING",
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	return nil
}

package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/user/internal/domain/entity"
	"github.com/dealance/services/user/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

// FollowService handles follow/unfollow operations.
type FollowService struct {
	followRepo  repository.FollowRepository
	blockRepo   repository.BlockRepository
	profileRepo repository.ProfileRepository
	cacheRepo   repository.CacheRepository
	log         zerolog.Logger
}

func NewFollowService(
	followRepo repository.FollowRepository,
	blockRepo repository.BlockRepository,
	profileRepo repository.ProfileRepository,
	cacheRepo repository.CacheRepository,
	log zerolog.Logger,
) *FollowService {
	return &FollowService{
		followRepo:  followRepo,
		blockRepo:   blockRepo,
		profileRepo: profileRepo,
		cacheRepo:   cacheRepo,
		log:         log,
	}
}

// Follow creates a follow relationship.
func (s *FollowService) Follow(ctx context.Context, followerID, targetUserID string) error {
	fID, err := uuid.Parse(followerID)
	if err != nil {
		return apperrors.ErrValidation("invalid follower ID")
	}
	tID, err := uuid.Parse(targetUserID)
	if err != nil {
		return apperrors.ErrValidation("invalid target user ID")
	}

	// Can't follow yourself
	if fID == tID {
		return apperrors.ErrValidation("cannot follow yourself")
	}

	// Check if blocked (either direction)
	blocked, _ := s.blockRepo.IsBlocked(ctx, tID, fID)
	if blocked {
		return apperrors.ErrForbidden("BLOCKED", "Cannot follow this user")
	}

	// Check if already following
	already, _ := s.followRepo.IsFollowing(ctx, fID, tID)
	if already {
		return nil // Idempotent
	}

	// Create follow
	if err := s.followRepo.Follow(ctx, fID, tID); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Update counters atomically
	_ = s.profileRepo.IncrementCounter(ctx, fID, "following_count", 1)
	_ = s.profileRepo.IncrementCounter(ctx, tID, "follower_count", 1)

	// Update cache
	_ = s.cacheRepo.IncrFollowingCount(ctx, followerID, 1)
	_ = s.cacheRepo.IncrFollowCount(ctx, targetUserID, 1)
	_ = s.cacheRepo.InvalidateProfile(ctx, followerID)
	_ = s.cacheRepo.InvalidateProfile(ctx, targetUserID)

	return nil
}

// Unfollow removes a follow relationship.
func (s *FollowService) Unfollow(ctx context.Context, followerID, targetUserID string) error {
	fID, err := uuid.Parse(followerID)
	if err != nil {
		return apperrors.ErrValidation("invalid follower ID")
	}
	tID, err := uuid.Parse(targetUserID)
	if err != nil {
		return apperrors.ErrValidation("invalid target user ID")
	}

	// Check if following
	following, _ := s.followRepo.IsFollowing(ctx, fID, tID)
	if !following {
		return nil // Idempotent
	}

	if err := s.followRepo.Unfollow(ctx, fID, tID); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Update counters
	_ = s.profileRepo.IncrementCounter(ctx, fID, "following_count", -1)
	_ = s.profileRepo.IncrementCounter(ctx, tID, "follower_count", -1)

	// Update cache
	_ = s.cacheRepo.IncrFollowingCount(ctx, followerID, -1)
	_ = s.cacheRepo.IncrFollowCount(ctx, targetUserID, -1)
	_ = s.cacheRepo.InvalidateProfile(ctx, followerID)
	_ = s.cacheRepo.InvalidateProfile(ctx, targetUserID)

	return nil
}

// GetFollowers returns a paginated list of followers.
func (s *FollowService) GetFollowers(ctx context.Context, userID string, limit int) ([]entity.ProfileListItem, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user ID")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	follows, err := s.followRepo.GetFollowers(ctx, id, limit, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	items := make([]entity.ProfileListItem, 0, len(follows))
	for _, f := range follows {
		profile, err := s.profileRepo.GetByID(ctx, f.FollowerID)
		if err != nil {
			continue
		}
		items = append(items, entity.ProfileListItem{
			ID:          profile.ID.String(),
			Username:    profile.Username,
			DisplayName: profile.DisplayName,
			AvatarURL:   profile.AvatarURL,
			Bio:         profile.Bio,
			Profession:  profile.Profession,
			IsVerified:  profile.IsVerified,
		})
	}

	return items, nil
}

// GetFollowing returns a paginated list of following users.
func (s *FollowService) GetFollowing(ctx context.Context, userID string, limit int) ([]entity.ProfileListItem, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user ID")
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	follows, err := s.followRepo.GetFollowing(ctx, id, limit, nil)
	if err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	items := make([]entity.ProfileListItem, 0, len(follows))
	for _, f := range follows {
		profile, err := s.profileRepo.GetByID(ctx, f.FollowingID)
		if err != nil {
			continue
		}
		items = append(items, entity.ProfileListItem{
			ID:          profile.ID.String(),
			Username:    profile.Username,
			DisplayName: profile.DisplayName,
			AvatarURL:   profile.AvatarURL,
			Bio:         profile.Bio,
			Profession:  profile.Profession,
			IsVerified:  profile.IsVerified,
		})
	}

	return items, nil
}

// GetFollowCounts returns follower and following counts.
func (s *FollowService) GetFollowCounts(ctx context.Context, userID string) (*entity.FollowCountsResponse, error) {
	// Try cache
	followers, following, err := s.cacheRepo.GetFollowCounts(ctx, userID)
	if err == nil {
		return &entity.FollowCountsResponse{
			FollowerCount:  followers,
			FollowingCount: following,
		}, nil
	}

	// Cache miss — query DB
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, apperrors.ErrValidation("invalid user ID")
	}
	followers, _ = s.followRepo.GetFollowerCount(ctx, id)
	following, _ = s.followRepo.GetFollowingCount(ctx, id)

	// Cache for next time
	_ = s.cacheRepo.CacheFollowCounts(ctx, userID, followers, following)

	return &entity.FollowCountsResponse{
		FollowerCount:  followers,
		FollowingCount: following,
	}, nil
}

// BlockUser blocks another user and auto-unfollows both directions.
func (s *FollowService) BlockUser(ctx context.Context, blockerID, targetUserID, reason string) error {
	bID, err := uuid.Parse(blockerID)
	if err != nil {
		return apperrors.ErrValidation("invalid blocker ID")
	}
	tID, err := uuid.Parse(targetUserID)
	if err != nil {
		return apperrors.ErrValidation("invalid target user ID")
	}

	if bID == tID {
		return apperrors.ErrValidation("cannot block yourself")
	}

	// Block
	if err := s.blockRepo.Block(ctx, bID, tID, reason); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	// Auto-unfollow both directions
	_ = s.Unfollow(ctx, blockerID, targetUserID)
	_ = s.Unfollow(ctx, targetUserID, blockerID)

	return nil
}

// UnblockUser removes a block.
func (s *FollowService) UnblockUser(ctx context.Context, blockerID, targetUserID string) error {
	bID, err := uuid.Parse(blockerID)
	if err != nil {
		return apperrors.ErrValidation("invalid blocker ID")
	}
	tID, err := uuid.Parse(targetUserID)
	if err != nil {
		return apperrors.ErrValidation("invalid target user ID")
	}

	if err := s.blockRepo.Unblock(ctx, bID, tID); err != nil {
		return apperrors.ErrInternal().WithInternal(err)
	}

	return nil
}

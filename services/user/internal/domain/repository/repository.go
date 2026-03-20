package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dealance/services/user/internal/domain/entity"
)

// ProfileRepository handles profile CRUD.
type ProfileRepository interface {
	Create(ctx context.Context, profile *entity.Profile) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error)
	GetByUsername(ctx context.Context, username string) (*entity.Profile, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error
	Search(ctx context.Context, query string, limit int, cursorTime *time.Time, cursorID *string) ([]entity.Profile, error)
}

// FollowRepository handles follow/unfollow relationships.
type FollowRepository interface {
	Follow(ctx context.Context, followerID, followingID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error
	IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error)
	GetFollowers(ctx context.Context, userID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Follow, error)
	GetFollowing(ctx context.Context, userID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Follow, error)
	GetFollowerCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetFollowingCount(ctx context.Context, userID uuid.UUID) (int, error)
	GetMutualFollowers(ctx context.Context, userA, userB uuid.UUID, limit int) ([]uuid.UUID, error)
}

// BlockRepository handles blocked user relationships.
type BlockRepository interface {
	Block(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error
	Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error
	IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)
	GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]entity.BlockedUser, error)
}

// SettingsRepository handles user settings.
type SettingsRepository interface {
	Create(ctx context.Context, userID uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error)
	Update(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error
}

// ProfileMediaRepository handles profile media items.
type ProfileMediaRepository interface {
	Create(ctx context.Context, media *entity.ProfileMedia) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ProfileMedia, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, mediaType string) ([]entity.ProfileMedia, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// EntrepreneurProfileRepository handles entrepreneur-specific data.
type EntrepreneurProfileRepository interface {
	Upsert(ctx context.Context, ep *entity.EntrepreneurProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.EntrepreneurProfile, error)
}

// InvestorProfileRepository handles investor-specific data.
type InvestorProfileRepository interface {
	Upsert(ctx context.Context, ip *entity.InvestorProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.InvestorProfile, error)
}

// CacheRepository handles Redis caching for user data.
type CacheRepository interface {
	CacheProfile(ctx context.Context, profile *entity.Profile) error
	GetCachedProfile(ctx context.Context, userID string) (*entity.Profile, error)
	InvalidateProfile(ctx context.Context, userID string) error
	CacheFollowCounts(ctx context.Context, userID string, followers, following int) error
	GetFollowCounts(ctx context.Context, userID string) (followers int, following int, err error)
	IncrFollowCount(ctx context.Context, userID string, delta int) error
	IncrFollowingCount(ctx context.Context, userID string, delta int) error
}

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/dealance/services/startup/internal/domain/entity"
)

type StartupRepository interface {
	Create(ctx context.Context, s *entity.Startup) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Startup, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Startup, error)
	GetByFounder(ctx context.Context, founderID uuid.UUID) ([]entity.Startup, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
	Search(ctx context.Context, query, sector, stage, country string, limit int, cursorTime *time.Time) ([]entity.Startup, error)
	IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error
}

type FundingRoundRepository interface {
	Create(ctx context.Context, r *entity.FundingRound) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.FundingRound, error)
	GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.FundingRound, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type TeamMemberRepository interface {
	Create(ctx context.Context, m *entity.TeamMember) error
	GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.TeamMember, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type StartupMediaRepository interface {
	Create(ctx context.Context, m *entity.StartupMedia) error
	GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.StartupMedia, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type StartupMetricRepository interface {
	Create(ctx context.Context, m *entity.StartupMetric) error
	GetByStartup(ctx context.Context, startupID uuid.UUID, metricType string) ([]entity.StartupMetric, error)
	GetLatest(ctx context.Context, startupID uuid.UUID) ([]entity.StartupMetric, error)
}

type StartupFollowRepository interface {
	Follow(ctx context.Context, userID, startupID uuid.UUID) error
	Unfollow(ctx context.Context, userID, startupID uuid.UUID) error
	IsFollowing(ctx context.Context, userID, startupID uuid.UUID) (bool, error)
	GetFollowers(ctx context.Context, startupID uuid.UUID, limit int) ([]entity.StartupFollow, error)
}

type CacheRepository interface {
	CacheStartup(ctx context.Context, startup *entity.Startup) error
	GetCachedStartup(ctx context.Context, id string) (*entity.Startup, error)
	InvalidateStartup(ctx context.Context, id string) error
}

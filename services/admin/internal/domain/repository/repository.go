package repository

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/dealance/services/admin/internal/domain/entity"
)

type AdminUserRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.AdminUser, error)
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AdminAuditLog) error
	GetRecent(ctx context.Context, limit int) ([]entity.AdminAuditLog, error)
	GetByAdmin(ctx context.Context, adminID uuid.UUID, limit int) ([]entity.AdminAuditLog, error)
}

type StatsRepository interface {
	GetLatest(ctx context.Context) (*entity.PlatformStats, error)
	GetByDateRange(ctx context.Context, from, to time.Time) ([]entity.PlatformStats, error)
	Upsert(ctx context.Context, stats *entity.PlatformStats) error
}

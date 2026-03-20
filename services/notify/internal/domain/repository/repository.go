package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/dealance/services/notify/internal/domain/entity"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *entity.Notification) error
	GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
}

type DeviceTokenRepository interface {
	Upsert(ctx context.Context, d *entity.DeviceToken) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.DeviceToken, error)
	Deactivate(ctx context.Context, token string) error
}

type PreferencesRepository interface {
	Upsert(ctx context.Context, p *entity.NotificationPreferences) error
	GetByUser(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error)
}

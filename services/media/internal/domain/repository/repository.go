package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/dealance/services/media/internal/domain/entity"
)

type MediaRepository interface {
	Create(ctx context.Context, m *entity.MediaUpload) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.MediaUpload, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.MediaUpload, error)
	GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]entity.MediaUpload, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type VariantRepository interface {
	Create(ctx context.Context, v *entity.MediaVariant) error
	GetByUpload(ctx context.Context, uploadID uuid.UUID) ([]entity.MediaVariant, error)
}

type PresignedURLRepository interface {
	Create(ctx context.Context, p *entity.PresignedURL) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PresignedURL, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

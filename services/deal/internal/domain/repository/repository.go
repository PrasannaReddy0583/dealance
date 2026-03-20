package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/dealance/services/deal/internal/domain/entity"
)

type DealRepository interface {
	Create(ctx context.Context, d *entity.Deal) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Deal, error)
	GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.Deal, error)
	GetByCreator(ctx context.Context, creatorID uuid.UUID) ([]entity.Deal, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type ParticipantRepository interface {
	Create(ctx context.Context, p *entity.DealParticipant) error
	GetByDealAndUser(ctx context.Context, dealID, userID uuid.UUID) (*entity.DealParticipant, error)
	GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealParticipant, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.DealParticipant, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type DocumentRepository interface {
	Create(ctx context.Context, d *entity.DealDocument) error
	GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealDocument, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type MilestoneRepository interface {
	Create(ctx context.Context, m *entity.DealMilestone) error
	GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealMilestone, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type NDARepository interface {
	Create(ctx context.Context, n *entity.DealNDA) error
	GetByDealAndUser(ctx context.Context, dealID, userID uuid.UUID) (*entity.DealNDA, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type NegotiationRepository interface {
	Create(ctx context.Context, n *entity.DealNegotiation) error
	GetByDeal(ctx context.Context, dealID uuid.UUID, limit int) ([]entity.DealNegotiation, error)
}

type EscrowRepository interface {
	Create(ctx context.Context, e *entity.DealEscrow) error
	GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealEscrow, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type CacheRepository interface {
	CacheDeal(ctx context.Context, deal *entity.Deal) error
	GetCachedDeal(ctx context.Context, id string) (*entity.Deal, error)
	InvalidateDeal(ctx context.Context, id string) error
}

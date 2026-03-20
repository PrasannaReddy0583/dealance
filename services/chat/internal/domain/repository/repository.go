package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/dealance/services/chat/internal/domain/entity"
)

type ConversationRepository interface {
	Create(ctx context.Context, c *entity.Conversation) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Conversation, error)
	UpdateLastMessage(ctx context.Context, id uuid.UUID) error
}

type ParticipantRepository interface {
	Add(ctx context.Context, p *entity.ConversationParticipant) error
	GetByConversation(ctx context.Context, convID uuid.UUID) ([]entity.ConversationParticipant, error)
	GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.ConversationParticipant, error)
	IsParticipant(ctx context.Context, convID, userID uuid.UUID) (bool, error)
	IncrementUnread(ctx context.Context, convID uuid.UUID, excludeUserID uuid.UUID) error
	ResetUnread(ctx context.Context, convID, userID uuid.UUID) error
}

type MessageRepository interface {
	Create(ctx context.Context, m *entity.Message) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Message, error)
	GetByConversation(ctx context.Context, convID uuid.UUID, limit int) ([]entity.Message, error)
	Update(ctx context.Context, id uuid.UUID, body string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}

type ReadReceiptRepository interface {
	Create(ctx context.Context, r *entity.ReadReceipt) error
	GetByMessage(ctx context.Context, messageID uuid.UUID) ([]entity.ReadReceipt, error)
}

type CacheRepository interface {
	SetUserOnline(ctx context.Context, userID string) error
	SetUserOffline(ctx context.Context, userID string) error
	IsOnline(ctx context.Context, userID string) (bool, error)
	PublishMessage(ctx context.Context, convID string, payload []byte) error
}

package entity

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	ConvType      string         `db:"conv_type" json:"conv_type"`
	Title         sql.NullString `db:"title" json:"title,omitempty"`
	DealID        *uuid.UUID     `db:"deal_id" json:"deal_id,omitempty"`
	CreatedBy     uuid.UUID      `db:"created_by" json:"created_by"`
	LastMessageAt time.Time      `db:"last_message_at" json:"last_message_at"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

type ConversationParticipant struct {
	ConversationID uuid.UUID `db:"conversation_id" json:"conversation_id"`
	UserID         uuid.UUID `db:"user_id" json:"user_id"`
	Role           string    `db:"role" json:"role"`
	UnreadCount    int       `db:"unread_count" json:"unread_count"`
	Muted          bool      `db:"muted" json:"muted"`
	LastReadAt     time.Time `db:"last_read_at" json:"last_read_at"`
	JoinedAt       time.Time `db:"joined_at" json:"joined_at"`
}

type Message struct {
	ID             uuid.UUID      `db:"id" json:"id"`
	ConversationID uuid.UUID      `db:"conversation_id" json:"conversation_id"`
	SenderID       uuid.UUID      `db:"sender_id" json:"sender_id"`
	MessageType    string         `db:"message_type" json:"message_type"`
	Body           string         `db:"body" json:"body"`
	MediaURL       sql.NullString `db:"media_url" json:"media_url,omitempty"`
	MediaType      sql.NullString `db:"media_type" json:"media_type,omitempty"`
	FileName       sql.NullString `db:"file_name" json:"file_name,omitempty"`
	FileSize       sql.NullInt64  `db:"file_size" json:"file_size,omitempty"`
	ReplyToID      *uuid.UUID     `db:"reply_to_id" json:"reply_to_id,omitempty"`
	IsEdited       bool           `db:"is_edited" json:"is_edited"`
	IsDeleted      bool           `db:"is_deleted" json:"is_deleted"`
	CreatedAt      time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updated_at"`
}

type ReadReceipt struct {
	MessageID uuid.UUID `db:"message_id" json:"message_id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	ReadAt    time.Time `db:"read_at" json:"read_at"`
}

type MessageReaction struct {
	ID        uuid.UUID `db:"id" json:"id"`
	MessageID uuid.UUID `db:"message_id" json:"message_id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Emoji     string    `db:"emoji" json:"emoji"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

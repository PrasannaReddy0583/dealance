package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/dealance/services/chat/internal/domain/entity"
)

// --- ConversationRepo ---
type ConversationRepo struct{ db *sqlx.DB }
func NewConversationRepo(db *sqlx.DB) *ConversationRepo { return &ConversationRepo{db: db} }

func (r *ConversationRepo) Create(ctx context.Context, c *entity.Conversation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO conversations (id, conv_type, title, deal_id, created_by) VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.ConvType, c.Title, c.DealID, c.CreatedBy)
	return err
}
func (r *ConversationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error) {
	var c entity.Conversation; return &c, r.db.GetContext(ctx, &c, `SELECT * FROM conversations WHERE id = $1`, id)
}
func (r *ConversationRepo) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Conversation, error) {
	var convs []entity.Conversation
	return convs, r.db.SelectContext(ctx, &convs,
		`SELECT c.* FROM conversations c JOIN conversation_participants p ON c.id = p.conversation_id WHERE p.user_id = $1 ORDER BY c.last_message_at DESC LIMIT $2`, userID, limit)
}
func (r *ConversationRepo) UpdateLastMessage(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE conversations SET last_message_at = NOW() WHERE id = $1`, id); return err
}

// --- ParticipantRepo ---
type ParticipantRepo struct{ db *sqlx.DB }
func NewParticipantRepo(db *sqlx.DB) *ParticipantRepo { return &ParticipantRepo{db: db} }

func (r *ParticipantRepo) Add(ctx context.Context, p *entity.ConversationParticipant) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO conversation_participants (conversation_id, user_id, role) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		p.ConversationID, p.UserID, p.Role)
	return err
}
func (r *ParticipantRepo) GetByConversation(ctx context.Context, convID uuid.UUID) ([]entity.ConversationParticipant, error) {
	var ps []entity.ConversationParticipant
	return ps, r.db.SelectContext(ctx, &ps, `SELECT * FROM conversation_participants WHERE conversation_id = $1`, convID)
}
func (r *ParticipantRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.ConversationParticipant, error) {
	var ps []entity.ConversationParticipant
	return ps, r.db.SelectContext(ctx, &ps, `SELECT * FROM conversation_participants WHERE user_id = $1`, userID)
}
func (r *ParticipantRepo) IsParticipant(ctx context.Context, convID, userID uuid.UUID) (bool, error) {
	var exists bool
	return exists, r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2)`, convID, userID)
}
func (r *ParticipantRepo) IncrementUnread(ctx context.Context, convID uuid.UUID, excludeUserID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversation_participants SET unread_count = unread_count + 1 WHERE conversation_id = $1 AND user_id != $2`, convID, excludeUserID)
	return err
}
func (r *ParticipantRepo) ResetUnread(ctx context.Context, convID, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE conversation_participants SET unread_count = 0, last_read_at = NOW() WHERE conversation_id = $1 AND user_id = $2`, convID, userID)
	return err
}

// --- MessageRepo ---
type MessageRepo struct{ db *sqlx.DB }
func NewMessageRepo(db *sqlx.DB) *MessageRepo { return &MessageRepo{db: db} }

func (r *MessageRepo) Create(ctx context.Context, m *entity.Message) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO messages (id, conversation_id, sender_id, message_type, body, media_url, reply_to_id) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		m.ID, m.ConversationID, m.SenderID, m.MessageType, m.Body, m.MediaURL, m.ReplyToID)
	return err
}
func (r *MessageRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Message, error) {
	var m entity.Message; return &m, r.db.GetContext(ctx, &m, `SELECT * FROM messages WHERE id = $1`, id)
}
func (r *MessageRepo) GetByConversation(ctx context.Context, convID uuid.UUID, limit int) ([]entity.Message, error) {
	var ms []entity.Message
	return ms, r.db.SelectContext(ctx, &ms,
		`SELECT * FROM messages WHERE conversation_id = $1 AND is_deleted = false ORDER BY created_at DESC LIMIT $2`, convID, limit)
}
func (r *MessageRepo) Update(ctx context.Context, id uuid.UUID, body string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE messages SET body = $1, is_edited = true WHERE id = $2`, body, id); return err
}
func (r *MessageRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE messages SET is_deleted = true, body = '[deleted]' WHERE id = $1`, id); return err
}

// --- ReadReceiptRepo ---
type ReadReceiptRepo struct{ db *sqlx.DB }
func NewReadReceiptRepo(db *sqlx.DB) *ReadReceiptRepo { return &ReadReceiptRepo{db: db} }

func (r *ReadReceiptRepo) Create(ctx context.Context, rr *entity.ReadReceipt) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO read_receipts (message_id, user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, rr.MessageID, rr.UserID)
	return err
}
func (r *ReadReceiptRepo) GetByMessage(ctx context.Context, messageID uuid.UUID) ([]entity.ReadReceipt, error) {
	var rrs []entity.ReadReceipt
	return rrs, r.db.SelectContext(ctx, &rrs, `SELECT * FROM read_receipts WHERE message_id = $1`, messageID)
}

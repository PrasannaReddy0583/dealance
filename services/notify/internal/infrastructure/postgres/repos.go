package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/dealance/services/notify/internal/domain/entity"
)

type NotificationRepo struct{ db *sqlx.DB }
func NewNotificationRepo(db *sqlx.DB) *NotificationRepo { return &NotificationRepo{db: db} }

func (r *NotificationRepo) Create(ctx context.Context, n *entity.Notification) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (id,user_id,notif_type,title,body,channel,entity_type,entity_id,status) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		n.ID, n.UserID, n.NotifType, n.Title, n.Body, n.Channel, n.EntityType, n.EntityID, n.Status)
	return err
}
func (r *NotificationRepo) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.Notification, error) {
	var ns []entity.Notification
	return ns, r.db.SelectContext(ctx, &ns, `SELECT * FROM notifications WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
}
func (r *NotificationRepo) MarkRead(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read=true, read_at=NOW() WHERE id=$1`, id); return err
}
func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notifications SET is_read=true, read_at=NOW() WHERE user_id=$1 AND is_read=false`, userID); return err
}
func (r *NotificationRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int; return count, r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND is_read=false`, userID)
}

type DeviceTokenRepo struct{ db *sqlx.DB }
func NewDeviceTokenRepo(db *sqlx.DB) *DeviceTokenRepo { return &DeviceTokenRepo{db: db} }

func (r *DeviceTokenRepo) Upsert(ctx context.Context, d *entity.DeviceToken) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO device_tokens (id,user_id,device_token,platform) VALUES ($1,$2,$3,$4) ON CONFLICT (device_token) DO UPDATE SET user_id=$2, is_active=true`,
		d.ID, d.UserID, d.DeviceToken, d.Platform)
	return err
}
func (r *DeviceTokenRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.DeviceToken, error) {
	var ds []entity.DeviceToken
	return ds, r.db.SelectContext(ctx, &ds, `SELECT * FROM device_tokens WHERE user_id=$1 AND is_active=true`, userID)
}
func (r *DeviceTokenRepo) Deactivate(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE device_tokens SET is_active=false WHERE device_token=$1`, token); return err
}

type PreferencesRepo struct{ db *sqlx.DB }
func NewPreferencesRepo(db *sqlx.DB) *PreferencesRepo { return &PreferencesRepo{db: db} }

func (r *PreferencesRepo) Upsert(ctx context.Context, p *entity.NotificationPreferences) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notification_preferences (user_id,push_enabled,email_enabled,deal_updates,chat_messages,content_reactions,new_followers,marketing)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT (user_id) DO UPDATE SET push_enabled=$2,email_enabled=$3,deal_updates=$4,chat_messages=$5,content_reactions=$6,new_followers=$7,marketing=$8`,
		p.UserID, p.PushEnabled, p.EmailEnabled, p.DealUpdates, p.ChatMessages, p.ContentReactions, p.NewFollowers, p.Marketing)
	return err
}
func (r *PreferencesRepo) GetByUser(ctx context.Context, userID uuid.UUID) (*entity.NotificationPreferences, error) {
	var p entity.NotificationPreferences; return &p, r.db.GetContext(ctx, &p, `SELECT * FROM notification_preferences WHERE user_id=$1`, userID)
}

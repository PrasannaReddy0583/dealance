package entity

import (
	"database/sql"
	"time"
	"github.com/google/uuid"
)

type Notification struct {
	ID         uuid.UUID      `db:"id" json:"id"`
	UserID     uuid.UUID      `db:"user_id" json:"user_id"`
	NotifType  string         `db:"notif_type" json:"notif_type"`
	Title      string         `db:"title" json:"title"`
	Body       string         `db:"body" json:"body"`
	Channel    string         `db:"channel" json:"channel"`
	EntityType sql.NullString `db:"entity_type" json:"entity_type,omitempty"`
	EntityID   *uuid.UUID     `db:"entity_id" json:"entity_id,omitempty"`
	IsRead     bool           `db:"is_read" json:"is_read"`
	ReadAt     sql.NullTime   `db:"read_at" json:"read_at,omitempty"`
	SentAt     sql.NullTime   `db:"sent_at" json:"sent_at,omitempty"`
	Status     string         `db:"status" json:"status"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}

type DeviceToken struct {
	ID          uuid.UUID `db:"id" json:"id"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	DeviceToken string    `db:"device_token" json:"device_token"`
	Platform    string    `db:"platform" json:"platform"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type NotificationPreferences struct {
	UserID          uuid.UUID `db:"user_id" json:"user_id"`
	PushEnabled     bool      `db:"push_enabled" json:"push_enabled"`
	EmailEnabled    bool      `db:"email_enabled" json:"email_enabled"`
	SMSEnabled      bool      `db:"sms_enabled" json:"sms_enabled"`
	DealUpdates     bool      `db:"deal_updates" json:"deal_updates"`
	ChatMessages    bool      `db:"chat_messages" json:"chat_messages"`
	ContentReactions bool     `db:"content_reactions" json:"content_reactions"`
	NewFollowers    bool      `db:"new_followers" json:"new_followers"`
	Marketing       bool      `db:"marketing" json:"marketing"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type SendNotificationRequest struct {
	UserID     string `json:"user_id" validate:"required,uuid"`
	NotifType  string `json:"notif_type" validate:"required"`
	Title      string `json:"title" validate:"required,max=200"`
	Body       string `json:"body" validate:"required"`
	Channel    string `json:"channel,omitempty" validate:"omitempty,oneof=PUSH EMAIL SMS IN_APP"`
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
}

type RegisterDeviceRequest struct {
	DeviceToken string `json:"device_token" validate:"required"`
	Platform    string `json:"platform" validate:"required,oneof=IOS ANDROID WEB"`
}

type UpdatePreferencesRequest struct {
	PushEnabled      *bool `json:"push_enabled,omitempty"`
	EmailEnabled     *bool `json:"email_enabled,omitempty"`
	DealUpdates      *bool `json:"deal_updates,omitempty"`
	ChatMessages     *bool `json:"chat_messages,omitempty"`
	ContentReactions *bool `json:"content_reactions,omitempty"`
	NewFollowers     *bool `json:"new_followers,omitempty"`
	Marketing        *bool `json:"marketing,omitempty"`
}

package entity

import (
	"database/sql"
	"time"
	"github.com/google/uuid"
)

type MediaUpload struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	UserID       uuid.UUID      `db:"user_id" json:"user_id"`
	FileName     string         `db:"file_name" json:"file_name"`
	OriginalName string         `db:"original_name" json:"original_name"`
	MimeType     string         `db:"mime_type" json:"mime_type"`
	FileSize     int64          `db:"file_size" json:"file_size"`
	StoragePath  string         `db:"storage_path" json:"storage_path"`
	CDNUrl       sql.NullString `db:"cdn_url" json:"cdn_url,omitempty"`
	Bucket       string         `db:"bucket" json:"bucket"`
	UploadType   string         `db:"upload_type" json:"upload_type"`
	EntityType   sql.NullString `db:"entity_type" json:"entity_type,omitempty"`
	EntityID     *uuid.UUID     `db:"entity_id" json:"entity_id,omitempty"`
	Status       string         `db:"status" json:"status"`
	Width        sql.NullInt32  `db:"width" json:"width,omitempty"`
	Height       sql.NullInt32  `db:"height" json:"height,omitempty"`
	DurationMs   sql.NullInt32  `db:"duration_ms" json:"duration_ms,omitempty"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

type MediaVariant struct {
	ID          uuid.UUID     `db:"id" json:"id"`
	UploadID    uuid.UUID     `db:"upload_id" json:"upload_id"`
	VariantType string        `db:"variant_type" json:"variant_type"`
	StoragePath string        `db:"storage_path" json:"storage_path"`
	CDNUrl      sql.NullString `db:"cdn_url" json:"cdn_url,omitempty"`
	Width       sql.NullInt32 `db:"width" json:"width,omitempty"`
	Height      sql.NullInt32 `db:"height" json:"height,omitempty"`
	FileSize    sql.NullInt64 `db:"file_size" json:"file_size,omitempty"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
}

type PresignedURL struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	FileName  string    `db:"file_name" json:"file_name"`
	MimeType  sql.NullString `db:"mime_type" json:"mime_type,omitempty"`
	UploadURL string    `db:"upload_url" json:"upload_url"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type RequestUploadURLRequest struct {
	FileName   string `json:"file_name" validate:"required"`
	MimeType   string `json:"mime_type" validate:"required"`
	UploadType string `json:"upload_type" validate:"required,oneof=AVATAR COVER POST_IMAGE POST_VIDEO PITCH_DECK DOCUMENT"`
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
}

type UploadURLResponse struct {
	UploadID  string `json:"upload_id"`
	UploadURL string `json:"upload_url"`
	ExpiresAt string `json:"expires_at"`
}

type ConfirmUploadRequest struct {
	UploadID string `json:"upload_id" validate:"required,uuid"`
	FileSize int64  `json:"file_size" validate:"required,gt=0"`
}

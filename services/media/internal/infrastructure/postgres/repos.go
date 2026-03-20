package postgres

import (
	"context"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/dealance/services/media/internal/domain/entity"
)

type MediaRepo struct{ db *sqlx.DB }
func NewMediaRepo(db *sqlx.DB) *MediaRepo { return &MediaRepo{db: db} }

func (r *MediaRepo) Create(ctx context.Context, m *entity.MediaUpload) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO media_uploads (id,user_id,file_name,original_name,mime_type,file_size,storage_path,bucket,upload_type,entity_type,entity_id,status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		m.ID, m.UserID, m.FileName, m.OriginalName, m.MimeType, m.FileSize, m.StoragePath, m.Bucket, m.UploadType, m.EntityType, m.EntityID, m.Status)
	return err
}
func (r *MediaRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.MediaUpload, error) {
	var m entity.MediaUpload; return &m, r.db.GetContext(ctx, &m, `SELECT * FROM media_uploads WHERE id=$1`, id)
}
func (r *MediaRepo) GetByUser(ctx context.Context, userID uuid.UUID, limit int) ([]entity.MediaUpload, error) {
	var ms []entity.MediaUpload
	return ms, r.db.SelectContext(ctx, &ms, `SELECT * FROM media_uploads WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, userID, limit)
}
func (r *MediaRepo) GetByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]entity.MediaUpload, error) {
	var ms []entity.MediaUpload
	return ms, r.db.SelectContext(ctx, &ms, `SELECT * FROM media_uploads WHERE entity_type=$1 AND entity_id=$2 ORDER BY created_at`, entityType, entityID)
}
func (r *MediaRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE media_uploads SET status=$1 WHERE id=$2`, status, id); return err
}
func (r *MediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM media_uploads WHERE id=$1`, id); return err
}

type VariantRepo struct{ db *sqlx.DB }
func NewVariantRepo(db *sqlx.DB) *VariantRepo { return &VariantRepo{db: db} }

func (r *VariantRepo) Create(ctx context.Context, v *entity.MediaVariant) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO media_variants (id,upload_id,variant_type,storage_path,width,height,file_size) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		v.ID, v.UploadID, v.VariantType, v.StoragePath, v.Width, v.Height, v.FileSize)
	return err
}
func (r *VariantRepo) GetByUpload(ctx context.Context, uploadID uuid.UUID) ([]entity.MediaVariant, error) {
	var vs []entity.MediaVariant
	return vs, r.db.SelectContext(ctx, &vs, `SELECT * FROM media_variants WHERE upload_id=$1`, uploadID)
}

type PresignedURLRepo struct{ db *sqlx.DB }
func NewPresignedURLRepo(db *sqlx.DB) *PresignedURLRepo { return &PresignedURLRepo{db: db} }

func (r *PresignedURLRepo) Create(ctx context.Context, p *entity.PresignedURL) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO presigned_urls (id,user_id,file_name,mime_type,upload_url,expires_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		p.ID, p.UserID, p.FileName, p.MimeType, p.UploadURL, p.ExpiresAt)
	return err
}
func (r *PresignedURLRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.PresignedURL, error) {
	var p entity.PresignedURL; return &p, r.db.GetContext(ctx, &p, `SELECT * FROM presigned_urls WHERE id=$1`, id)
}
func (r *PresignedURLRepo) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE presigned_urls SET used=true WHERE id=$1`, id); return err
}

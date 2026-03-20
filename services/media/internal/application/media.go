package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/media/internal/domain/entity"
	"github.com/dealance/services/media/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

type MediaService struct {
	mediaRepo    repository.MediaRepository
	presignRepo  repository.PresignedURLRepository
	log          zerolog.Logger
}

func NewMediaService(mediaRepo repository.MediaRepository, presignRepo repository.PresignedURLRepository, log zerolog.Logger) *MediaService {
	return &MediaService{mediaRepo: mediaRepo, presignRepo: presignRepo, log: log}
}

func (s *MediaService) RequestUploadURL(ctx context.Context, userID string, req entity.RequestUploadURLRequest) (*entity.UploadURLResponse, error) {
	uID, _ := uuid.Parse(userID)
	uploadID := uuid.New()
	fileName := fmt.Sprintf("%s/%s/%s_%s", req.UploadType, userID, uploadID.String()[:8], req.FileName)
	uploadURL := fmt.Sprintf("https://s3.ap-south-1.amazonaws.com/dealance-media/%s?X-Amz-Expires=3600", fileName)
	expiresAt := time.Now().Add(1 * time.Hour)

	presigned := &entity.PresignedURL{
		ID: uploadID, UserID: uID, FileName: fileName,
		MimeType: sql.NullString{String: req.MimeType, Valid: true},
		UploadURL: uploadURL, ExpiresAt: expiresAt,
	}
	if err := s.presignRepo.Create(ctx, presigned); err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }

	// Create media record
	media := &entity.MediaUpload{
		ID: uploadID, UserID: uID, FileName: fileName, OriginalName: req.FileName,
		MimeType: req.MimeType, StoragePath: fileName, Bucket: "dealance-media",
		UploadType: req.UploadType, Status: "PENDING",
	}
	if req.EntityType != "" { media.EntityType = sql.NullString{String: req.EntityType, Valid: true} }
	if req.EntityID != "" { eID, _ := uuid.Parse(req.EntityID); media.EntityID = &eID }
	_ = s.mediaRepo.Create(ctx, media)

	return &entity.UploadURLResponse{UploadID: uploadID.String(), UploadURL: uploadURL, ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z")}, nil
}

func (s *MediaService) ConfirmUpload(ctx context.Context, userID string, req entity.ConfirmUploadRequest) error {
	uID, _ := uuid.Parse(userID)
	mID, _ := uuid.Parse(req.UploadID)
	media, err := s.mediaRepo.GetByID(ctx, mID)
	if err != nil { return apperrors.ErrNotFound("Upload") }
	if media.UserID != uID { return apperrors.ErrForbidden("NOT_OWNER", "Not your upload") }
	return s.mediaRepo.UpdateStatus(ctx, mID, "READY")
}

func (s *MediaService) GetMyUploads(ctx context.Context, userID string, limit int) ([]entity.MediaUpload, error) {
	uID, _ := uuid.Parse(userID)
	if limit <= 0 || limit > 50 { limit = 20 }
	return s.mediaRepo.GetByUser(ctx, uID, limit)
}

func (s *MediaService) DeleteUpload(ctx context.Context, userID, uploadID string) error {
	uID, _ := uuid.Parse(userID)
	mID, _ := uuid.Parse(uploadID)
	media, err := s.mediaRepo.GetByID(ctx, mID)
	if err != nil { return apperrors.ErrNotFound("Upload") }
	if media.UserID != uID { return apperrors.ErrForbidden("NOT_OWNER", "Not your upload") }
	return s.mediaRepo.Delete(ctx, mID)
}

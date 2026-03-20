package application

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/admin/internal/domain/entity"
	"github.com/dealance/services/admin/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

type AdminService struct {
	adminRepo repository.AdminUserRepository
	auditRepo repository.AuditLogRepository
	statsRepo repository.StatsRepository
	log       zerolog.Logger
}

func NewAdminService(adminRepo repository.AdminUserRepository, auditRepo repository.AuditLogRepository, statsRepo repository.StatsRepository, log zerolog.Logger) *AdminService {
	return &AdminService{adminRepo: adminRepo, auditRepo: auditRepo, statsRepo: statsRepo, log: log}
}

func (s *AdminService) GetDashboardStats(ctx context.Context, adminUserID string) (*entity.PlatformStats, error) {
	uID, _ := uuid.Parse(adminUserID)
	isAdmin, _ := s.adminRepo.IsAdmin(ctx, uID)
	if !isAdmin { return nil, apperrors.ErrForbidden("NOT_ADMIN", "Admin access required") }
	stats, err := s.statsRepo.GetLatest(ctx)
	if err != nil { return nil, apperrors.ErrNotFound("Stats") }
	return stats, nil
}

func (s *AdminService) ModerateContent(ctx context.Context, adminUserID string, req entity.ContentModerationRequest, ip string) error {
	uID, _ := uuid.Parse(adminUserID)
	admin, err := s.adminRepo.GetByUserID(ctx, uID)
	if err != nil { return apperrors.ErrForbidden("NOT_ADMIN", "Admin access required") }

	eID, _ := uuid.Parse(req.EntityID)
	auditEntry := &entity.AdminAuditLog{
		ID: uuid.New(), AdminID: admin.ID, Action: req.Action + "_" + req.EntityType,
		EntityType: sql.NullString{String: req.EntityType, Valid: true}, EntityID: &eID,
		IPAddress: sql.NullString{String: ip, Valid: true},
	}
	return s.auditRepo.Create(ctx, auditEntry)
}

func (s *AdminService) GetAuditLog(ctx context.Context, adminUserID string, limit int) ([]entity.AdminAuditLog, error) {
	uID, _ := uuid.Parse(adminUserID)
	isAdmin, _ := s.adminRepo.IsAdmin(ctx, uID)
	if !isAdmin { return nil, apperrors.ErrForbidden("NOT_ADMIN", "Admin access required") }
	if limit <= 0 || limit > 100 { limit = 50 }
	return s.auditRepo.GetRecent(ctx, limit)
}

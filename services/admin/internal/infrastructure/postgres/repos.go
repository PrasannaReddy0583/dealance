package postgres

import (
	"context"
	"time"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/dealance/services/admin/internal/domain/entity"
)

type AdminUserRepo struct{ db *sqlx.DB }
func NewAdminUserRepo(db *sqlx.DB) *AdminUserRepo { return &AdminUserRepo{db: db} }

func (r *AdminUserRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.AdminUser, error) {
	var a entity.AdminUser; return &a, r.db.GetContext(ctx, &a, `SELECT * FROM admin_users WHERE user_id=$1`, userID)
}
func (r *AdminUserRepo) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool; return exists, r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM admin_users WHERE user_id=$1)`, userID)
}

type AuditLogRepo struct{ db *sqlx.DB }
func NewAuditLogRepo(db *sqlx.DB) *AuditLogRepo { return &AuditLogRepo{db: db} }

func (r *AuditLogRepo) Create(ctx context.Context, l *entity.AdminAuditLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO admin_audit_log (id,admin_id,action,entity_type,entity_id,ip_address) VALUES ($1,$2,$3,$4,$5,$6)`,
		l.ID, l.AdminID, l.Action, l.EntityType, l.EntityID, l.IPAddress)
	return err
}
func (r *AuditLogRepo) GetRecent(ctx context.Context, limit int) ([]entity.AdminAuditLog, error) {
	var ls []entity.AdminAuditLog
	return ls, r.db.SelectContext(ctx, &ls, `SELECT * FROM admin_audit_log ORDER BY created_at DESC LIMIT $1`, limit)
}
func (r *AuditLogRepo) GetByAdmin(ctx context.Context, adminID uuid.UUID, limit int) ([]entity.AdminAuditLog, error) {
	var ls []entity.AdminAuditLog
	return ls, r.db.SelectContext(ctx, &ls, `SELECT * FROM admin_audit_log WHERE admin_id=$1 ORDER BY created_at DESC LIMIT $2`, adminID, limit)
}

type StatsRepo struct{ db *sqlx.DB }
func NewStatsRepo(db *sqlx.DB) *StatsRepo { return &StatsRepo{db: db} }

func (r *StatsRepo) GetLatest(ctx context.Context) (*entity.PlatformStats, error) {
	var s entity.PlatformStats; return &s, r.db.GetContext(ctx, &s, `SELECT * FROM platform_stats ORDER BY stat_date DESC LIMIT 1`)
}
func (r *StatsRepo) GetByDateRange(ctx context.Context, from, to time.Time) ([]entity.PlatformStats, error) {
	var ss []entity.PlatformStats
	return ss, r.db.SelectContext(ctx, &ss, `SELECT * FROM platform_stats WHERE stat_date BETWEEN $1 AND $2 ORDER BY stat_date`, from, to)
}
func (r *StatsRepo) Upsert(ctx context.Context, s *entity.PlatformStats) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO platform_stats (id,stat_date,total_users,active_users,total_startups,total_deals,total_invested_paise,total_posts,new_users_today)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT (stat_date) DO UPDATE SET total_users=$3,active_users=$4,total_startups=$5,total_deals=$6,total_invested_paise=$7,total_posts=$8,new_users_today=$9`,
		s.ID, s.StatDate, s.TotalUsers, s.ActiveUsers, s.TotalStartups, s.TotalDeals, s.TotalInvestedPaise, s.TotalPosts, s.NewUsersToday)
	return err
}

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dealance/services/startup/internal/domain/entity"
)

// --- StartupRepo ---

type StartupRepo struct{ db *sqlx.DB }

func NewStartupRepo(db *sqlx.DB) *StartupRepo { return &StartupRepo{db: db} }

func (r *StartupRepo) Create(ctx context.Context, s *entity.Startup) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO startups (id, founder_id, name, slug, tagline, description, sector, stage, business_model, website, founded_year, headquarters, country, incorporation_type, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		s.ID, s.FounderID, s.Name, s.Slug, s.Tagline, s.Description, s.Sector, s.Stage, s.BusinessModel,
		s.Website, s.FoundedYear, s.Headquarters, s.Country, s.IncorporationType, s.Tags)
	return err
}

func (r *StartupRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Startup, error) {
	var s entity.Startup
	return &s, r.db.GetContext(ctx, &s, `SELECT * FROM startups WHERE id = $1`, id)
}

func (r *StartupRepo) GetBySlug(ctx context.Context, slug string) (*entity.Startup, error) {
	var s entity.Startup
	return &s, r.db.GetContext(ctx, &s, `SELECT * FROM startups WHERE slug = $1`, slug)
}

func (r *StartupRepo) GetByFounder(ctx context.Context, founderID uuid.UUID) ([]entity.Startup, error) {
	var startups []entity.Startup
	return startups, r.db.SelectContext(ctx, &startups,
		`SELECT * FROM startups WHERE founder_id = $1 ORDER BY created_at DESC`, founderID)
}

func (r *StartupRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	sets, args := buildUpdateArgs(updates)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE startups SET %s WHERE id = $%d`, sets, len(args)), args...)
	return err
}

func (r *StartupRepo) Search(ctx context.Context, query, sector, stage, country string, limit int, cursorTime *time.Time) ([]entity.Startup, error) {
	var startups []entity.Startup
	conditions := []string{"status = 'ACTIVE'"}
	args := []interface{}{}
	i := 1
	if query != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR tagline ILIKE $%d OR description ILIKE $%d)", i, i, i))
		args = append(args, "%"+query+"%")
		i++
	}
	if sector != "" {
		conditions = append(conditions, fmt.Sprintf("sector = $%d", i))
		args = append(args, sector)
		i++
	}
	if stage != "" {
		conditions = append(conditions, fmt.Sprintf("stage = $%d", i))
		args = append(args, stage)
		i++
	}
	if country != "" {
		conditions = append(conditions, fmt.Sprintf("country = $%d", i))
		args = append(args, country)
		i++
	}
	if cursorTime != nil {
		conditions = append(conditions, fmt.Sprintf("created_at < $%d", i))
		args = append(args, cursorTime)
		i++
	}
	args = append(args, limit)
	q := fmt.Sprintf(`SELECT * FROM startups WHERE %s ORDER BY created_at DESC LIMIT $%d`,
		strings.Join(conditions, " AND "), i)
	return startups, r.db.SelectContext(ctx, &startups, q, args...)
}

func (r *StartupRepo) IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error {
	q := fmt.Sprintf(`UPDATE startups SET %s = %s + $1 WHERE id = $2`, column, column)
	_, err := r.db.ExecContext(ctx, q, delta, id)
	return err
}

// --- FundingRoundRepo ---

type FundingRoundRepo struct{ db *sqlx.DB }

func NewFundingRoundRepo(db *sqlx.DB) *FundingRoundRepo { return &FundingRoundRepo{db: db} }

func (r *FundingRoundRepo) Create(ctx context.Context, fr *entity.FundingRound) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO funding_rounds (id, startup_id, round_type, amount_paise, valuation_paise, target_paise, min_ticket_paise, equity_offered, instrument_type, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		fr.ID, fr.StartupID, fr.RoundType, fr.AmountPaise, fr.ValuationPaise, fr.TargetPaise,
		fr.MinTicketPaise, fr.EquityOffered, fr.InstrumentType, fr.Description)
	return err
}

func (r *FundingRoundRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.FundingRound, error) {
	var fr entity.FundingRound
	return &fr, r.db.GetContext(ctx, &fr, `SELECT * FROM funding_rounds WHERE id = $1`, id)
}

func (r *FundingRoundRepo) GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.FundingRound, error) {
	var rounds []entity.FundingRound
	return rounds, r.db.SelectContext(ctx, &rounds,
		`SELECT * FROM funding_rounds WHERE startup_id = $1 ORDER BY created_at DESC`, startupID)
}

func (r *FundingRoundRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	sets, args := buildUpdateArgs(updates)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE funding_rounds SET %s WHERE id = $%d`, sets, len(args)), args...)
	return err
}

// --- TeamMemberRepo ---

type TeamMemberRepo struct{ db *sqlx.DB }

func NewTeamMemberRepo(db *sqlx.DB) *TeamMemberRepo { return &TeamMemberRepo{db: db} }

func (r *TeamMemberRepo) Create(ctx context.Context, m *entity.TeamMember) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO team_members (id, startup_id, user_id, name, role, title, bio, linkedin_url, is_founder, display_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		m.ID, m.StartupID, m.UserID, m.Name, m.Role, m.Title, m.Bio, m.LinkedInURL, m.IsFounder, m.DisplayOrder)
	return err
}

func (r *TeamMemberRepo) GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.TeamMember, error) {
	var members []entity.TeamMember
	return members, r.db.SelectContext(ctx, &members,
		`SELECT * FROM team_members WHERE startup_id = $1 ORDER BY display_order, is_founder DESC`, startupID)
}

func (r *TeamMemberRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM team_members WHERE id = $1`, id)
	return err
}

// --- StartupMediaRepo ---

type StartupMediaRepo struct{ db *sqlx.DB }

func NewStartupMediaRepo(db *sqlx.DB) *StartupMediaRepo { return &StartupMediaRepo{db: db} }

func (r *StartupMediaRepo) Create(ctx context.Context, m *entity.StartupMedia) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO startup_media (id, startup_id, media_type, title, media_url, thumbnail_url, file_size, is_public, display_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.StartupID, m.MediaType, m.Title, m.MediaURL, m.ThumbnailURL, m.FileSize, m.IsPublic, m.DisplayOrder)
	return err
}

func (r *StartupMediaRepo) GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.StartupMedia, error) {
	var media []entity.StartupMedia
	return media, r.db.SelectContext(ctx, &media,
		`SELECT * FROM startup_media WHERE startup_id = $1 ORDER BY display_order`, startupID)
}

func (r *StartupMediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM startup_media WHERE id = $1`, id)
	return err
}

// --- StartupMetricRepo ---

type StartupMetricRepo struct{ db *sqlx.DB }

func NewStartupMetricRepo(db *sqlx.DB) *StartupMetricRepo { return &StartupMetricRepo{db: db} }

func (r *StartupMetricRepo) Create(ctx context.Context, m *entity.StartupMetric) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO startup_metrics (id, startup_id, metric_type, value, currency, period)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		m.ID, m.StartupID, m.MetricType, m.Value, m.Currency, m.Period)
	return err
}

func (r *StartupMetricRepo) GetByStartup(ctx context.Context, startupID uuid.UUID, metricType string) ([]entity.StartupMetric, error) {
	var metrics []entity.StartupMetric
	if metricType != "" {
		return metrics, r.db.SelectContext(ctx, &metrics,
			`SELECT * FROM startup_metrics WHERE startup_id = $1 AND metric_type = $2 ORDER BY recorded_at DESC`, startupID, metricType)
	}
	return metrics, r.db.SelectContext(ctx, &metrics,
		`SELECT * FROM startup_metrics WHERE startup_id = $1 ORDER BY recorded_at DESC`, startupID)
}

func (r *StartupMetricRepo) GetLatest(ctx context.Context, startupID uuid.UUID) ([]entity.StartupMetric, error) {
	var metrics []entity.StartupMetric
	return metrics, r.db.SelectContext(ctx, &metrics,
		`SELECT DISTINCT ON (metric_type) * FROM startup_metrics WHERE startup_id = $1 ORDER BY metric_type, recorded_at DESC`, startupID)
}

// --- StartupFollowRepo ---

type StartupFollowRepo struct{ db *sqlx.DB }

func NewStartupFollowRepo(db *sqlx.DB) *StartupFollowRepo { return &StartupFollowRepo{db: db} }

func (r *StartupFollowRepo) Follow(ctx context.Context, userID, startupID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO startup_follows (user_id, startup_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, startupID)
	return err
}

func (r *StartupFollowRepo) Unfollow(ctx context.Context, userID, startupID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM startup_follows WHERE user_id = $1 AND startup_id = $2`, userID, startupID)
	return err
}

func (r *StartupFollowRepo) IsFollowing(ctx context.Context, userID, startupID uuid.UUID) (bool, error) {
	var exists bool
	return exists, r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM startup_follows WHERE user_id = $1 AND startup_id = $2)`, userID, startupID)
}

func (r *StartupFollowRepo) GetFollowers(ctx context.Context, startupID uuid.UUID, limit int) ([]entity.StartupFollow, error) {
	var follows []entity.StartupFollow
	return follows, r.db.SelectContext(ctx, &follows,
		`SELECT * FROM startup_follows WHERE startup_id = $1 ORDER BY created_at DESC LIMIT $2`, startupID, limit)
}

// --- helpers ---

func buildUpdateArgs(updates map[string]interface{}) (string, []interface{}) {
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	return strings.Join(setClauses, ", "), args
}

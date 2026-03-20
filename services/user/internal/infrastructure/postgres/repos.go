package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dealance/services/user/internal/domain/entity"
)

// --- ProfileRepo ---

type ProfileRepo struct{ db *sqlx.DB }

func NewProfileRepo(db *sqlx.DB) *ProfileRepo { return &ProfileRepo{db: db} }

func (r *ProfileRepo) Create(ctx context.Context, p *entity.Profile) error {
	query := `INSERT INTO profiles (id, username, display_name, bio, avatar_url, cover_url, location, website, linkedin_url, twitter_url, profession, company, experience_years, is_public)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Username, p.DisplayName, p.Bio, p.AvatarURL, p.CoverURL,
		p.Location, p.Website, p.LinkedInURL, p.TwitterURL,
		p.Profession, p.Company, p.ExperienceYears, p.IsPublic,
	)
	return err
}

func (r *ProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Profile, error) {
	var p entity.Profile
	err := r.db.GetContext(ctx, &p, `SELECT * FROM profiles WHERE id = $1`, id)
	return &p, err
}

func (r *ProfileRepo) GetByUsername(ctx context.Context, username string) (*entity.Profile, error) {
	var p entity.Profile
	err := r.db.GetContext(ctx, &p, `SELECT * FROM profiles WHERE username = $1`, username)
	return &p, err
}

func (r *ProfileRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM profiles WHERE username = $1)`, username)
	return exists, err
}

func (r *ProfileRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf(`UPDATE profiles SET %s WHERE id = $%d`, strings.Join(setClauses, ", "), i)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *ProfileRepo) IncrementCounter(ctx context.Context, id uuid.UUID, column string, delta int) error {
	query := fmt.Sprintf(`UPDATE profiles SET %s = %s + $1 WHERE id = $2`, column, column)
	_, err := r.db.ExecContext(ctx, query, delta, id)
	return err
}

func (r *ProfileRepo) Search(ctx context.Context, query string, limit int, cursorTime *time.Time, cursorID *string) ([]entity.Profile, error) {
	searchPattern := "%" + query + "%"
	var profiles []entity.Profile
	var err error

	if cursorTime != nil && cursorID != nil {
		err = r.db.SelectContext(ctx, &profiles,
			`SELECT * FROM profiles
			WHERE (display_name ILIKE $1 OR username ILIKE $1 OR profession ILIKE $1)
			AND is_public = true
			AND (created_at, id::text) < ($3, $4)
			ORDER BY created_at DESC, id DESC
			LIMIT $2`,
			searchPattern, limit, cursorTime, cursorID,
		)
	} else {
		err = r.db.SelectContext(ctx, &profiles,
			`SELECT * FROM profiles
			WHERE (display_name ILIKE $1 OR username ILIKE $1 OR profession ILIKE $1)
			AND is_public = true
			ORDER BY created_at DESC, id DESC
			LIMIT $2`,
			searchPattern, limit,
		)
	}
	return profiles, err
}

// --- FollowRepo ---

type FollowRepo struct{ db *sqlx.DB }

func NewFollowRepo(db *sqlx.DB) *FollowRepo { return &FollowRepo{db: db} }

func (r *FollowRepo) Follow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO follows (follower_id, following_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		followerID, followingID,
	)
	return err
}

func (r *FollowRepo) Unfollow(ctx context.Context, followerID, followingID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`,
		followerID, followingID,
	)
	return err
}

func (r *FollowRepo) IsFollowing(ctx context.Context, followerID, followingID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2)`,
		followerID, followingID,
	)
	return exists, err
}

func (r *FollowRepo) GetFollowers(ctx context.Context, userID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Follow, error) {
	var follows []entity.Follow
	var err error
	if cursorTime != nil {
		err = r.db.SelectContext(ctx, &follows,
			`SELECT * FROM follows WHERE following_id = $1 AND created_at < $3 ORDER BY created_at DESC LIMIT $2`,
			userID, limit, cursorTime,
		)
	} else {
		err = r.db.SelectContext(ctx, &follows,
			`SELECT * FROM follows WHERE following_id = $1 ORDER BY created_at DESC LIMIT $2`,
			userID, limit,
		)
	}
	return follows, err
}

func (r *FollowRepo) GetFollowing(ctx context.Context, userID uuid.UUID, limit int, cursorTime *time.Time) ([]entity.Follow, error) {
	var follows []entity.Follow
	var err error
	if cursorTime != nil {
		err = r.db.SelectContext(ctx, &follows,
			`SELECT * FROM follows WHERE follower_id = $1 AND created_at < $3 ORDER BY created_at DESC LIMIT $2`,
			userID, limit, cursorTime,
		)
	} else {
		err = r.db.SelectContext(ctx, &follows,
			`SELECT * FROM follows WHERE follower_id = $1 ORDER BY created_at DESC LIMIT $2`,
			userID, limit,
		)
	}
	return follows, err
}

func (r *FollowRepo) GetFollowerCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM follows WHERE following_id = $1`, userID)
	return count, err
}

func (r *FollowRepo) GetFollowingCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM follows WHERE follower_id = $1`, userID)
	return count, err
}

func (r *FollowRepo) GetMutualFollowers(ctx context.Context, userA, userB uuid.UUID, limit int) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.SelectContext(ctx, &ids,
		`SELECT f1.follower_id FROM follows f1
		INNER JOIN follows f2 ON f1.follower_id = f2.follower_id
		WHERE f1.following_id = $1 AND f2.following_id = $2
		LIMIT $3`,
		userA, userB, limit,
	)
	return ids, err
}

// --- BlockRepo ---

type BlockRepo struct{ db *sqlx.DB }

func NewBlockRepo(db *sqlx.DB) *BlockRepo { return &BlockRepo{db: db} }

func (r *BlockRepo) Block(ctx context.Context, blockerID, blockedID uuid.UUID, reason string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO blocked_users (blocker_id, blocked_id, reason) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		blockerID, blockedID, reason,
	)
	return err
}

func (r *BlockRepo) Unblock(ctx context.Context, blockerID, blockedID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM blocked_users WHERE blocker_id = $1 AND blocked_id = $2`,
		blockerID, blockedID,
	)
	return err
}

func (r *BlockRepo) IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists,
		`SELECT EXISTS(SELECT 1 FROM blocked_users WHERE blocker_id = $1 AND blocked_id = $2)`,
		blockerID, blockedID,
	)
	return exists, err
}

func (r *BlockRepo) GetBlockedUsers(ctx context.Context, blockerID uuid.UUID) ([]entity.BlockedUser, error) {
	var blocked []entity.BlockedUser
	err := r.db.SelectContext(ctx, &blocked,
		`SELECT * FROM blocked_users WHERE blocker_id = $1 ORDER BY created_at DESC`, blockerID,
	)
	return blocked, err
}

// --- SettingsRepo ---

type SettingsRepo struct{ db *sqlx.DB }

func NewSettingsRepo(db *sqlx.DB) *SettingsRepo { return &SettingsRepo{db: db} }

func (r *SettingsRepo) Create(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_settings (user_id) VALUES ($1) ON CONFLICT DO NOTHING`, userID,
	)
	return err
}

func (r *SettingsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.UserSettings, error) {
	var s entity.UserSettings
	err := r.db.GetContext(ctx, &s, `SELECT * FROM user_settings WHERE user_id = $1`, userID)
	return &s, err
}

func (r *SettingsRepo) Update(ctx context.Context, userID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	args = append(args, userID)
	query := fmt.Sprintf(`UPDATE user_settings SET %s WHERE user_id = $%d`, strings.Join(setClauses, ", "), i)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// --- ProfileMediaRepo ---

type ProfileMediaRepo struct{ db *sqlx.DB }

func NewProfileMediaRepo(db *sqlx.DB) *ProfileMediaRepo { return &ProfileMediaRepo{db: db} }

func (r *ProfileMediaRepo) Create(ctx context.Context, m *entity.ProfileMedia) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO profile_media (id, user_id, media_type, title, description, media_url, thumbnail_url, display_order, is_visible)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.UserID, m.MediaType, m.Title, m.Description, m.MediaURL, m.ThumbnailURL, m.DisplayOrder, m.IsVisible,
	)
	return err
}

func (r *ProfileMediaRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.ProfileMedia, error) {
	var m entity.ProfileMedia
	err := r.db.GetContext(ctx, &m, `SELECT * FROM profile_media WHERE id = $1`, id)
	return &m, err
}

func (r *ProfileMediaRepo) GetByUserID(ctx context.Context, userID uuid.UUID, mediaType string) ([]entity.ProfileMedia, error) {
	var items []entity.ProfileMedia
	var err error
	if mediaType != "" {
		err = r.db.SelectContext(ctx, &items,
			`SELECT * FROM profile_media WHERE user_id = $1 AND media_type = $2 ORDER BY display_order, created_at DESC`, userID, mediaType)
	} else {
		err = r.db.SelectContext(ctx, &items,
			`SELECT * FROM profile_media WHERE user_id = $1 ORDER BY display_order, created_at DESC`, userID)
	}
	return items, err
}

func (r *ProfileMediaRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	setClauses := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	args = append(args, id)
	query := fmt.Sprintf(`UPDATE profile_media SET %s WHERE id = $%d`, strings.Join(setClauses, ", "), i)
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *ProfileMediaRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM profile_media WHERE id = $1`, id)
	return err
}

// --- EntrepreneurProfileRepo ---

type EntrepreneurProfileRepo struct{ db *sqlx.DB }

func NewEntrepreneurProfileRepo(db *sqlx.DB) *EntrepreneurProfileRepo {
	return &EntrepreneurProfileRepo{db: db}
}

func (r *EntrepreneurProfileRepo) Upsert(ctx context.Context, ep *entity.EntrepreneurProfile) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO entrepreneur_profiles (user_id, sectors, skills, education, work_history)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (user_id) DO UPDATE SET sectors=$2, skills=$3, education=$4, work_history=$5`,
		ep.UserID, ep.Sectors, ep.Skills, ep.Education, ep.WorkHistory,
	)
	return err
}

func (r *EntrepreneurProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.EntrepreneurProfile, error) {
	var ep entity.EntrepreneurProfile
	err := r.db.GetContext(ctx, &ep, `SELECT * FROM entrepreneur_profiles WHERE user_id = $1`, userID)
	return &ep, err
}

// --- InvestorProfileRepo ---

type InvestorProfileRepo struct{ db *sqlx.DB }

func NewInvestorProfileRepo(db *sqlx.DB) *InvestorProfileRepo {
	return &InvestorProfileRepo{db: db}
}

func (r *InvestorProfileRepo) Upsert(ctx context.Context, ip *entity.InvestorProfile) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO investor_profiles (user_id, investor_type, investment_range_min_paise, investment_range_max_paise, preferred_sectors, preferred_stages, investment_thesis)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (user_id) DO UPDATE SET investor_type=$2, investment_range_min_paise=$3, investment_range_max_paise=$4, preferred_sectors=$5, preferred_stages=$6, investment_thesis=$7`,
		ip.UserID, ip.InvestorType, ip.InvestmentRangeMinPaise, ip.InvestmentRangeMaxPaise,
		ip.PreferredSectors, ip.PreferredStages, ip.InvestmentThesis,
	)
	return err
}

func (r *InvestorProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.InvestorProfile, error) {
	var ip entity.InvestorProfile
	err := r.db.GetContext(ctx, &ip, `SELECT * FROM investor_profiles WHERE user_id = $1`, userID)
	return &ip, err
}

package entity

import (
	"database/sql"
	"time"
	"github.com/google/uuid"
)

type AdminUser struct {
	ID          uuid.UUID `db:"id" json:"id"`
	UserID      uuid.UUID `db:"user_id" json:"user_id"`
	Role        string    `db:"role" json:"role"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type AdminAuditLog struct {
	ID         uuid.UUID      `db:"id" json:"id"`
	AdminID    uuid.UUID      `db:"admin_id" json:"admin_id"`
	Action     string         `db:"action" json:"action"`
	EntityType sql.NullString `db:"entity_type" json:"entity_type,omitempty"`
	EntityID   *uuid.UUID     `db:"entity_id" json:"entity_id,omitempty"`
	IPAddress  sql.NullString `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}

type PlatformStats struct {
	ID                uuid.UUID `db:"id" json:"id"`
	StatDate          time.Time `db:"stat_date" json:"stat_date"`
	TotalUsers        int       `db:"total_users" json:"total_users"`
	ActiveUsers       int       `db:"active_users" json:"active_users"`
	TotalStartups     int       `db:"total_startups" json:"total_startups"`
	TotalDeals        int       `db:"total_deals" json:"total_deals"`
	TotalInvestedPaise int64   `db:"total_invested_paise" json:"total_invested_paise"`
	TotalPosts        int       `db:"total_posts" json:"total_posts"`
	NewUsersToday     int       `db:"new_users_today" json:"new_users_today"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
}

type BanUserRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Reason string `json:"reason" validate:"required"`
}

type ContentModerationRequest struct {
	EntityType string `json:"entity_type" validate:"required,oneof=POST COMMENT USER STARTUP"`
	EntityID   string `json:"entity_id" validate:"required,uuid"`
	Action     string `json:"action" validate:"required,oneof=HIDE REMOVE WARN BAN"`
	Reason     string `json:"reason" validate:"required"`
}

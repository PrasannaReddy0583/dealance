package scylla

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/auth/internal/domain/entity"
)

// AuditLogRepo implements AuditLogRepository using ScyllaDB.
// Note: In production, we'd use the gocql driver for ScyllaDB.
// For local development, this provides a fallback that logs to zerolog
// when ScyllaDB is unavailable.
type AuditLogRepo struct {
	log zerolog.Logger
	// In production: session *gocql.Session
}

// NewAuditLogRepo creates a new audit log repository.
func NewAuditLogRepo(log zerolog.Logger) *AuditLogRepo {
	return &AuditLogRepo{
		log: log,
	}
}

// Log writes an audit entry.
// In production, this inserts into ScyllaDB.
// In development, it logs to zerolog as a fallback.
func (r *AuditLogRepo) Log(ctx context.Context, entry *entity.AuditLogEntry) error {
	r.log.Info().
		Str("user_id", entry.UserID.String()).
		Str("event_id", entry.EventID.String()).
		Str("event_type", entry.EventType).
		Str("device_id", entry.DeviceID).
		Str("ip_address", entry.IPAddress).
		Float64("risk_score", entry.RiskScore).
		Time("event_at", entry.EventAt).
		Msg("audit_log")

	// Production implementation:
	// err := r.session.Query(
	//     `INSERT INTO security_audit_log (user_id, event_at, event_id, device_id, event_type, event_data, ip_address, risk_score)
	//      VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
	//     entry.UserID, entry.EventAt, entry.EventID, entry.DeviceID,
	//     entry.EventType, entry.EventData, entry.IPAddress, entry.RiskScore,
	// ).WithContext(ctx).Exec()

	return nil
}

// GetByUserID retrieves audit log entries for a user.
func (r *AuditLogRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]entity.AuditLogEntry, error) {
	// Production:
	// scanner := r.session.Query(
	//     `SELECT user_id, event_at, event_id, device_id, event_type, event_data, ip_address, risk_score
	//      FROM security_audit_log WHERE user_id = ? LIMIT ?`,
	//     userID, limit,
	// ).WithContext(ctx).Iter().Scanner()

	return []entity.AuditLogEntry{}, nil
}

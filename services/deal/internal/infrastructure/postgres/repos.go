package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/dealance/services/deal/internal/domain/entity"
)

func buildUpdateArgs(updates map[string]interface{}) (string, []interface{}) {
	sets := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	i := 1
	for col, val := range updates {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	return strings.Join(sets, ", "), args
}

// --- DealRepo ---
type DealRepo struct{ db *sqlx.DB }
func NewDealRepo(db *sqlx.DB) *DealRepo { return &DealRepo{db: db} }

func (r *DealRepo) Create(ctx context.Context, d *entity.Deal) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO deals (id, startup_id, funding_round_id, title, description, deal_type, status, amount_paise, min_ticket_paise, max_participants, equity_pct, valuation_paise, terms_summary, requires_nda, requires_kyc, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		d.ID, d.StartupID, d.FundingRoundID, d.Title, d.Description, d.DealType, d.Status,
		d.AmountPaise, d.MinTicketPaise, d.MaxParticipants, d.EquityPct, d.ValuationPaise,
		d.TermsSummary, d.RequiresNDA, d.RequiresKYC, d.CreatedBy)
	return err
}
func (r *DealRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Deal, error) {
	var d entity.Deal; return &d, r.db.GetContext(ctx, &d, `SELECT * FROM deals WHERE id = $1`, id)
}
func (r *DealRepo) GetByStartup(ctx context.Context, startupID uuid.UUID) ([]entity.Deal, error) {
	var deals []entity.Deal
	return deals, r.db.SelectContext(ctx, &deals, `SELECT * FROM deals WHERE startup_id = $1 ORDER BY created_at DESC`, startupID)
}
func (r *DealRepo) GetByCreator(ctx context.Context, creatorID uuid.UUID) ([]entity.Deal, error) {
	var deals []entity.Deal
	return deals, r.db.SelectContext(ctx, &deals, `SELECT * FROM deals WHERE created_by = $1 ORDER BY created_at DESC`, creatorID)
}
func (r *DealRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	sets, args := buildUpdateArgs(updates)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE deals SET %s WHERE id = $%d`, sets, len(args)), args...)
	return err
}

// --- ParticipantRepo ---
type ParticipantRepo struct{ db *sqlx.DB }
func NewParticipantRepo(db *sqlx.DB) *ParticipantRepo { return &ParticipantRepo{db: db} }

func (r *ParticipantRepo) Create(ctx context.Context, p *entity.DealParticipant) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO deal_participants (id, deal_id, user_id, role, status) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (deal_id, user_id) DO NOTHING`,
		p.ID, p.DealID, p.UserID, p.Role, p.Status)
	return err
}
func (r *ParticipantRepo) GetByDealAndUser(ctx context.Context, dealID, userID uuid.UUID) (*entity.DealParticipant, error) {
	var p entity.DealParticipant
	return &p, r.db.GetContext(ctx, &p, `SELECT * FROM deal_participants WHERE deal_id = $1 AND user_id = $2`, dealID, userID)
}
func (r *ParticipantRepo) GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealParticipant, error) {
	var ps []entity.DealParticipant
	return ps, r.db.SelectContext(ctx, &ps, `SELECT * FROM deal_participants WHERE deal_id = $1 ORDER BY created_at`, dealID)
}
func (r *ParticipantRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.DealParticipant, error) {
	var ps []entity.DealParticipant
	return ps, r.db.SelectContext(ctx, &ps, `SELECT * FROM deal_participants WHERE user_id = $1 ORDER BY created_at DESC`, userID)
}
func (r *ParticipantRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	sets, args := buildUpdateArgs(updates)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE deal_participants SET %s WHERE id = $%d`, sets, len(args)), args...)
	return err
}

// --- DocumentRepo ---
type DocumentRepo struct{ db *sqlx.DB }
func NewDocumentRepo(db *sqlx.DB) *DocumentRepo { return &DocumentRepo{db: db} }

func (r *DocumentRepo) Create(ctx context.Context, d *entity.DealDocument) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO deal_documents (id, deal_id, uploaded_by, doc_type, title, file_url, file_size, mime_type, is_confidential, access_level)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		d.ID, d.DealID, d.UploadedBy, d.DocType, d.Title, d.FileURL, d.FileSize, d.MimeType, d.IsConfidential, d.AccessLevel)
	return err
}
func (r *DocumentRepo) GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealDocument, error) {
	var docs []entity.DealDocument
	return docs, r.db.SelectContext(ctx, &docs, `SELECT * FROM deal_documents WHERE deal_id = $1 ORDER BY created_at DESC`, dealID)
}
func (r *DocumentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM deal_documents WHERE id = $1`, id); return err
}

// --- MilestoneRepo ---
type MilestoneRepo struct{ db *sqlx.DB }
func NewMilestoneRepo(db *sqlx.DB) *MilestoneRepo { return &MilestoneRepo{db: db} }

func (r *MilestoneRepo) Create(ctx context.Context, m *entity.DealMilestone) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO deal_milestones (id, deal_id, title, description, milestone_type, status, due_date, display_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.ID, m.DealID, m.Title, m.Description, m.MilestoneType, m.Status, m.DueDate, m.DisplayOrder)
	return err
}
func (r *MilestoneRepo) GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealMilestone, error) {
	var ms []entity.DealMilestone
	return ms, r.db.SelectContext(ctx, &ms, `SELECT * FROM deal_milestones WHERE deal_id = $1 ORDER BY display_order`, dealID)
}
func (r *MilestoneRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	sets, args := buildUpdateArgs(updates)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE deal_milestones SET %s WHERE id = $%d`, sets, len(args)), args...)
	return err
}

// --- NDARepo ---
type NDARepo struct{ db *sqlx.DB }
func NewNDARepo(db *sqlx.DB) *NDARepo { return &NDARepo{db: db} }

func (r *NDARepo) Create(ctx context.Context, n *entity.DealNDA) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO deal_ndas (id, deal_id, user_id, status) VALUES ($1,$2,$3,$4) ON CONFLICT (deal_id, user_id) DO NOTHING`,
		n.ID, n.DealID, n.UserID, n.Status)
	return err
}
func (r *NDARepo) GetByDealAndUser(ctx context.Context, dealID, userID uuid.UUID) (*entity.DealNDA, error) {
	var n entity.DealNDA
	return &n, r.db.GetContext(ctx, &n, `SELECT * FROM deal_ndas WHERE deal_id = $1 AND user_id = $2`, dealID, userID)
}
func (r *NDARepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	sets, args := buildUpdateArgs(updates)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE deal_ndas SET %s WHERE id = $%d`, sets, len(args)), args...)
	return err
}

// --- NegotiationRepo ---
type NegotiationRepo struct{ db *sqlx.DB }
func NewNegotiationRepo(db *sqlx.DB) *NegotiationRepo { return &NegotiationRepo{db: db} }

func (r *NegotiationRepo) Create(ctx context.Context, n *entity.DealNegotiation) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO deal_negotiations (id, deal_id, sender_id, message_type, body, amount_paise, equity_pct, parent_id, status)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		n.ID, n.DealID, n.SenderID, n.MessageType, n.Body, n.AmountPaise, n.EquityPct, n.ParentID, n.Status)
	return err
}
func (r *NegotiationRepo) GetByDeal(ctx context.Context, dealID uuid.UUID, limit int) ([]entity.DealNegotiation, error) {
	var ns []entity.DealNegotiation
	return ns, r.db.SelectContext(ctx, &ns, `SELECT * FROM deal_negotiations WHERE deal_id = $1 ORDER BY created_at DESC LIMIT $2`, dealID, limit)
}

// --- EscrowRepo ---
type EscrowRepo struct{ db *sqlx.DB }
func NewEscrowRepo(db *sqlx.DB) *EscrowRepo { return &EscrowRepo{db: db} }

func (r *EscrowRepo) Create(ctx context.Context, e *entity.DealEscrow) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO deal_escrow (id, deal_id, participant_id, amount_paise, status, escrow_ref)
		VALUES ($1,$2,$3,$4,$5,$6)`, e.ID, e.DealID, e.ParticipantID, e.AmountPaise, e.Status, e.EscrowRef)
	return err
}
func (r *EscrowRepo) GetByDeal(ctx context.Context, dealID uuid.UUID) ([]entity.DealEscrow, error) {
	var es []entity.DealEscrow
	return es, r.db.SelectContext(ctx, &es, `SELECT * FROM deal_escrow WHERE deal_id = $1 ORDER BY created_at`, dealID)
}
func (r *EscrowRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	sets, args := buildUpdateArgs(updates)
	args = append(args, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE deal_escrow SET %s WHERE id = $%d`, sets, len(args)), args...)
	return err
}

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/dealance/services/wallet/internal/domain/entity"
)

func buildUpdateArgs(u map[string]interface{}) (string, []interface{}) {
	s := make([]string, 0, len(u)); a := make([]interface{}, 0, len(u)); i := 1
	for c, v := range u { s = append(s, fmt.Sprintf("%s = $%d", c, i)); a = append(a, v); i++ }
	return strings.Join(s, ", "), a
}

// --- WalletRepo ---
type WalletRepo struct{ db *sqlx.DB }
func NewWalletRepo(db *sqlx.DB) *WalletRepo { return &WalletRepo{db: db} }

func (r *WalletRepo) Create(ctx context.Context, w *entity.Wallet) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO wallets (id, user_id, currency) VALUES ($1,$2,$3) ON CONFLICT (user_id) DO NOTHING`, w.ID, w.UserID, w.Currency)
	return err
}
func (r *WalletRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error) {
	var w entity.Wallet; return &w, r.db.GetContext(ctx, &w, `SELECT * FROM wallets WHERE id = $1`, id)
}
func (r *WalletRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error) {
	var w entity.Wallet; return &w, r.db.GetContext(ctx, &w, `SELECT * FROM wallets WHERE user_id = $1`, userID)
}
func (r *WalletRepo) UpdateBalance(ctx context.Context, id uuid.UUID, balanceDelta, lockedDelta int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE wallets SET balance_paise = balance_paise + $1, locked_paise = locked_paise + $2 WHERE id = $3`,
		balanceDelta, lockedDelta, id)
	return err
}
func (r *WalletRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE wallets SET status = $1 WHERE id = $2`, status, id)
	return err
}

// --- LedgerRepo ---
type LedgerRepo struct{ db *sqlx.DB }
func NewLedgerRepo(db *sqlx.DB) *LedgerRepo { return &LedgerRepo{db: db} }

func (r *LedgerRepo) Create(ctx context.Context, e *entity.LedgerEntry) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO ledger_entries (id, wallet_id, entry_type, amount_paise, balance_after, category, reference_type, reference_id, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		e.ID, e.WalletID, e.EntryType, e.AmountPaise, e.BalanceAfter, e.Category, e.ReferenceType, e.ReferenceID, e.Description)
	return err
}
func (r *LedgerRepo) GetByWallet(ctx context.Context, walletID uuid.UUID, limit int, before *time.Time) ([]entity.LedgerEntry, error) {
	var entries []entity.LedgerEntry
	if before != nil {
		return entries, r.db.SelectContext(ctx, &entries,
			`SELECT * FROM ledger_entries WHERE wallet_id = $1 AND created_at < $3 ORDER BY created_at DESC LIMIT $2`, walletID, limit, before)
	}
	return entries, r.db.SelectContext(ctx, &entries,
		`SELECT * FROM ledger_entries WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2`, walletID, limit)
}

// --- TransactionRepo ---
type TransactionRepo struct{ db *sqlx.DB }
func NewTransactionRepo(db *sqlx.DB) *TransactionRepo { return &TransactionRepo{db: db} }

func (r *TransactionRepo) Create(ctx context.Context, tx *entity.Transaction) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions (id, wallet_id, tx_type, amount_paise, fee_paise, net_paise, status, payment_method, payment_ref, counterparty_id, deal_id, description)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		tx.ID, tx.WalletID, tx.TxType, tx.AmountPaise, tx.FeePaise, tx.NetPaise, tx.Status,
		tx.PaymentMethod, tx.PaymentRef, tx.CounterpartyID, tx.DealID, tx.Description)
	return err
}
func (r *TransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var tx entity.Transaction; return &tx, r.db.GetContext(ctx, &tx, `SELECT * FROM transactions WHERE id = $1`, id)
}
func (r *TransactionRepo) GetByWallet(ctx context.Context, walletID uuid.UUID, limit int, before *time.Time) ([]entity.Transaction, error) {
	var txs []entity.Transaction
	if before != nil {
		return txs, r.db.SelectContext(ctx, &txs,
			`SELECT * FROM transactions WHERE wallet_id = $1 AND created_at < $3 ORDER BY created_at DESC LIMIT $2`, walletID, limit, before)
	}
	return txs, r.db.SelectContext(ctx, &txs,
		`SELECT * FROM transactions WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2`, walletID, limit)
}
func (r *TransactionRepo) Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 { return nil }
	s, a := buildUpdateArgs(updates); a = append(a, id)
	_, err := r.db.ExecContext(ctx, fmt.Sprintf(`UPDATE transactions SET %s WHERE id = $%d`, s, len(a)), a...)
	return err
}

// --- BankAccountRepo ---
type BankAccountRepo struct{ db *sqlx.DB }
func NewBankAccountRepo(db *sqlx.DB) *BankAccountRepo { return &BankAccountRepo{db: db} }

func (r *BankAccountRepo) Create(ctx context.Context, a *entity.BankAccount) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bank_accounts (id, user_id, account_holder, account_number, ifsc_code, bank_name, account_type, is_primary)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		a.ID, a.UserID, a.AccountHolder, a.AccountNumber, a.IFSCCode, a.BankName, a.AccountType, a.IsPrimary)
	return err
}
func (r *BankAccountRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.BankAccount, error) {
	var accounts []entity.BankAccount
	return accounts, r.db.SelectContext(ctx, &accounts, `SELECT * FROM bank_accounts WHERE user_id = $1 ORDER BY is_primary DESC`, userID)
}
func (r *BankAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.BankAccount, error) {
	var a entity.BankAccount; return &a, r.db.GetContext(ctx, &a, `SELECT * FROM bank_accounts WHERE id = $1`, id)
}
func (r *BankAccountRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM bank_accounts WHERE id = $1`, id); return err
}

// --- WebhookRepo ---
type WebhookRepo struct{ db *sqlx.DB }
func NewWebhookRepo(db *sqlx.DB) *WebhookRepo { return &WebhookRepo{db: db} }

func (r *WebhookRepo) Create(ctx context.Context, w *entity.PaymentWebhook) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO payment_webhooks (id, provider, event_type, event_id, payload, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		w.ID, w.Provider, w.EventType, w.EventID, "{}", w.Status)
	return err
}
func (r *WebhookRepo) ExistsByEventID(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	return exists, r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM payment_webhooks WHERE event_id = $1)`, eventID)
}
func (r *WebhookRepo) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payment_webhooks SET status = 'PROCESSED', processed_at = NOW() WHERE id = $1`, id)
	return err
}

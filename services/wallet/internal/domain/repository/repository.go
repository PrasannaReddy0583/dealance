package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/dealance/services/wallet/internal/domain/entity"
)

type WalletRepository interface {
	Create(ctx context.Context, w *entity.Wallet) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Wallet, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Wallet, error)
	UpdateBalance(ctx context.Context, id uuid.UUID, balanceDelta, lockedDelta int64) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type LedgerRepository interface {
	Create(ctx context.Context, entry *entity.LedgerEntry) error
	GetByWallet(ctx context.Context, walletID uuid.UUID, limit int, before *time.Time) ([]entity.LedgerEntry, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *entity.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	GetByWallet(ctx context.Context, walletID uuid.UUID, limit int, before *time.Time) ([]entity.Transaction, error)
	Update(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type BankAccountRepository interface {
	Create(ctx context.Context, account *entity.BankAccount) error
	GetByUser(ctx context.Context, userID uuid.UUID) ([]entity.BankAccount, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.BankAccount, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type WebhookRepository interface {
	Create(ctx context.Context, webhook *entity.PaymentWebhook) error
	ExistsByEventID(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, id uuid.UUID) error
}

type CacheRepository interface {
	CacheBalance(ctx context.Context, userID string, balance, locked int64) error
	GetBalance(ctx context.Context, userID string) (balance, locked int64, err error)
	InvalidateBalance(ctx context.Context, userID string) error
}

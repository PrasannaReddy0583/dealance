package application

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dealance/services/wallet/internal/domain/entity"
	"github.com/dealance/services/wallet/internal/domain/repository"
	apperrors "github.com/dealance/shared/domain/errors"
)

type WalletService struct {
	walletRepo  repository.WalletRepository
	ledgerRepo  repository.LedgerRepository
	txRepo      repository.TransactionRepository
	bankRepo    repository.BankAccountRepository
	cacheRepo   repository.CacheRepository
	log         zerolog.Logger
}

func NewWalletService(
	walletRepo repository.WalletRepository, ledgerRepo repository.LedgerRepository,
	txRepo repository.TransactionRepository, bankRepo repository.BankAccountRepository,
	cacheRepo repository.CacheRepository, log zerolog.Logger,
) *WalletService {
	return &WalletService{walletRepo: walletRepo, ledgerRepo: ledgerRepo, txRepo: txRepo, bankRepo: bankRepo, cacheRepo: cacheRepo, log: log}
}

// GetOrCreateWallet returns user's wallet, creating one if it doesn't exist
func (s *WalletService) GetOrCreateWallet(ctx context.Context, userID string) (*entity.WalletResponse, error) {
	uID, _ := uuid.Parse(userID)
	wallet, err := s.walletRepo.GetByUserID(ctx, uID)
	if err != nil {
		// Create new wallet
		wallet = &entity.Wallet{ID: uuid.New(), UserID: uID, Currency: "INR", Status: "ACTIVE"}
		if err := s.walletRepo.Create(ctx, wallet); err != nil {
			return nil, apperrors.ErrInternal().WithInternal(err)
		}
		wallet, _ = s.walletRepo.GetByUserID(ctx, uID)
	}
	return s.toWalletResponse(wallet), nil
}

// GetBalance from cache or DB
func (s *WalletService) GetBalance(ctx context.Context, userID string) (*entity.WalletResponse, error) {
	uID, _ := uuid.Parse(userID)
	// Try cache first
	if bal, locked, err := s.cacheRepo.GetBalance(ctx, userID); err == nil {
		return &entity.WalletResponse{UserID: userID, BalancePaise: bal, LockedPaise: locked, AvailablePaise: bal - locked, Currency: "INR", Status: "ACTIVE"}, nil
	}
	wallet, err := s.walletRepo.GetByUserID(ctx, uID)
	if err != nil { return nil, apperrors.ErrNotFound("Wallet") }
	_ = s.cacheRepo.CacheBalance(ctx, userID, wallet.BalancePaise, wallet.LockedPaise)
	return s.toWalletResponse(wallet), nil
}

// Deposit — add funds to wallet
func (s *WalletService) Deposit(ctx context.Context, userID string, req entity.DepositRequest) (*entity.TransactionListItem, error) {
	uID, _ := uuid.Parse(userID)
	wallet, err := s.walletRepo.GetByUserID(ctx, uID)
	if err != nil { return nil, apperrors.ErrNotFound("Wallet") }
	if wallet.Status != "ACTIVE" { return nil, apperrors.ErrForbidden("WALLET_FROZEN", "Wallet is not active") }

	tx := &entity.Transaction{
		ID: uuid.New(), WalletID: wallet.ID, TxType: "DEPOSIT", AmountPaise: req.AmountPaise,
		FeePaise: 0, NetPaise: req.AmountPaise, Status: "COMPLETED",
		PaymentMethod: sql.NullString{String: req.PaymentMethod, Valid: true},
		PaymentRef: sql.NullString{String: req.PaymentRef, Valid: req.PaymentRef != ""},
	}
	if err := s.txRepo.Create(ctx, tx); err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }

	// Update balance
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, req.AmountPaise, 0); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	// Ledger entry
	newBalance := wallet.BalancePaise + req.AmountPaise
	ledger := &entity.LedgerEntry{
		ID: uuid.New(), WalletID: wallet.ID, EntryType: "CREDIT", AmountPaise: req.AmountPaise,
		BalanceAfter: newBalance, Category: "DEPOSIT",
		Description: sql.NullString{String: "Deposit via " + req.PaymentMethod, Valid: true},
	}
	_ = s.ledgerRepo.Create(ctx, ledger)
	_ = s.cacheRepo.InvalidateBalance(ctx, userID)

	return &entity.TransactionListItem{
		ID: tx.ID.String(), TxType: "DEPOSIT", AmountPaise: req.AmountPaise,
		NetPaise: req.AmountPaise, Status: "COMPLETED", CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// Withdraw — withdraw to bank account
func (s *WalletService) Withdraw(ctx context.Context, userID string, req entity.WithdrawRequest) (*entity.TransactionListItem, error) {
	uID, _ := uuid.Parse(userID)
	wallet, err := s.walletRepo.GetByUserID(ctx, uID)
	if err != nil { return nil, apperrors.ErrNotFound("Wallet") }
	if wallet.Status != "ACTIVE" { return nil, apperrors.ErrForbidden("WALLET_FROZEN", "Wallet is not active") }

	available := wallet.BalancePaise - wallet.LockedPaise
	if req.AmountPaise > available {
		return nil, apperrors.ErrValidation("insufficient balance")
	}

	// Verify bank account ownership
	baID, _ := uuid.Parse(req.BankAccountID)
	ba, err := s.bankRepo.GetByID(ctx, baID)
	if err != nil { return nil, apperrors.ErrNotFound("Bank account") }
	if ba.UserID != uID { return nil, apperrors.ErrForbidden("NOT_OWNER", "Bank account does not belong to user") }

	tx := &entity.Transaction{
		ID: uuid.New(), WalletID: wallet.ID, TxType: "WITHDRAWAL", AmountPaise: req.AmountPaise,
		FeePaise: 0, NetPaise: req.AmountPaise, Status: "PROCESSING",
		Description: sql.NullString{String: "Withdrawal to " + ba.BankName, Valid: true},
	}
	if err := s.txRepo.Create(ctx, tx); err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }

	// Debit balance
	if err := s.walletRepo.UpdateBalance(ctx, wallet.ID, -req.AmountPaise, 0); err != nil {
		return nil, apperrors.ErrInternal().WithInternal(err)
	}

	newBalance := wallet.BalancePaise - req.AmountPaise
	ledger := &entity.LedgerEntry{
		ID: uuid.New(), WalletID: wallet.ID, EntryType: "DEBIT", AmountPaise: req.AmountPaise,
		BalanceAfter: newBalance, Category: "WITHDRAWAL",
		Description: sql.NullString{String: "Withdrawal to " + ba.BankName, Valid: true},
	}
	_ = s.ledgerRepo.Create(ctx, ledger)
	_ = s.cacheRepo.InvalidateBalance(ctx, userID)

	return &entity.TransactionListItem{
		ID: tx.ID.String(), TxType: "WITHDRAWAL", AmountPaise: req.AmountPaise,
		NetPaise: req.AmountPaise, Status: "PROCESSING", CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// Transfer — wallet-to-wallet transfer
func (s *WalletService) Transfer(ctx context.Context, userID string, req entity.TransferRequest) error {
	fromUID, _ := uuid.Parse(userID)
	toUID, _ := uuid.Parse(req.ToUserID)

	fromWallet, err := s.walletRepo.GetByUserID(ctx, fromUID)
	if err != nil { return apperrors.ErrNotFound("Sender wallet") }
	toWallet, err := s.walletRepo.GetByUserID(ctx, toUID)
	if err != nil { return apperrors.ErrNotFound("Recipient wallet") }

	available := fromWallet.BalancePaise - fromWallet.LockedPaise
	if req.AmountPaise > available { return apperrors.ErrValidation("insufficient balance") }

	// Debit sender
	_ = s.walletRepo.UpdateBalance(ctx, fromWallet.ID, -req.AmountPaise, 0)
	// Credit receiver
	_ = s.walletRepo.UpdateBalance(ctx, toWallet.ID, req.AmountPaise, 0)

	// Sender tx + ledger
	sTx := &entity.Transaction{ID: uuid.New(), WalletID: fromWallet.ID, TxType: "TRANSFER", AmountPaise: req.AmountPaise, NetPaise: req.AmountPaise, Status: "COMPLETED", CounterpartyID: &toWallet.ID, Description: sql.NullString{String: req.Description, Valid: req.Description != ""}}
	_ = s.txRepo.Create(ctx, sTx)
	sLedger := &entity.LedgerEntry{ID: uuid.New(), WalletID: fromWallet.ID, EntryType: "DEBIT", AmountPaise: req.AmountPaise, BalanceAfter: fromWallet.BalancePaise - req.AmountPaise, Category: "TRANSFER"}
	_ = s.ledgerRepo.Create(ctx, sLedger)

	// Receiver tx + ledger
	rTx := &entity.Transaction{ID: uuid.New(), WalletID: toWallet.ID, TxType: "TRANSFER", AmountPaise: req.AmountPaise, NetPaise: req.AmountPaise, Status: "COMPLETED", CounterpartyID: &fromWallet.ID}
	_ = s.txRepo.Create(ctx, rTx)
	rLedger := &entity.LedgerEntry{ID: uuid.New(), WalletID: toWallet.ID, EntryType: "CREDIT", AmountPaise: req.AmountPaise, BalanceAfter: toWallet.BalancePaise + req.AmountPaise, Category: "TRANSFER"}
	_ = s.ledgerRepo.Create(ctx, rLedger)

	_ = s.cacheRepo.InvalidateBalance(ctx, userID)
	_ = s.cacheRepo.InvalidateBalance(ctx, req.ToUserID)
	return nil
}

// GetTransactions — paginated history
func (s *WalletService) GetTransactions(ctx context.Context, userID string, limit int) ([]entity.TransactionListItem, error) {
	uID, _ := uuid.Parse(userID)
	wallet, err := s.walletRepo.GetByUserID(ctx, uID)
	if err != nil { return nil, apperrors.ErrNotFound("Wallet") }
	if limit <= 0 || limit > 50 { limit = 20 }
	txs, err := s.txRepo.GetByWallet(ctx, wallet.ID, limit, nil)
	if err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }
	items := make([]entity.TransactionListItem, len(txs))
	for i, tx := range txs {
		items[i] = entity.TransactionListItem{
			ID: tx.ID.String(), TxType: tx.TxType, AmountPaise: tx.AmountPaise,
			FeePaise: tx.FeePaise, NetPaise: tx.NetPaise, Status: tx.Status,
			CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if tx.Description.Valid { items[i].Description = tx.Description.String }
	}
	return items, nil
}

// GetLedger — double-entry ledger history
func (s *WalletService) GetLedger(ctx context.Context, userID string, limit int) ([]entity.LedgerListItem, error) {
	uID, _ := uuid.Parse(userID)
	wallet, err := s.walletRepo.GetByUserID(ctx, uID)
	if err != nil { return nil, apperrors.ErrNotFound("Wallet") }
	if limit <= 0 || limit > 50 { limit = 20 }
	entries, err := s.ledgerRepo.GetByWallet(ctx, wallet.ID, limit, nil)
	if err != nil { return nil, apperrors.ErrInternal().WithInternal(err) }
	items := make([]entity.LedgerListItem, len(entries))
	for i, e := range entries {
		items[i] = entity.LedgerListItem{
			ID: e.ID.String(), EntryType: e.EntryType, AmountPaise: e.AmountPaise,
			BalanceAfter: e.BalanceAfter, Category: e.Category,
			CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if e.Description.Valid { items[i].Description = e.Description.String }
	}
	return items, nil
}

// AddBankAccount
func (s *WalletService) AddBankAccount(ctx context.Context, userID string, req entity.AddBankAccountRequest) error {
	uID, _ := uuid.Parse(userID)
	accType := "SAVINGS"
	if req.AccountType != "" { accType = req.AccountType }
	account := &entity.BankAccount{
		ID: uuid.New(), UserID: uID, AccountHolder: req.AccountHolder,
		AccountNumber: req.AccountNumber, IFSCCode: req.IFSCCode, BankName: req.BankName,
		AccountType: accType, IsPrimary: req.IsPrimary,
	}
	return s.bankRepo.Create(ctx, account)
}

func (s *WalletService) GetBankAccounts(ctx context.Context, userID string) ([]entity.BankAccount, error) {
	uID, _ := uuid.Parse(userID)
	return s.bankRepo.GetByUser(ctx, uID)
}

func (s *WalletService) RemoveBankAccount(ctx context.Context, userID, accountID string) error {
	uID, _ := uuid.Parse(userID)
	aID, _ := uuid.Parse(accountID)
	account, err := s.bankRepo.GetByID(ctx, aID)
	if err != nil { return apperrors.ErrNotFound("Bank account") }
	if account.UserID != uID { return apperrors.ErrForbidden("NOT_OWNER", "Not your bank account") }
	return s.bankRepo.Delete(ctx, aID)
}

func (s *WalletService) toWalletResponse(w *entity.Wallet) *entity.WalletResponse {
	return &entity.WalletResponse{
		ID: w.ID.String(), UserID: w.UserID.String(),
		BalancePaise: w.BalancePaise, LockedPaise: w.LockedPaise,
		AvailablePaise: w.BalancePaise - w.LockedPaise,
		Currency: w.Currency, Status: w.Status, KYCVerified: w.KYCVerified,
	}
}

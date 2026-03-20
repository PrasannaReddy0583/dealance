package entity

type WalletResponse struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	BalancePaise int64  `json:"balance_paise"`
	LockedPaise  int64  `json:"locked_paise"`
	AvailablePaise int64 `json:"available_paise"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	KYCVerified  bool   `json:"kyc_verified"`
}

type DepositRequest struct {
	AmountPaise   int64  `json:"amount_paise" validate:"required,gt=0"`
	PaymentMethod string `json:"payment_method" validate:"required,oneof=UPI NEFT RTGS IMPS NET_BANKING"`
	PaymentRef    string `json:"payment_ref,omitempty"`
}

type WithdrawRequest struct {
	AmountPaise   int64  `json:"amount_paise" validate:"required,gt=0"`
	BankAccountID string `json:"bank_account_id" validate:"required,uuid"`
}

type TransferRequest struct {
	ToUserID    string `json:"to_user_id" validate:"required,uuid"`
	AmountPaise int64  `json:"amount_paise" validate:"required,gt=0"`
	Description string `json:"description,omitempty"`
}

type AddBankAccountRequest struct {
	AccountHolder string `json:"account_holder" validate:"required,max=200"`
	AccountNumber string `json:"account_number" validate:"required,max=30"`
	IFSCCode      string `json:"ifsc_code" validate:"required,len=11"`
	BankName      string `json:"bank_name" validate:"required,max=100"`
	AccountType   string `json:"account_type,omitempty" validate:"omitempty,oneof=SAVINGS CURRENT"`
	IsPrimary     bool   `json:"is_primary,omitempty"`
}

type TransactionListItem struct {
	ID          string `json:"id"`
	TxType      string `json:"tx_type"`
	AmountPaise int64  `json:"amount_paise"`
	FeePaise    int64  `json:"fee_paise"`
	NetPaise    int64  `json:"net_paise"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type LedgerListItem struct {
	ID           string `json:"id"`
	EntryType    string `json:"entry_type"`
	AmountPaise  int64  `json:"amount_paise"`
	BalanceAfter int64  `json:"balance_after"`
	Category     string `json:"category"`
	Description  string `json:"description,omitempty"`
	CreatedAt    string `json:"created_at"`
}

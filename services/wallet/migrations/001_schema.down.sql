DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;
DROP TABLE IF EXISTS payment_webhooks;
DROP TABLE IF EXISTS bank_accounts;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS wallets;

DROP INDEX IF EXISTS idx_payments_contact;
DROP INDEX IF EXISTS idx_payments_company;
DROP INDEX IF EXISTS idx_bank_transactions_account;
DROP INDEX IF EXISTS idx_bank_transactions_company;
DROP INDEX IF EXISTS idx_bank_accounts_company;

DROP TABLE IF EXISTS cash_transactions;
DROP TABLE IF EXISTS bank_reconciliations;
DROP TABLE IF EXISTS payment_lines;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS bank_transactions;
DROP TABLE IF EXISTS bank_accounts;

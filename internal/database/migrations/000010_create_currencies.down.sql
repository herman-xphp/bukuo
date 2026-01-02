-- Remove currency columns from journal_lines
ALTER TABLE journal_lines DROP COLUMN IF EXISTS exchange_rate;
ALTER TABLE journal_lines DROP COLUMN IF EXISTS foreign_amount;
ALTER TABLE journal_lines DROP COLUMN IF EXISTS currency_id;

-- Remove currency column from accounts
ALTER TABLE accounts DROP COLUMN IF EXISTS currency_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_exchange_rates_date;
DROP INDEX IF EXISTS idx_exchange_rates_pair;
DROP INDEX IF EXISTS idx_exchange_rates_company;
DROP INDEX IF EXISTS idx_currencies_base;
DROP INDEX IF EXISTS idx_currencies_company;

-- Drop tables
DROP TABLE IF EXISTS exchange_rates;
DROP TABLE IF EXISTS currencies;

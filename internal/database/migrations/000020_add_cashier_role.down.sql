-- Revert CASHIER role from users table
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('OWNER', 'ADMIN', 'ACCOUNTANT', 'VIEWER'));

-- Note: Any users with CASHIER role will need to be manually updated before running this migration

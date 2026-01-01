-- Remove user lockout columns
ALTER TABLE users DROP COLUMN IF EXISTS failed_attempts;
ALTER TABLE users DROP COLUMN IF EXISTS locked_until;

-- Drop audit logs table
DROP TABLE IF EXISTS audit_logs;

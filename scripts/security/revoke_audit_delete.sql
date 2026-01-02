-- REVOKE DELETE PERMISSION ON AUDIT LOGS
-- This script ensures that the application user (and others) cannot delete audit logs,
-- enforcing immutability at the database level.

-- Replace 'bukuo_user' with your actual application database user
REVOKE DELETE ON TABLE audit_logs FROM bukuo_user;

-- Optional: Create a trigger to prevent accidental deletion by even the owner (if needed)
CREATE OR REPLACE FUNCTION prevent_audit_log_deletion()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Deletion of audit logs is strictly prohibited to maintain audit trail integrity.';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_audit_log_deletion
BEFORE DELETE ON audit_logs
FOR EACH ROW
EXECUTE FUNCTION prevent_audit_log_deletion();

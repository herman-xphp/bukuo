-- Remove approval columns
ALTER TABLE journal_entries DROP COLUMN IF EXISTS approved_at;
ALTER TABLE journal_entries DROP COLUMN IF EXISTS approved_by;
ALTER TABLE journal_entries DROP COLUMN IF EXISTS rejected_at;
ALTER TABLE journal_entries DROP COLUMN IF EXISTS rejected_by;
ALTER TABLE journal_entries DROP COLUMN IF EXISTS reject_reason;

-- Remove status check constraint
ALTER TABLE journal_entries DROP CONSTRAINT IF EXISTS journal_entries_status_check;

-- Drop index
DROP INDEX IF EXISTS idx_journal_entries_pending;

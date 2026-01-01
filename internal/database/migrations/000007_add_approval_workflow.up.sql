-- Add approval workflow columns to journal_entries
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS approved_at TIMESTAMPTZ;
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS approved_by UUID REFERENCES users(id);
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS rejected_at TIMESTAMPTZ;
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS rejected_by UUID REFERENCES users(id);
ALTER TABLE journal_entries ADD COLUMN IF NOT EXISTS reject_reason TEXT;

-- Update status constraint if exists (or add check constraint)
-- Note: PostgreSQL doesn't have ENUM, using varchar with valid values check
ALTER TABLE journal_entries DROP CONSTRAINT IF EXISTS journal_entries_status_check;
ALTER TABLE journal_entries ADD CONSTRAINT journal_entries_status_check 
  CHECK (status IN ('DRAFT', 'PENDING_APPROVAL', 'APPROVED', 'REJECTED', 'POSTED', 'REVERSED'));

-- Add index for pending approval queries
CREATE INDEX IF NOT EXISTS idx_journal_entries_pending ON journal_entries(company_id, status) 
  WHERE status = 'PENDING_APPROVAL';

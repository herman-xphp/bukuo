-- Accounting Periods
CREATE TABLE accounting_periods (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id UUID NOT NULL REFERENCES companies(id),
  name VARCHAR(50) NOT NULL,
  start_date DATE NOT NULL,
  end_date DATE NOT NULL,
  status VARCHAR(20) DEFAULT 'OPEN' CHECK (status IN ('OPEN','CLOSED', 'LOCKED')),
  closed_at TIMESTAMPTZ,
  closed_by UUID,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(company_id, start_date, end_date)
);

CREATE INDEX idx_periods_company ON accounting_periods(company_id);
CREATE INDEX idx_periods_status ON accounting_periods(status);

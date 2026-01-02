-- Tax Rates
CREATE TABLE tax_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(20) NOT NULL,
    type VARCHAR(20) NOT NULL,
    rate DECIMAL(5, 2) NOT NULL,
    sales_account_id UUID REFERENCES accounts(id),
    purchase_account_id UUID REFERENCES accounts(id),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, code)
);

-- Tax Returns
CREATE TABLE tax_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    period_id UUID REFERENCES periods(id),
    tax_rate_id UUID REFERENCES tax_rates(id),
    return_no VARCHAR(50) NOT NULL,
    return_date DATE NOT NULL,
    taxable_amount DECIMAL(20, 2) DEFAULT 0,
    tax_amount DECIMAL(20, 2) DEFAULT 0,
    credits DECIMAL(20, 2) DEFAULT 0,
    payable_amount DECIMAL(20, 2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'DRAFT',
    notes TEXT,
    journal_id UUID REFERENCES journals(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, return_no)
);

CREATE INDEX idx_tax_returns_company ON tax_returns(company_id);
CREATE INDEX idx_tax_returns_period ON tax_returns(period_id);

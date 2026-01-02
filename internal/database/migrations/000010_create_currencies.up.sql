-- Currencies table
CREATE TABLE currencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(3) NOT NULL,
    name VARCHAR(100) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    decimal_places INT DEFAULT 2,
    is_base BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX idx_currencies_company ON currencies(company_id);
CREATE INDEX idx_currencies_base ON currencies(company_id, is_base) WHERE is_base = true;

-- Exchange rates table
CREATE TABLE exchange_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    from_currency_id UUID NOT NULL REFERENCES currencies(id) ON DELETE CASCADE,
    to_currency_id UUID NOT NULL REFERENCES currencies(id) ON DELETE CASCADE,
    rate DECIMAL(20, 10) NOT NULL,
    effective_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, from_currency_id, to_currency_id, effective_date)
);

CREATE INDEX idx_exchange_rates_company ON exchange_rates(company_id);
CREATE INDEX idx_exchange_rates_pair ON exchange_rates(company_id, from_currency_id, to_currency_id);
CREATE INDEX idx_exchange_rates_date ON exchange_rates(company_id, effective_date);

-- Add currency support to accounts (optional currency-specific accounts)
ALTER TABLE accounts ADD COLUMN currency_id UUID REFERENCES currencies(id);

-- Add currency support to journal lines for foreign currency transactions
ALTER TABLE journal_lines ADD COLUMN currency_id UUID REFERENCES currencies(id);
ALTER TABLE journal_lines ADD COLUMN foreign_amount DECIMAL(20, 2);
ALTER TABLE journal_lines ADD COLUMN exchange_rate DECIMAL(20, 10);

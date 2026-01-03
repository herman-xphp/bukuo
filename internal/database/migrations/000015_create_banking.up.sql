-- Bank Accounts
CREATE TABLE bank_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id),
    bank_name VARCHAR(100) NOT NULL,
    account_number VARCHAR(50) NOT NULL,
    account_name VARCHAR(100) NOT NULL,
    currency_id UUID REFERENCES currencies(id),
    current_balance DECIMAL(20, 2) DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, account_number)
);

-- Bank Transactions
CREATE TABLE bank_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id),
    transaction_no VARCHAR(50) NOT NULL,
    transaction_date DATE NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    description TEXT,
    reference VARCHAR(100),
    journal_id UUID REFERENCES journal_entries(id),
    reconciled BOOLEAN DEFAULT FALSE,
    reconciled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, transaction_no)
);

-- Payments (AR Collection / AP Payment)
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    payment_no VARCHAR(50) NOT NULL,
    payment_type VARCHAR(10) NOT NULL CHECK (payment_type IN ('RECEIVE', 'PAY')),
    contact_id UUID NOT NULL REFERENCES contacts(id),
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id),
    payment_date DATE NOT NULL,
    payment_method VARCHAR(20) NOT NULL,
    total_amount DECIMAL(20, 2) NOT NULL,
    notes TEXT,
    journal_id UUID REFERENCES journal_entries(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, payment_no)
);

-- Payment Lines
CREATE TABLE payment_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    invoice_id UUID NOT NULL,
    invoice_no VARCHAR(50),
    invoice_amount DECIMAL(20, 2) NOT NULL,
    paid_amount DECIMAL(20, 2) NOT NULL
);

-- Bank Reconciliations
CREATE TABLE bank_reconciliations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(id),
    period_id UUID REFERENCES accounting_periods(id),
    statement_date DATE NOT NULL,
    statement_balance DECIMAL(20, 2) NOT NULL,
    book_balance DECIMAL(20, 2) NOT NULL,
    difference DECIMAL(20, 2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'DRAFT',
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Cash Transactions
CREATE TABLE cash_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    transaction_no VARCHAR(50) NOT NULL,
    transaction_date DATE NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    description TEXT,
    category VARCHAR(50),
    journal_id UUID REFERENCES journal_entries(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, transaction_no)
);

-- Indexes
CREATE INDEX idx_bank_accounts_company ON bank_accounts(company_id);
CREATE INDEX idx_bank_transactions_company ON bank_transactions(company_id);
CREATE INDEX idx_bank_transactions_account ON bank_transactions(bank_account_id);
CREATE INDEX idx_payments_company ON payments(company_id);
CREATE INDEX idx_payments_contact ON payments(contact_id);

-- Create contacts table
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    contact_type VARCHAR(20) NOT NULL CHECK (contact_type IN ('CUSTOMER', 'SUPPLIER', 'BOTH')),
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    tax_id VARCHAR(50),
    credit_limit DECIMAL(20, 2) DEFAULT 0,
    payment_term_days INT DEFAULT 30,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, code)
);

-- Indexes for better query performance
CREATE INDEX idx_contacts_company ON contacts(company_id);
CREATE INDEX idx_contacts_type ON contacts(company_id, contact_type);
CREATE INDEX idx_contacts_name ON contacts(company_id, name);
CREATE INDEX idx_contacts_active ON contacts(company_id, is_active);

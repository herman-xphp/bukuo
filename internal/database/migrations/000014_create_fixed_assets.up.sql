-- Asset Categories
CREATE TABLE asset_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    default_life_months INT DEFAULT 60,
    default_method VARCHAR(30) DEFAULT 'STRAIGHT_LINE',
    asset_account_id UUID REFERENCES accounts(id),
    depreciation_account_id UUID REFERENCES accounts(id),
    accum_depreciation_account_id UUID REFERENCES accounts(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, name)
);

-- Fixed Assets
CREATE TABLE fixed_assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    asset_code VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES asset_categories(id),
    acquisition_date DATE NOT NULL,
    acquisition_cost DECIMAL(20, 2) NOT NULL,
    residual_value DECIMAL(20, 2) DEFAULT 0,
    useful_life_months INT NOT NULL,
    depreciation_method VARCHAR(30) NOT NULL DEFAULT 'STRAIGHT_LINE',
    accumulated_depreciation DECIMAL(20, 2) DEFAULT 0,
    net_book_value DECIMAL(20, 2) NOT NULL,
    asset_account_id UUID REFERENCES accounts(id),
    depreciation_account_id UUID REFERENCES accounts(id),
    accum_depreciation_account_id UUID REFERENCES accounts(id),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    disposal_date DATE,
    disposal_amount DECIMAL(20, 2) DEFAULT 0,
    location VARCHAR(200),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, asset_code)
);

-- Depreciation Entries
CREATE TABLE depreciation_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES fixed_assets(id),
    period_id UUID REFERENCES accounting_periods(id),
    depreciation_date DATE NOT NULL,
    depreciation_amount DECIMAL(20, 2) NOT NULL,
    accum_depreciation_after DECIMAL(20, 2) NOT NULL,
    net_book_value_after DECIMAL(20, 2) NOT NULL,
    journal_id UUID REFERENCES journal_entries(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_fixed_assets_company ON fixed_assets(company_id);
CREATE INDEX idx_fixed_assets_category ON fixed_assets(category_id);
CREATE INDEX idx_depreciation_entries_asset ON depreciation_entries(asset_id);
CREATE INDEX idx_depreciation_entries_period ON depreciation_entries(period_id);

-- Units of Measure
CREATE TABLE units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX idx_units_company ON units(company_id);

-- Product Categories
CREATE TABLE product_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    parent_id UUID REFERENCES product_categories(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_product_categories_company ON product_categories(company_id);
CREATE INDEX idx_product_categories_parent ON product_categories(parent_id);

-- Products
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('GOODS', 'SERVICE')),
    category_id UUID REFERENCES product_categories(id) ON DELETE SET NULL,
    unit_id UUID NOT NULL REFERENCES units(id),
    description TEXT,
    sales_price DECIMAL(20, 2) DEFAULT 0,
    purchase_price DECIMAL(20, 2) DEFAULT 0,
    sales_account_id UUID NOT NULL REFERENCES accounts(id),
    purchase_account_id UUID NOT NULL REFERENCES accounts(id),
    inventory_account_id UUID REFERENCES accounts(id),
    is_active BOOLEAN DEFAULT TRUE,
    min_stock DECIMAL(20, 4) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX idx_products_company ON products(company_id);
CREATE INDEX idx_products_type ON products(company_id, type);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_active ON products(company_id, is_active);

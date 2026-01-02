-- Warehouses
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, code)
);

CREATE INDEX idx_warehouses_company ON warehouses(company_id);

-- Inventory Transactions
CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    transaction_no VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('STOCK_IN', 'STOCK_OUT', 'TRANSFER', 'ADJUSTMENT')),
    product_id UUID NOT NULL REFERENCES products(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    to_warehouse_id UUID REFERENCES warehouses(id),
    quantity DECIMAL(20, 4) NOT NULL,
    unit_cost DECIMAL(20, 2) DEFAULT 0,
    total_cost DECIMAL(20, 2) DEFAULT 0,
    reference VARCHAR(100),
    notes TEXT,
    transaction_date TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, transaction_no)
);

CREATE INDEX idx_inv_tx_company ON inventory_transactions(company_id);
CREATE INDEX idx_inv_tx_product ON inventory_transactions(product_id);
CREATE INDEX idx_inv_tx_warehouse ON inventory_transactions(warehouse_id);
CREATE INDEX idx_inv_tx_date ON inventory_transactions(transaction_date);

-- Product Stocks (current stock per product per warehouse)
CREATE TABLE product_stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    quantity DECIMAL(20, 4) DEFAULT 0,
    average_cost DECIMAL(20, 2) DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, product_id, warehouse_id)
);

CREATE INDEX idx_product_stocks_company ON product_stocks(company_id);
CREATE INDEX idx_product_stocks_product ON product_stocks(product_id);

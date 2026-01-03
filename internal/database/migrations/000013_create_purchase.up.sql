-- Purchase Requests
CREATE TABLE purchase_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    request_no VARCHAR(50) NOT NULL,
    request_date DATE NOT NULL,
    requested_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, request_no)
);

CREATE TABLE purchase_request_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL REFERENCES purchase_requests(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    description TEXT,
    quantity DECIMAL(20, 4) NOT NULL
);

-- Purchase Orders
CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    order_no VARCHAR(50) NOT NULL,
    supplier_id UUID NOT NULL REFERENCES contacts(id),
    request_id UUID REFERENCES purchase_requests(id),
    order_date DATE NOT NULL,
    delivery_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    subtotal DECIMAL(20, 2) DEFAULT 0,
    tax_amount DECIMAL(20, 2) DEFAULT 0,
    discount_amount DECIMAL(20, 2) DEFAULT 0,
    total DECIMAL(20, 2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, order_no)
);

CREATE TABLE purchase_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    description TEXT,
    quantity DECIMAL(20, 4) NOT NULL,
    received_qty DECIMAL(20, 4) DEFAULT 0,
    unit_price DECIMAL(20, 2) NOT NULL,
    discount_pct DECIMAL(5, 2) DEFAULT 0,
    tax_pct DECIMAL(5, 2) DEFAULT 0,
    line_total DECIMAL(20, 2) NOT NULL
);

-- Goods Received Notes
CREATE TABLE goods_received_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    grn_no VARCHAR(50) NOT NULL,
    supplier_id UUID NOT NULL REFERENCES contacts(id),
    order_id UUID NOT NULL REFERENCES purchase_orders(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    receive_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, grn_no)
);

CREATE TABLE grn_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    grn_id UUID NOT NULL REFERENCES goods_received_notes(id) ON DELETE CASCADE,
    order_line_id UUID NOT NULL REFERENCES purchase_order_lines(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(20, 4) NOT NULL,
    unit_cost DECIMAL(20, 2) NOT NULL
);

-- Purchase Invoices
CREATE TABLE purchase_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    invoice_no VARCHAR(50) NOT NULL,
    supplier_id UUID NOT NULL REFERENCES contacts(id),
    order_id UUID REFERENCES purchase_orders(id),
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    subtotal DECIMAL(20, 2) DEFAULT 0,
    tax_amount DECIMAL(20, 2) DEFAULT 0,
    discount_amount DECIMAL(20, 2) DEFAULT 0,
    total DECIMAL(20, 2) DEFAULT 0,
    paid_amount DECIMAL(20, 2) DEFAULT 0,
    notes TEXT,
    journal_id UUID REFERENCES journal_entries(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, invoice_no)
);

CREATE TABLE purchase_invoice_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES purchase_invoices(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    description TEXT,
    quantity DECIMAL(20, 4) NOT NULL,
    unit_price DECIMAL(20, 2) NOT NULL,
    discount_pct DECIMAL(5, 2) DEFAULT 0,
    tax_pct DECIMAL(5, 2) DEFAULT 0,
    line_total DECIMAL(20, 2) NOT NULL
);

-- Purchase Returns
CREATE TABLE purchase_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    return_no VARCHAR(50) NOT NULL,
    supplier_id UUID NOT NULL REFERENCES contacts(id),
    invoice_id UUID NOT NULL REFERENCES purchase_invoices(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    return_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    total DECIMAL(20, 2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, return_no)
);

CREATE TABLE purchase_return_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_id UUID NOT NULL REFERENCES purchase_returns(id) ON DELETE CASCADE,
    invoice_line_id UUID NOT NULL REFERENCES purchase_invoice_lines(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(20, 4) NOT NULL,
    unit_price DECIMAL(20, 2) NOT NULL,
    line_total DECIMAL(20, 2) NOT NULL
);

-- Indexes
CREATE INDEX idx_purchase_orders_company ON purchase_orders(company_id);
CREATE INDEX idx_purchase_orders_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_purchase_invoices_company ON purchase_invoices(company_id);
CREATE INDEX idx_purchase_invoices_supplier ON purchase_invoices(supplier_id);
CREATE INDEX idx_grn_company ON goods_received_notes(company_id);

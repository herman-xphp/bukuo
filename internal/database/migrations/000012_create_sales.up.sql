-- Sales Quotations
CREATE TABLE sales_quotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    quotation_no VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL REFERENCES contacts(id),
    quotation_date DATE NOT NULL,
    valid_until DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    subtotal DECIMAL(20, 2) DEFAULT 0,
    tax_amount DECIMAL(20, 2) DEFAULT 0,
    discount_amount DECIMAL(20, 2) DEFAULT 0,
    total DECIMAL(20, 2) DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, quotation_no)
);

CREATE TABLE sales_quotation_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quotation_id UUID NOT NULL REFERENCES sales_quotations(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    description TEXT,
    quantity DECIMAL(20, 4) NOT NULL,
    unit_price DECIMAL(20, 2) NOT NULL,
    discount_pct DECIMAL(5, 2) DEFAULT 0,
    tax_pct DECIMAL(5, 2) DEFAULT 0,
    line_total DECIMAL(20, 2) NOT NULL
);

-- Sales Orders
CREATE TABLE sales_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    order_no VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL REFERENCES contacts(id),
    quotation_id UUID REFERENCES sales_quotations(id),
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

CREATE TABLE sales_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    description TEXT,
    quantity DECIMAL(20, 4) NOT NULL,
    delivered_qty DECIMAL(20, 4) DEFAULT 0,
    unit_price DECIMAL(20, 2) NOT NULL,
    discount_pct DECIMAL(5, 2) DEFAULT 0,
    tax_pct DECIMAL(5, 2) DEFAULT 0,
    line_total DECIMAL(20, 2) NOT NULL
);

-- Delivery Orders
CREATE TABLE delivery_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    delivery_no VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL REFERENCES contacts(id),
    order_id UUID NOT NULL REFERENCES sales_orders(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    delivery_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, delivery_no)
);

CREATE TABLE delivery_order_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id UUID NOT NULL REFERENCES delivery_orders(id) ON DELETE CASCADE,
    order_line_id UUID NOT NULL REFERENCES sales_order_lines(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(20, 4) NOT NULL
);

-- Sales Invoices
CREATE TABLE sales_invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    invoice_no VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL REFERENCES contacts(id),
    order_id UUID REFERENCES sales_orders(id),
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

CREATE TABLE sales_invoice_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES sales_invoices(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    description TEXT,
    quantity DECIMAL(20, 4) NOT NULL,
    unit_price DECIMAL(20, 2) NOT NULL,
    discount_pct DECIMAL(5, 2) DEFAULT 0,
    tax_pct DECIMAL(5, 2) DEFAULT 0,
    line_total DECIMAL(20, 2) NOT NULL
);

-- Sales Returns
CREATE TABLE sales_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    return_no VARCHAR(50) NOT NULL,
    customer_id UUID NOT NULL REFERENCES contacts(id),
    invoice_id UUID NOT NULL REFERENCES sales_invoices(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    return_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    total DECIMAL(20, 2) DEFAULT 0,
    notes TEXT,
    credit_note_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(company_id, return_no)
);

CREATE TABLE sales_return_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    return_id UUID NOT NULL REFERENCES sales_returns(id) ON DELETE CASCADE,
    invoice_line_id UUID NOT NULL REFERENCES sales_invoice_lines(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(20, 4) NOT NULL,
    unit_price DECIMAL(20, 2) NOT NULL,
    line_total DECIMAL(20, 2) NOT NULL
);

-- Indexes
CREATE INDEX idx_sales_quotations_company ON sales_quotations(company_id);
CREATE INDEX idx_sales_quotations_customer ON sales_quotations(customer_id);
CREATE INDEX idx_sales_orders_company ON sales_orders(company_id);
CREATE INDEX idx_sales_orders_customer ON sales_orders(customer_id);
CREATE INDEX idx_sales_invoices_company ON sales_invoices(company_id);
CREATE INDEX idx_sales_invoices_customer ON sales_invoices(customer_id);
CREATE INDEX idx_delivery_orders_company ON delivery_orders(company_id);
CREATE INDEX idx_sales_returns_company ON sales_returns(company_id);

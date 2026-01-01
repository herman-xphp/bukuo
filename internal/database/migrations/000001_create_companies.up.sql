-- Companies (tenant/perusahaan)
CREATE TABLE companies (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  tax_id VARCHAR(50) NOT NULL,
  address TEXT,
  phone VARCHAR(20),
  email VARCHAR(255),
  fiscal_year_start INT DEFAULT 1, -- Bulan mulai tahun fiskal (1=Januari)
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_companies_name ON companies(name);

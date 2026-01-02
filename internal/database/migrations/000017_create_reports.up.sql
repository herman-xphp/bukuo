-- Report Templates
CREATE TABLE report_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    config TEXT NOT NULL, -- JSON
    created_by UUID NOT NULL REFERENCES users(id),
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Report Schedules
CREATE TABLE report_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES report_templates(id),
    frequency VARCHAR(20) NOT NULL,
    next_run_at TIMESTAMPTZ NOT NULL,
    email_recipients TEXT,
    format VARCHAR(10) NOT NULL DEFAULT 'PDF',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Generated Reports (Archive)
CREATE TABLE generated_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    template_id UUID REFERENCES report_templates(id),
    schedule_id UUID REFERENCES report_schedules(id),
    generated_at TIMESTAMPTZ DEFAULT NOW(),
    generated_by UUID REFERENCES users(id),
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    report_params TEXT -- JSON
);

CREATE INDEX idx_report_templates_company ON report_templates(company_id);
CREATE INDEX idx_generated_reports_company ON generated_reports(company_id);

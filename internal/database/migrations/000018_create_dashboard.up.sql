-- Dashboard Widgets
CREATE TABLE dashboard_widgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    datasource VARCHAR(30) NOT NULL,
    config TEXT, -- JSON
    position_x INT DEFAULT 0,
    position_y INT DEFAULT 0,
    width INT DEFAULT 1,
    height INT DEFAULT 1,
    is_visible BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Dashboard Preferences
CREATE TABLE user_dashboard_preferences (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'LIGHT',
    layout_type VARCHAR(20) DEFAULT 'DEFAULT',
    refresh_interval INT DEFAULT 300,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, company_id)
);

CREATE INDEX idx_dashboard_widgets_user ON dashboard_widgets(user_id, company_id);

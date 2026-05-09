CREATE TABLE project_accounts (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    label TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'bearer',
    scheme TEXT NOT NULL DEFAULT '',
    username TEXT NOT NULL DEFAULT '',
    password_enc TEXT NOT NULL DEFAULT '',
    api_key_enc TEXT NOT NULL DEFAULT '',
    api_key_header TEXT NOT NULL DEFAULT '',
    api_key_in TEXT NOT NULL DEFAULT 'header',
    token_enc TEXT NOT NULL DEFAULT '',
    refresh_token_enc TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMP,
    oauth_config_json TEXT NOT NULL DEFAULT '',
    totp_secret_enc TEXT NOT NULL DEFAULT '',
    totp_param TEXT NOT NULL DEFAULT '',
    login_endpoint_id TEXT NOT NULL DEFAULT '',
    login_body_template TEXT NOT NULL DEFAULT '',
    token_path TEXT NOT NULL DEFAULT '',
    user_json TEXT NOT NULL DEFAULT '',
    cookies_json TEXT NOT NULL DEFAULT '',
    headers_json TEXT NOT NULL DEFAULT '',
    is_default INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE (project_id, label)
);

CREATE INDEX idx_project_accounts_project ON project_accounts(project_id, sort_order);
CREATE INDEX idx_project_accounts_default ON project_accounts(project_id, is_default);

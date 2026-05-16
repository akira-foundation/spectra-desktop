CREATE TABLE license (
    id TEXT PRIMARY KEY CHECK (id = 'local'),
    customer_id TEXT NOT NULL DEFAULT '',
    customer_email TEXT NOT NULL DEFAULT '',
    customer_name TEXT NOT NULL DEFAULT '',
    access_token_enc TEXT NOT NULL DEFAULT '',
    plan TEXT NOT NULL DEFAULT '',
    cycle TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'inactive',
    valid_until TEXT NOT NULL DEFAULT '',
    activated_at TEXT NOT NULL DEFAULT '',
    last_verified_at TEXT NOT NULL DEFAULT '',
    license_key_id TEXT NOT NULL DEFAULT '',
    license_algorithm TEXT NOT NULL DEFAULT '',
    license_payload TEXT NOT NULL DEFAULT '',
    license_signature TEXT NOT NULL DEFAULT '',
    features_json TEXT NOT NULL DEFAULT '{}',
    device_id TEXT NOT NULL DEFAULT '',
    cancel_at_period_end INTEGER NOT NULL DEFAULT 0,
    cancel_at TEXT NOT NULL DEFAULT '',
    target_plan TEXT NOT NULL DEFAULT '',
    grace_period INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL
);

INSERT INTO license (id, status, updated_at) VALUES ('local', 'inactive', CURRENT_TIMESTAMP);

CREATE TABLE billing_usage_buffer (
    id TEXT PRIMARY KEY,
    feature TEXT NOT NULL,
    amount INTEGER NOT NULL DEFAULT 1,
    occurred_at TIMESTAMP NOT NULL,
    flushed INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_billing_usage_buffer_flushed
    ON billing_usage_buffer(flushed, occurred_at);

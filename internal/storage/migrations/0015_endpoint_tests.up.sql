CREATE TABLE endpoint_tests (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    endpoint_key TEXT NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL,
    json_path TEXT NOT NULL DEFAULT '',
    op TEXT NOT NULL DEFAULT '',
    expected TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX idx_endpoint_tests_lookup ON endpoint_tests(project_id, endpoint_key, sort_order);

ALTER TABLE request_history ADD COLUMN test_results_json TEXT NOT NULL DEFAULT '';

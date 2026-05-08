ALTER TABLE projects ADD COLUMN api_filter_mode TEXT NOT NULL DEFAULT 'auto';
ALTER TABLE projects ADD COLUMN api_filter_value TEXT NOT NULL DEFAULT '';

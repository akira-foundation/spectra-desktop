ALTER TABLE projects DROP COLUMN active_environment_id;
DROP INDEX IF EXISTS idx_environments_project;
DROP TABLE IF EXISTS environments;

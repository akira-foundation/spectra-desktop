ALTER TABLE endpoints ADD COLUMN auth_role_override TEXT NOT NULL DEFAULT '';
ALTER TABLE endpoints ADD COLUMN token_path_override TEXT NOT NULL DEFAULT '';

package domain

import "context"

const (
	SettingActiveProjectID = "active_project_id"
	SettingPHPBinaryPath   = "php_binary_path"
)

type SettingsRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

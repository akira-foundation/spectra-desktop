package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

type SettingsRepository struct {
	db *bun.DB
}

func NewSettingsRepository(db *bun.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

var _ domain.SettingsRepository = (*SettingsRepository)(nil)

func (r *SettingsRepository) Get(ctx context.Context, key string) (string, error) {
	var row model.Setting
	err := r.db.NewSelect().Model(&row).Where("key = ?", key).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return row.Value, nil
}

func (r *SettingsRepository) Set(ctx context.Context, key, value string) error {
	row := model.Setting{Key: key, Value: value, UpdatedAt: time.Now().UTC()}
	_, err := r.db.NewInsert().
		Model(&row).
		On("CONFLICT (key) DO UPDATE").
		Set("value = EXCLUDED.value").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *SettingsRepository) Delete(ctx context.Context, key string) error {
	_, err := r.db.NewDelete().Model((*model.Setting)(nil)).Where("key = ?", key).Exec(ctx)
	return err
}

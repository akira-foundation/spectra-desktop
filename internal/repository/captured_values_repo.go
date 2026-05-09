package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/repository/model"
)

type CapturedValuesRepository struct {
	db *bun.DB
}

func NewCapturedValuesRepository(db *bun.DB) *CapturedValuesRepository {
	return &CapturedValuesRepository{db: db}
}

type CapturedValueRow struct {
	Name        string
	Value       string
	EndpointKey string
	CapturedAt  time.Time
}

func (r *CapturedValuesRepository) ListByProject(ctx context.Context, projectID string) ([]CapturedValueRow, error) {
	var rows []model.CapturedValue
	if err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]CapturedValueRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, CapturedValueRow{
			Name:        row.Name,
			Value:       row.Value,
			EndpointKey: row.EndpointKey,
			CapturedAt:  row.CapturedAt,
		})
	}
	return out, nil
}

func (r *CapturedValuesRepository) Upsert(ctx context.Context, projectID, name, value, endpointKey string, at time.Time) error {
	row := &model.CapturedValue{
		ProjectID:   projectID,
		Name:        name,
		Value:       value,
		EndpointKey: endpointKey,
		CapturedAt:  at,
	}
	_, err := r.db.NewInsert().
		Model(row).
		On("CONFLICT (project_id, name) DO UPDATE").
		Set("value = EXCLUDED.value").
		Set("endpoint_key = EXCLUDED.endpoint_key").
		Set("captured_at = EXCLUDED.captured_at").
		Exec(ctx)
	return err
}

func (r *CapturedValuesRepository) DeleteByProject(ctx context.Context, projectID string) error {
	_, err := r.db.NewDelete().
		Model((*model.CapturedValue)(nil)).
		Where("project_id = ?", projectID).
		Exec(ctx)
	return err
}

func (r *CapturedValuesRepository) DeleteByEndpoint(ctx context.Context, projectID, endpointKey string, keep map[string]bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var rows []model.CapturedValue
	if err := tx.NewSelect().
		Model(&rows).
		Where("project_id = ? AND endpoint_key = ?", projectID, endpointKey).
		Scan(ctx); err != nil {
		return err
	}
	for _, row := range rows {
		if !keep[row.Name] {
			if _, err := tx.NewDelete().
				Model((*model.CapturedValue)(nil)).
				Where("project_id = ? AND name = ?", projectID, row.Name).
				Exec(ctx); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

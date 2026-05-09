package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/repository/model"
)

type DatasetRepository struct {
	db *bun.DB
}

func NewDatasetRepository(db *bun.DB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

func (r *DatasetRepository) Get(ctx context.Context, projectID, endpointKey string) (string, error) {
	row := new(model.EndpointDataset)
	err := r.db.NewSelect().
		Model(row).
		Where("project_id = ? AND endpoint_key = ?", projectID, endpointKey).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return "[]", nil
		}
		return "", err
	}
	if row.RowsJSON == "" {
		return "[]", nil
	}
	return row.RowsJSON, nil
}

func (r *DatasetRepository) Save(ctx context.Context, projectID, endpointKey, rowsJSON string) error {
	now := time.Now().UTC()
	row := &model.EndpointDataset{
		ProjectID:   projectID,
		EndpointKey: endpointKey,
		RowsJSON:    rowsJSON,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := r.db.NewInsert().
		Model(row).
		On("CONFLICT (project_id, endpoint_key) DO UPDATE").
		Set("rows_json = EXCLUDED.rows_json").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *DatasetRepository) Delete(ctx context.Context, projectID, endpointKey string) error {
	_, err := r.db.NewDelete().
		Model((*model.EndpointDataset)(nil)).
		Where("project_id = ? AND endpoint_key = ?", projectID, endpointKey).
		Exec(ctx)
	return err
}

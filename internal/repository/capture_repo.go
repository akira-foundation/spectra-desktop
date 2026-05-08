package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

type CaptureRepository struct {
	db *bun.DB
}

func NewCaptureRepository(db *bun.DB) *CaptureRepository {
	return &CaptureRepository{db: db}
}

var _ domain.CaptureRepository = (*CaptureRepository)(nil)

func (r *CaptureRepository) List(ctx context.Context, projectID, endpointKey string) ([]domain.EndpointCapture, error) {
	var rows []model.EndpointCapture
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		Where("endpoint_key = ?", endpointKey).
		OrderExpr("sort_order ASC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.EndpointCapture, 0, len(rows))
	for _, row := range rows {
		out = append(out, toCaptureDomain(row))
	}
	return out, nil
}

func (r *CaptureRepository) Replace(ctx context.Context, projectID, endpointKey string, captures []domain.EndpointCapture) error {
	now := time.Now().UTC()
	return r.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().
			Model((*model.EndpointCapture)(nil)).
			Where("project_id = ?", projectID).
			Where("endpoint_key = ?", endpointKey).
			Exec(ctx); err != nil {
			return err
		}
		if len(captures) == 0 {
			return nil
		}
		rows := make([]model.EndpointCapture, 0, len(captures))
		for i, c := range captures {
			id := c.ID
			if id == "" {
				id = uuid.NewString()
			}
			rows = append(rows, model.EndpointCapture{
				ID:          id,
				ProjectID:   projectID,
				EndpointKey: endpointKey,
				Name:        c.Name,
				Source:      c.Source,
				Path:        c.Path,
				SortOrder:   i,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}
		_, err := tx.NewInsert().Model(&rows).Exec(ctx)
		return err
	})
}

func (r *CaptureRepository) DeleteByEndpoint(ctx context.Context, projectID, endpointKey string) error {
	_, err := r.db.NewDelete().
		Model((*model.EndpointCapture)(nil)).
		Where("project_id = ?", projectID).
		Where("endpoint_key = ?", endpointKey).
		Exec(ctx)
	return err
}

func toCaptureDomain(row model.EndpointCapture) domain.EndpointCapture {
	return domain.EndpointCapture{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		EndpointKey: row.EndpointKey,
		Name:        row.Name,
		Source:      row.Source,
		Path:        row.Path,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

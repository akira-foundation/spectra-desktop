package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/repository/model"
)

type ScratchRepository struct {
	db *bun.DB
}

func NewScratchRepository(db *bun.DB) *ScratchRepository {
	return &ScratchRepository{db: db}
}

func (r *ScratchRepository) List(ctx context.Context, projectID string) ([]model.ScratchRequest, error) {
	var rows []model.ScratchRequest
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		Order("sort_order ASC", "created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *ScratchRepository) Save(ctx context.Context, req *model.ScratchRequest) error {
	now := time.Now().UTC()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	req.UpdatedAt = now
	_, err := r.db.NewInsert().
		Model(req).
		On("CONFLICT (id) DO UPDATE").
		Set("name = EXCLUDED.name").
		Set("method = EXCLUDED.method").
		Set("url = EXCLUDED.url").
		Set("headers_json = EXCLUDED.headers_json").
		Set("body = EXCLUDED.body").
		Set("response_json = EXCLUDED.response_json").
		Set("sort_order = EXCLUDED.sort_order").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *ScratchRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.ScratchRequest)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *ScratchRepository) DeleteByProject(ctx context.Context, projectID string) error {
	_, err := r.db.NewDelete().
		Model((*model.ScratchRequest)(nil)).
		Where("project_id = ?", projectID).
		Exec(ctx)
	return err
}

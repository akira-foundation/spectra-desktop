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

type MockRepository struct {
	db *bun.DB
}

func NewMockRepository(db *bun.DB) *MockRepository {
	return &MockRepository{db: db}
}

var _ domain.MockRepository = (*MockRepository)(nil)

func (r *MockRepository) List(ctx context.Context, projectID string) ([]domain.MockOverride, error) {
	var rows []model.MockOverride
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		Order("updated_at DESC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.MockOverride, len(rows))
	for i, row := range rows {
		out[i] = mockToDomain(row)
	}
	return out, nil
}

func (r *MockRepository) Get(ctx context.Context, projectID, endpointID string) (*domain.MockOverride, error) {
	var row model.MockOverride
	err := r.db.NewSelect().
		Model(&row).
		Where("project_id = ? AND endpoint_id = ?", projectID, endpointID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	out := mockToDomain(row)
	return &out, nil
}

func (r *MockRepository) Save(ctx context.Context, override domain.MockOverride) error {
	now := time.Now().UTC()
	if override.CreatedAt.IsZero() {
		override.CreatedAt = now
	}
	override.UpdatedAt = now
	row := mockToModel(override)
	_, err := r.db.NewInsert().
		Model(&row).
		On("CONFLICT (id) DO UPDATE").
		Set("enabled = EXCLUDED.enabled").
		Set("status = EXCLUDED.status").
		Set("latency_ms = EXCLUDED.latency_ms").
		Set("body = EXCLUDED.body").
		Set("headers_json = EXCLUDED.headers_json").
		Set("source = EXCLUDED.source").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *MockRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.MockOverride)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *MockRepository) DeleteByProject(ctx context.Context, projectID string) error {
	_, err := r.db.NewDelete().
		Model((*model.MockOverride)(nil)).
		Where("project_id = ?", projectID).
		Exec(ctx)
	return err
}

func mockToDomain(row model.MockOverride) domain.MockOverride {
	return domain.MockOverride{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		EndpointID:  row.EndpointID,
		Enabled:     row.Enabled,
		Status:      row.Status,
		LatencyMs:   row.LatencyMs,
		Body:        row.Body,
		HeadersJSON: row.HeadersJSON,
		Source:      domain.MockSource(row.Source),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func mockToModel(o domain.MockOverride) model.MockOverride {
	return model.MockOverride{
		ID:          o.ID,
		ProjectID:   o.ProjectID,
		EndpointID:  o.EndpointID,
		Enabled:     o.Enabled,
		Status:      o.Status,
		LatencyMs:   o.LatencyMs,
		Body:        o.Body,
		HeadersJSON: o.HeadersJSON,
		Source:      string(o.Source),
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
}

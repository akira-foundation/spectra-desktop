package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

type HistoryRepository struct {
	db *bun.DB
}

func NewHistoryRepository(db *bun.DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

var _ domain.HistoryRepository = (*HistoryRepository)(nil)

func (r *HistoryRepository) Save(ctx context.Context, entry domain.HistoryEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	row := model.RequestHistory{
		ID:              entry.ID,
		ProjectID:       entry.ProjectID,
		EndpointID:      entry.EndpointID,
		Method:          entry.Method,
		URL:             entry.URL,
		RequestHeaders:  entry.RequestHeaders,
		RequestBody:     entry.RequestBody,
		ResponseStatus:  entry.ResponseStatus,
		ResponseHeaders: entry.ResponseHeaders,
		ResponseBody:    entry.ResponseBody,
		DurationMs:      entry.DurationMs,
		SizeBytes:       entry.SizeBytes,
		Error:           entry.Error,
		TestResultsJSON: entry.TestResultsJSON,
		CreatedAt:       entry.CreatedAt,
	}
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)
	return err
}

func (r *HistoryRepository) List(ctx context.Context, projectID string, limit int) ([]domain.HistoryEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []model.RequestHistory
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		OrderExpr("created_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.HistoryEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, toHistoryDomain(row))
	}
	return out, nil
}

func (r *HistoryRepository) GetByID(ctx context.Context, id string) (*domain.HistoryEntry, error) {
	var row model.RequestHistory
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	entry := toHistoryDomain(row)
	return &entry, nil
}

func (r *HistoryRepository) LatestSuccessByEndpoint(ctx context.Context, projectID, endpointID string) (*domain.HistoryEntry, error) {
	var row model.RequestHistory
	err := r.db.NewSelect().
		Model(&row).
		Where("project_id = ?", projectID).
		Where("endpoint_id = ?", endpointID).
		Where("response_status >= 200 AND response_status < 300").
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	entry := toHistoryDomain(row)
	return &entry, nil
}

func (r *HistoryRepository) Clear(ctx context.Context, projectID string) error {
	_, err := r.db.NewDelete().
		Model((*model.RequestHistory)(nil)).
		Where("project_id = ?", projectID).
		Exec(ctx)
	return err
}

func (r *HistoryRepository) TrimOldest(ctx context.Context, projectID string, keep int) error {
	if keep <= 0 {
		return nil
	}
	_, err := r.db.NewDelete().
		Model((*model.RequestHistory)(nil)).
		Where("project_id = ?", projectID).
		Where("id NOT IN (?)",
			r.db.NewSelect().
				Model((*model.RequestHistory)(nil)).
				Column("id").
				Where("project_id = ?", projectID).
				OrderExpr("created_at DESC").
				Limit(keep),
		).
		Exec(ctx)
	return err
}

func toHistoryDomain(row model.RequestHistory) domain.HistoryEntry {
	return domain.HistoryEntry{
		ID:              row.ID,
		ProjectID:       row.ProjectID,
		EndpointID:      row.EndpointID,
		Method:          row.Method,
		URL:             row.URL,
		RequestHeaders:  row.RequestHeaders,
		RequestBody:     row.RequestBody,
		ResponseStatus:  row.ResponseStatus,
		ResponseHeaders: row.ResponseHeaders,
		ResponseBody:    row.ResponseBody,
		DurationMs:      row.DurationMs,
		SizeBytes:       row.SizeBytes,
		Error:           row.Error,
		TestResultsJSON: row.TestResultsJSON,
		CreatedAt:       row.CreatedAt,
	}
}

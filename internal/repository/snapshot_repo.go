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

type SnapshotRepository struct {
	db *bun.DB
}

func NewSnapshotRepository(db *bun.DB) *SnapshotRepository {
	return &SnapshotRepository{db: db}
}

var _ domain.SnapshotRepository = (*SnapshotRepository)(nil)

func (r *SnapshotRepository) Save(ctx context.Context, s domain.EndpointSnapshot) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	row := model.EndpointSnapshot{
		ID:            s.ID,
		ProjectID:     s.ProjectID,
		Hash:          s.Hash,
		PayloadJSON:   s.PayloadJSON,
		EndpointCount: s.EndpointCount,
		ScannedAt:     s.ScannedAt,
		CreatedAt:     s.CreatedAt,
	}
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)
	return err
}

func (r *SnapshotRepository) List(ctx context.Context, projectID string, limit int) ([]domain.EndpointSnapshot, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []model.EndpointSnapshot
	err := r.db.NewSelect().
		Model(&rows).
		Column("id", "project_id", "hash", "endpoint_count", "scanned_at", "created_at").
		Where("project_id = ?", projectID).
		OrderExpr("scanned_at DESC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.EndpointSnapshot, 0, len(rows))
	for _, row := range rows {
		out = append(out, toSnapshotDomain(row))
	}
	return out, nil
}

func (r *SnapshotRepository) GetByID(ctx context.Context, id string) (*domain.EndpointSnapshot, error) {
	var row model.EndpointSnapshot
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s := toSnapshotDomain(row)
	return &s, nil
}

func (r *SnapshotRepository) Latest(ctx context.Context, projectID string) (*domain.EndpointSnapshot, error) {
	var row model.EndpointSnapshot
	err := r.db.NewSelect().
		Model(&row).
		Where("project_id = ?", projectID).
		OrderExpr("scanned_at DESC").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s := toSnapshotDomain(row)
	return &s, nil
}

func (r *SnapshotRepository) Predecessor(ctx context.Context, projectID string, scannedAt time.Time) (*domain.EndpointSnapshot, error) {
	var row model.EndpointSnapshot
	err := r.db.NewSelect().
		Model(&row).
		Where("project_id = ?", projectID).
		Where("scanned_at < ?", scannedAt).
		OrderExpr("scanned_at DESC").
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s := toSnapshotDomain(row)
	return &s, nil
}

func (r *SnapshotRepository) TrimOldest(ctx context.Context, projectID string, keep int) error {
	if keep <= 0 {
		return nil
	}
	_, err := r.db.NewDelete().
		Model((*model.EndpointSnapshot)(nil)).
		Where("project_id = ?", projectID).
		Where("id NOT IN (?)",
			r.db.NewSelect().
				Model((*model.EndpointSnapshot)(nil)).
				Column("id").
				Where("project_id = ?", projectID).
				OrderExpr("scanned_at DESC").
				Limit(keep),
		).
		Exec(ctx)
	return err
}

func toSnapshotDomain(row model.EndpointSnapshot) domain.EndpointSnapshot {
	return domain.EndpointSnapshot{
		ID:            row.ID,
		ProjectID:     row.ProjectID,
		Hash:          row.Hash,
		PayloadJSON:   row.PayloadJSON,
		EndpointCount: row.EndpointCount,
		ScannedAt:     row.ScannedAt,
		CreatedAt:     row.CreatedAt,
	}
}

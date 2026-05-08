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

type TestRepository struct {
	db *bun.DB
}

func NewTestRepository(db *bun.DB) *TestRepository {
	return &TestRepository{db: db}
}

var _ domain.TestRepository = (*TestRepository)(nil)

func (r *TestRepository) List(ctx context.Context, projectID, endpointKey string) ([]domain.EndpointTest, error) {
	var rows []model.EndpointTest
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		Where("endpoint_key = ?", endpointKey).
		OrderExpr("sort_order ASC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.EndpointTest, 0, len(rows))
	for _, row := range rows {
		out = append(out, toTestDomain(row))
	}
	return out, nil
}

func (r *TestRepository) Replace(ctx context.Context, projectID, endpointKey string, tests []domain.EndpointTest) error {
	now := time.Now().UTC()
	return r.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().
			Model((*model.EndpointTest)(nil)).
			Where("project_id = ?", projectID).
			Where("endpoint_key = ?", endpointKey).
			Exec(ctx); err != nil {
			return err
		}
		if len(tests) == 0 {
			return nil
		}
		rows := make([]model.EndpointTest, 0, len(tests))
		for i, t := range tests {
			id := t.ID
			if id == "" {
				id = uuid.NewString()
			}
			rows = append(rows, model.EndpointTest{
				ID:          id,
				ProjectID:   projectID,
				EndpointKey: endpointKey,
				Name:        t.Name,
				Kind:        t.Kind,
				JSONPath:    t.JSONPath,
				Op:          t.Op,
				Expected:    t.Expected,
				SortOrder:   i,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}
		_, err := tx.NewInsert().Model(&rows).Exec(ctx)
		return err
	})
}

func (r *TestRepository) DeleteByEndpoint(ctx context.Context, projectID, endpointKey string) error {
	_, err := r.db.NewDelete().
		Model((*model.EndpointTest)(nil)).
		Where("project_id = ?", projectID).
		Where("endpoint_key = ?", endpointKey).
		Exec(ctx)
	return err
}

func toTestDomain(row model.EndpointTest) domain.EndpointTest {
	return domain.EndpointTest{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		EndpointKey: row.EndpointKey,
		Name:        row.Name,
		Kind:        row.Kind,
		JSONPath:    row.JSONPath,
		Op:          row.Op,
		Expected:    row.Expected,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

type EnvironmentRepository struct {
	db *bun.DB
}

func NewEnvironmentRepository(db *bun.DB) *EnvironmentRepository {
	return &EnvironmentRepository{db: db}
}

var _ domain.EnvironmentRepository = (*EnvironmentRepository)(nil)

func (r *EnvironmentRepository) List(ctx context.Context, projectID string) ([]domain.Environment, error) {
	var rows []model.Environment
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		OrderExpr("sort_order ASC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Environment, 0, len(rows))
	for _, row := range rows {
		out = append(out, toEnvDomain(row))
	}
	return out, nil
}

func (r *EnvironmentRepository) GetByID(ctx context.Context, id string) (*domain.Environment, error) {
	var row model.Environment
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	env := toEnvDomain(row)
	return &env, nil
}

func (r *EnvironmentRepository) Save(ctx context.Context, input domain.EnvironmentInput) (*domain.Environment, error) {
	now := time.Now().UTC()
	varsJSON, err := json.Marshal(input.Vars)
	if err != nil {
		return nil, err
	}
	if input.ID == "" {
		row := model.Environment{
			ID:        uuid.NewString(),
			ProjectID: input.ProjectID,
			Name:      input.Name,
			VarsJSON:  string(varsJSON),
			SortOrder: input.SortOrder,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if _, err := r.db.NewInsert().Model(&row).Exec(ctx); err != nil {
			return nil, err
		}
		env := toEnvDomain(row)
		return &env, nil
	}
	if _, err := r.db.NewUpdate().
		Model((*model.Environment)(nil)).
		Set("name = ?", input.Name).
		Set("vars_json = ?", string(varsJSON)).
		Set("sort_order = ?", input.SortOrder).
		Set("updated_at = ?", now).
		Where("id = ?", input.ID).
		Exec(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, input.ID)
}

func (r *EnvironmentRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.Environment)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func toEnvDomain(row model.Environment) domain.Environment {
	vars := map[string]string{}
	if row.VarsJSON != "" {
		_ = json.Unmarshal([]byte(row.VarsJSON), &vars)
	}
	return domain.Environment{
		ID:        row.ID,
		ProjectID: row.ProjectID,
		Name:      row.Name,
		Vars:      vars,
		SortOrder: row.SortOrder,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

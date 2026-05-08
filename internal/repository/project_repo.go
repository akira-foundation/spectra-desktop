package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

type ProjectRepository struct {
	db *bun.DB
}

func NewProjectRepository(db *bun.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

var _ domain.ProjectRepository = (*ProjectRepository)(nil)

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	var rows []model.Project
	if err := r.db.NewSelect().Model(&rows).OrderExpr("created_at ASC").Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToDomain())
	}
	return out, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var row model.Project
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	d := row.ToDomain()
	return &d, nil
}

func (r *ProjectRepository) GetByPath(ctx context.Context, path string) (*domain.Project, error) {
	var row model.Project
	err := r.db.NewSelect().Model(&row).Where("path = ?", path).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}
	d := row.ToDomain()
	return &d, nil
}

func (r *ProjectRepository) Save(ctx context.Context, input domain.ProjectInput) (*domain.Project, error) {
	now := time.Now().UTC()
	existing, err := r.GetByPath(ctx, input.Path)
	if err != nil && !errors.Is(err, domain.ErrProjectNotFound) {
		return nil, err
	}

	mode, value := normalizeAPIFilter(input.APIFilterMode, input.APIFilterValue)

	if existing != nil {
		existing.Name = strings.TrimSpace(input.Name)
		existing.Framework = input.Framework
		existing.FrameworkVersion = input.FrameworkVersion
		existing.APIFilterMode = mode
		existing.APIFilterValue = value
		if strings.TrimSpace(input.BaseURL) != "" {
			existing.BaseURL = strings.TrimSpace(input.BaseURL)
		}
		existing.UpdatedAt = now

		row := model.FromDomain(*existing)
		if _, err := r.db.NewUpdate().Model(&row).WherePK().Exec(ctx); err != nil {
			return nil, err
		}
		d := row.ToDomain()
		return &d, nil
	}

	id := input.ID
	if strings.TrimSpace(id) == "" {
		id = uuid.NewString()
	}
	row := model.Project{
		ID:               id,
		Name:             strings.TrimSpace(input.Name),
		Path:             input.Path,
		Framework:        input.Framework,
		FrameworkVersion: input.FrameworkVersion,
		Status:           string(domain.ProjectStatusDisconnected),
		APIFilterMode:    mode,
		APIFilterValue:   value,
		BaseURL:          strings.TrimSpace(input.BaseURL),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if _, err := r.db.NewInsert().Model(&row).Exec(ctx); err != nil {
		return nil, err
	}
	d := row.ToDomain()
	return &d, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.NewDelete().Model((*model.Project)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) UpdateStatus(ctx context.Context, id string, status domain.ProjectStatus) error {
	res, err := r.db.NewUpdate().
		Model((*model.Project)(nil)).
		Set("status = ?", string(status)).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) MarkSynced(ctx context.Context, id string) error {
	now := time.Now().UTC()
	res, err := r.db.NewUpdate().
		Model((*model.Project)(nil)).
		Set("last_synced_at = ?", now).
		Set("updated_at = ?", now).
		Set("status = ?", string(domain.ProjectStatusConnected)).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func normalizeAPIFilter(mode, value string) (string, string) {
	mode = strings.TrimSpace(strings.ToLower(mode))
	value = strings.TrimSpace(value)
	switch mode {
	case domain.APIFilterModeMiddleware, domain.APIFilterModePrefix, domain.APIFilterModeAll:
		// ok
	default:
		mode = domain.APIFilterModeAuto
	}
	if mode == domain.APIFilterModeAll {
		value = ""
	}
	return mode, value
}

func (r *ProjectRepository) UpdateBaseURL(ctx context.Context, id, baseURL string) error {
	res, err := r.db.NewUpdate().
		Model((*model.Project)(nil)).
		Set("base_url = ?", strings.TrimSpace(baseURL)).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) UpdateActiveEnvironment(ctx context.Context, id, envID string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Project)(nil)).
		Set("active_environment_id = ?", envID).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (r *ProjectRepository) UpdateAuthRoutes(ctx context.Context, id, loginID, logoutID, tokenPath string) error {
	res, err := r.db.NewUpdate().
		Model((*model.Project)(nil)).
		Set("login_endpoint_id = ?", loginID).
		Set("login_token_path = ?", tokenPath).
		Set("logout_endpoint_id = ?", logoutID).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

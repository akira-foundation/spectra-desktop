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

type AuthRepository struct {
	db *bun.DB
}

func NewAuthRepository(db *bun.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

var _ domain.AuthRepository = (*AuthRepository)(nil)

func (r *AuthRepository) Get(ctx context.Context, projectID string) (*domain.ProjectAuth, error) {
	var row model.ProjectAuth
	err := r.db.NewSelect().Model(&row).Where("project_id = ?", projectID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return modelToDomainAuth(row), nil
}

func (r *AuthRepository) Save(ctx context.Context, auth domain.ProjectAuth) error {
	now := time.Now().UTC()
	if auth.CapturedAt.IsZero() {
		auth.CapturedAt = now
	}
	auth.UpdatedAt = now
	row := model.ProjectAuth{
		ProjectID:            auth.ProjectID,
		Scheme:               auth.Scheme,
		Token:                auth.Token,
		TokenPath:            auth.TokenPath,
		UserJSON:             auth.UserJSON,
		CookiesJSON:          auth.CookiesJSON,
		HeadersJSON:          auth.HeadersJSON,
		ExpiresAt:            auth.ExpiresAt,
		CapturedFromEndpoint: auth.CapturedFromEndpoint,
		CapturedAt:           auth.CapturedAt,
		UpdatedAt:            auth.UpdatedAt,
	}
	_, err := r.db.NewInsert().
		Model(&row).
		On("CONFLICT (project_id) DO UPDATE").
		Set("scheme = EXCLUDED.scheme").
		Set("token = EXCLUDED.token").
		Set("token_path = EXCLUDED.token_path").
		Set("user_json = EXCLUDED.user_json").
		Set("cookies_json = EXCLUDED.cookies_json").
		Set("headers_json = EXCLUDED.headers_json").
		Set("expires_at = EXCLUDED.expires_at").
		Set("captured_from_endpoint = EXCLUDED.captured_from_endpoint").
		Set("captured_at = EXCLUDED.captured_at").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *AuthRepository) Clear(ctx context.Context, projectID string) error {
	_, err := r.db.NewDelete().Model((*model.ProjectAuth)(nil)).Where("project_id = ?", projectID).Exec(ctx)
	return err
}

func modelToDomainAuth(row model.ProjectAuth) *domain.ProjectAuth {
	return &domain.ProjectAuth{
		ProjectID:            row.ProjectID,
		Scheme:               row.Scheme,
		Token:                row.Token,
		TokenPath:            row.TokenPath,
		UserJSON:             row.UserJSON,
		CookiesJSON:          row.CookiesJSON,
		HeadersJSON:          row.HeadersJSON,
		ExpiresAt:            row.ExpiresAt,
		CapturedFromEndpoint: row.CapturedFromEndpoint,
		CapturedAt:           row.CapturedAt,
		UpdatedAt:            row.UpdatedAt,
	}
}

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

type AccountRepository struct {
	db *bun.DB
}

func NewAccountRepository(db *bun.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

var _ domain.AccountRepository = (*AccountRepository)(nil)

func (r *AccountRepository) List(ctx context.Context, projectID string) ([]domain.ProjectAccount, error) {
	var rows []model.ProjectAccount
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		Order("sort_order ASC", "created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ProjectAccount, len(rows))
	for i, row := range rows {
		out[i] = *modelToAccount(row)
	}
	return out, nil
}

func (r *AccountRepository) Get(ctx context.Context, id string) (*domain.ProjectAccount, error) {
	var row model.ProjectAccount
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return modelToAccount(row), nil
}

func (r *AccountRepository) GetDefault(ctx context.Context, projectID string) (*domain.ProjectAccount, error) {
	var row model.ProjectAccount
	err := r.db.NewSelect().
		Model(&row).
		Where("project_id = ? AND is_default = TRUE", projectID).
		Limit(1).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		// Fallback: first account by sort order
		err = r.db.NewSelect().
			Model(&row).
			Where("project_id = ?", projectID).
			Order("sort_order ASC", "created_at ASC").
			Limit(1).
			Scan(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return modelToAccount(row), nil
}

func (r *AccountRepository) Save(ctx context.Context, acc domain.ProjectAccount) error {
	now := time.Now().UTC()
	if acc.CreatedAt.IsZero() {
		acc.CreatedAt = now
	}
	acc.UpdatedAt = now
	row := accountToModel(acc)
	_, err := r.db.NewInsert().
		Model(&row).
		On("CONFLICT (id) DO UPDATE").
		Set("label = EXCLUDED.label").
		Set("kind = EXCLUDED.kind").
		Set("scheme = EXCLUDED.scheme").
		Set("username = EXCLUDED.username").
		Set("password_enc = EXCLUDED.password_enc").
		Set("api_key_enc = EXCLUDED.api_key_enc").
		Set("api_key_header = EXCLUDED.api_key_header").
		Set("api_key_in = EXCLUDED.api_key_in").
		Set("token_enc = EXCLUDED.token_enc").
		Set("refresh_token_enc = EXCLUDED.refresh_token_enc").
		Set("expires_at = EXCLUDED.expires_at").
		Set("oauth_config_json = EXCLUDED.oauth_config_json").
		Set("totp_secret_enc = EXCLUDED.totp_secret_enc").
		Set("totp_param = EXCLUDED.totp_param").
		Set("login_endpoint_id = EXCLUDED.login_endpoint_id").
		Set("login_body_template = EXCLUDED.login_body_template").
		Set("token_path = EXCLUDED.token_path").
		Set("user_json = EXCLUDED.user_json").
		Set("cookies_json = EXCLUDED.cookies_json").
		Set("headers_json = EXCLUDED.headers_json").
		Set("is_default = EXCLUDED.is_default").
		Set("sort_order = EXCLUDED.sort_order").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *AccountRepository) SetDefault(ctx context.Context, projectID, accountID string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewUpdate().
			Model((*model.ProjectAccount)(nil)).
			Set("is_default = FALSE").
			Where("project_id = ?", projectID).
			Exec(ctx); err != nil {
			return err
		}
		_, err := tx.NewUpdate().
			Model((*model.ProjectAccount)(nil)).
			Set("is_default = TRUE").
			Set("updated_at = ?", time.Now().UTC()).
			Where("id = ? AND project_id = ?", accountID, projectID).
			Exec(ctx)
		return err
	})
}

func (r *AccountRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.ProjectAccount)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func modelToAccount(row model.ProjectAccount) *domain.ProjectAccount {
	return &domain.ProjectAccount{
		ID:                row.ID,
		ProjectID:         row.ProjectID,
		Label:             row.Label,
		Kind:              domain.AccountKind(row.Kind),
		Scheme:            row.Scheme,
		Username:          row.Username,
		PasswordEnc:       row.PasswordEnc,
		APIKeyEnc:         row.APIKeyEnc,
		APIKeyHeader:      row.APIKeyHeader,
		APIKeyIn:          domain.APIKeyLocation(row.APIKeyIn),
		TokenEnc:          row.TokenEnc,
		RefreshTokenEnc:   row.RefreshTokenEnc,
		ExpiresAt:         row.ExpiresAt,
		OAuthConfigJSON:   row.OAuthConfigJSON,
		TOTPSecretEnc:     row.TOTPSecretEnc,
		TOTPParam:         row.TOTPParam,
		LoginEndpointID:   row.LoginEndpointID,
		LoginBodyTemplate: row.LoginBodyTemplate,
		TokenPath:         row.TokenPath,
		UserJSON:          row.UserJSON,
		CookiesJSON:       row.CookiesJSON,
		HeadersJSON:       row.HeadersJSON,
		IsDefault:         row.IsDefault,
		SortOrder:         row.SortOrder,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

func accountToModel(acc domain.ProjectAccount) model.ProjectAccount {
	return model.ProjectAccount{
		ID:                acc.ID,
		ProjectID:         acc.ProjectID,
		Label:             acc.Label,
		Kind:              string(acc.Kind),
		Scheme:            acc.Scheme,
		Username:          acc.Username,
		PasswordEnc:       acc.PasswordEnc,
		APIKeyEnc:         acc.APIKeyEnc,
		APIKeyHeader:      acc.APIKeyHeader,
		APIKeyIn:          string(acc.APIKeyIn),
		TokenEnc:          acc.TokenEnc,
		RefreshTokenEnc:   acc.RefreshTokenEnc,
		ExpiresAt:         acc.ExpiresAt,
		OAuthConfigJSON:   acc.OAuthConfigJSON,
		TOTPSecretEnc:     acc.TOTPSecretEnc,
		TOTPParam:         acc.TOTPParam,
		LoginEndpointID:   acc.LoginEndpointID,
		LoginBodyTemplate: acc.LoginBodyTemplate,
		TokenPath:         acc.TokenPath,
		UserJSON:          acc.UserJSON,
		CookiesJSON:       acc.CookiesJSON,
		HeadersJSON:       acc.HeadersJSON,
		IsDefault:         acc.IsDefault,
		SortOrder:         acc.SortOrder,
		CreatedAt:         acc.CreatedAt,
		UpdatedAt:         acc.UpdatedAt,
	}
}

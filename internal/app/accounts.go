package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	authpkg "spectra-desktop/internal/auth"
	"spectra-desktop/internal/domain"
)

// ProjectAccountDTO is the wire format exposed to the frontend. Secret
// fields are never returned in clear text — the boolean *Set flags let the
// UI know whether a value exists without leaking it.
type ProjectAccountDTO struct {
	ID                string            `json:"id"`
	ProjectID         string            `json:"projectID"`
	Label             string            `json:"label"`
	Kind              string            `json:"kind"`
	Scheme            string            `json:"scheme"`
	Username          string            `json:"username"`
	HasPassword       bool              `json:"hasPassword"`
	HasAPIKey         bool              `json:"hasApiKey"`
	APIKeyHeader      string            `json:"apiKeyHeader"`
	APIKeyIn          string            `json:"apiKeyIn"`
	HasToken          bool              `json:"hasToken"`
	TokenPreview      string            `json:"tokenPreview,omitempty"`
	HasRefreshToken   bool              `json:"hasRefreshToken"`
	ExpiresAt         *time.Time        `json:"expiresAt,omitempty"`
	OAuth             *OAuthConfigDTO   `json:"oauth,omitempty"`
	HasTOTP           bool              `json:"hasTotp"`
	TOTPParam         string            `json:"totpParam"`
	LoginEndpointID   string            `json:"loginEndpointId"`
	LoginBodyTemplate string            `json:"loginBodyTemplate"`
	TokenPath         string            `json:"tokenPath"`
	User              map[string]any    `json:"user,omitempty"`
	HasCookies        bool              `json:"hasCookies"`
	ExtraHeaders      map[string]string `json:"extraHeaders,omitempty"`
	IsDefault         bool              `json:"isDefault"`
	SortOrder         int               `json:"sortOrder"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
}

// OAuthConfigDTO mirrors auth.OAuth2Config without secrets.
type OAuthConfigDTO struct {
	GrantType    string   `json:"grantType"`
	TokenURL     string   `json:"tokenUrl"`
	ClientID     string   `json:"clientId"`
	HasSecret    bool     `json:"hasSecret"`
	Scopes       []string `json:"scopes,omitempty"`
	Audience     string   `json:"audience,omitempty"`
	Username     string   `json:"username,omitempty"`
}

// SaveAccountInput accepts secrets in clear text — they are encrypted
// before being persisted. To leave a secret unchanged on update, omit
// the field; explicit empty string means "clear it".
type SaveAccountInput struct {
	ID                string                 `json:"id,omitempty"`
	ProjectID         string                 `json:"projectID"`
	Label             string                 `json:"label"`
	Kind              string                 `json:"kind"`
	Scheme            string                 `json:"scheme,omitempty"`
	Username          string                 `json:"username,omitempty"`
	Password          *string                `json:"password,omitempty"`
	APIKey            *string                `json:"apiKey,omitempty"`
	APIKeyHeader      string                 `json:"apiKeyHeader,omitempty"`
	APIKeyIn          string                 `json:"apiKeyIn,omitempty"`
	Token             *string                `json:"token,omitempty"`
	RefreshToken      *string                `json:"refreshToken,omitempty"`
	OAuth             *SaveOAuthInput        `json:"oauth,omitempty"`
	TOTPSecret        *string                `json:"totpSecret,omitempty"`
	TOTPParam         string                 `json:"totpParam,omitempty"`
	LoginEndpointID   string                 `json:"loginEndpointId,omitempty"`
	LoginBodyTemplate string                 `json:"loginBodyTemplate,omitempty"`
	TokenPath         string                 `json:"tokenPath,omitempty"`
	ExtraHeaders      map[string]string      `json:"extraHeaders,omitempty"`
	IsDefault         bool                   `json:"isDefault,omitempty"`
	SortOrder         int                    `json:"sortOrder,omitempty"`
	ExpiresAt         *time.Time             `json:"expiresAt,omitempty"`
	User              map[string]any         `json:"user,omitempty"`
}

// SaveOAuthInput is OAuthConfigDTO with the secret in clear text.
type SaveOAuthInput struct {
	GrantType    string   `json:"grantType"`
	TokenURL     string   `json:"tokenUrl"`
	ClientID     string   `json:"clientId"`
	ClientSecret *string  `json:"clientSecret,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Audience     string   `json:"audience,omitempty"`
	Username     string   `json:"username,omitempty"`
}

// ListProjectAccounts returns all accounts for the project. On first call
// it lazily migrates the legacy project_auth row into a "Default" account.
func (a *App) ListProjectAccounts(projectID string) ([]ProjectAccountDTO, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project id required")
	}
	if a.accounts == nil {
		return []ProjectAccountDTO{}, nil
	}
	rows, err := a.accounts.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		if migrated, err := a.migrateLegacyAuth(projectID); err == nil && migrated {
			rows, err = a.accounts.List(a.ctx, projectID)
			if err != nil {
				return nil, err
			}
		}
	}
	out := make([]ProjectAccountDTO, len(rows))
	for i, row := range rows {
		out[i] = a.accountToDTO(row)
	}
	return out, nil
}

// SaveProjectAccount creates or updates an account. Secrets are encrypted
// before being written. ID is generated when missing.
func (a *App) SaveProjectAccount(input SaveAccountInput) (*ProjectAccountDTO, error) {
	if a.accounts == nil || a.vault == nil {
		return nil, fmt.Errorf("accounts subsystem not initialized")
	}
	if input.ProjectID == "" {
		return nil, fmt.Errorf("project id required")
	}
	if strings.TrimSpace(input.Label) == "" {
		return nil, fmt.Errorf("label required")
	}
	kind := strings.TrimSpace(input.Kind)
	if kind == "" {
		kind = string(domain.AccountKindBearer)
	}

	var existing *domain.ProjectAccount
	if input.ID != "" {
		current, err := a.accounts.Get(a.ctx, input.ID)
		if err != nil {
			return nil, err
		}
		existing = current
	}

	acc := domain.ProjectAccount{
		ID:                strings.TrimSpace(input.ID),
		ProjectID:         input.ProjectID,
		Label:             strings.TrimSpace(input.Label),
		Kind:              domain.AccountKind(kind),
		Scheme:            strings.TrimSpace(input.Scheme),
		Username:          strings.TrimSpace(input.Username),
		APIKeyHeader:      strings.TrimSpace(input.APIKeyHeader),
		APIKeyIn:          domain.APIKeyLocation(strings.TrimSpace(input.APIKeyIn)),
		TOTPParam:         strings.TrimSpace(input.TOTPParam),
		LoginEndpointID:   strings.TrimSpace(input.LoginEndpointID),
		LoginBodyTemplate: input.LoginBodyTemplate,
		TokenPath:         strings.TrimSpace(input.TokenPath),
		IsDefault:         input.IsDefault,
		SortOrder:         input.SortOrder,
		ExpiresAt:         input.ExpiresAt,
	}
	if acc.ID == "" {
		acc.ID = uuid.NewString()
	}
	if acc.APIKeyIn == "" {
		acc.APIKeyIn = domain.APIKeyInHeader
	}
	if existing != nil {
		acc.CreatedAt = existing.CreatedAt
		acc.PasswordEnc = existing.PasswordEnc
		acc.APIKeyEnc = existing.APIKeyEnc
		acc.TokenEnc = existing.TokenEnc
		acc.RefreshTokenEnc = existing.RefreshTokenEnc
		acc.OAuthConfigJSON = existing.OAuthConfigJSON
		acc.TOTPSecretEnc = existing.TOTPSecretEnc
		acc.UserJSON = existing.UserJSON
		acc.CookiesJSON = existing.CookiesJSON
		acc.HeadersJSON = existing.HeadersJSON
	}

	if input.Password != nil {
		enc, err := a.vault.Encrypt(*input.Password)
		if err != nil {
			return nil, err
		}
		acc.PasswordEnc = enc
	}
	if input.APIKey != nil {
		enc, err := a.vault.Encrypt(*input.APIKey)
		if err != nil {
			return nil, err
		}
		acc.APIKeyEnc = enc
	}
	if input.Token != nil {
		enc, err := a.vault.Encrypt(*input.Token)
		if err != nil {
			return nil, err
		}
		acc.TokenEnc = enc
	}
	if input.RefreshToken != nil {
		enc, err := a.vault.Encrypt(*input.RefreshToken)
		if err != nil {
			return nil, err
		}
		acc.RefreshTokenEnc = enc
	}
	if input.TOTPSecret != nil {
		enc, err := a.vault.Encrypt(*input.TOTPSecret)
		if err != nil {
			return nil, err
		}
		acc.TOTPSecretEnc = enc
	}
	if input.OAuth != nil {
		cfg := authpkg.OAuth2Config{
			GrantType: input.OAuth.GrantType,
			TokenURL:  input.OAuth.TokenURL,
			ClientID:  input.OAuth.ClientID,
			Scopes:    input.OAuth.Scopes,
			Audience:  input.OAuth.Audience,
			Username:  input.OAuth.Username,
		}
		if input.OAuth.ClientSecret != nil {
			enc, err := a.vault.Encrypt(*input.OAuth.ClientSecret)
			if err != nil {
				return nil, err
			}
			// Stored alongside the JSON config; we keep it inside JSON so
			// rotating a secret is a single column write.
			cfg.ClientSecret = enc
		} else if existing != nil && existing.OAuthConfigJSON != "" {
			var prior authpkg.OAuth2Config
			if err := json.Unmarshal([]byte(existing.OAuthConfigJSON), &prior); err == nil {
				cfg.ClientSecret = prior.ClientSecret
			}
		}
		raw, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		acc.OAuthConfigJSON = string(raw)
	}
	if input.User != nil {
		raw, err := json.Marshal(input.User)
		if err != nil {
			return nil, err
		}
		acc.UserJSON = string(raw)
	}
	if input.ExtraHeaders != nil {
		raw, err := json.Marshal(input.ExtraHeaders)
		if err != nil {
			return nil, err
		}
		acc.HeadersJSON = string(raw)
	}

	if err := a.accounts.Save(a.ctx, acc); err != nil {
		return nil, err
	}
	if input.IsDefault {
		if err := a.accounts.SetDefault(a.ctx, input.ProjectID, acc.ID); err != nil {
			return nil, err
		}
	}
	saved, err := a.accounts.Get(a.ctx, acc.ID)
	if err != nil || saved == nil {
		return nil, err
	}
	dto := a.accountToDTO(*saved)
	return &dto, nil
}

// SetDefaultProjectAccount marks an account as the project's active default.
func (a *App) SetDefaultProjectAccount(projectID, accountID string) error {
	if a.accounts == nil {
		return fmt.Errorf("accounts subsystem not initialized")
	}
	return a.accounts.SetDefault(a.ctx, projectID, accountID)
}

// DeleteProjectAccount removes an account.
func (a *App) DeleteProjectAccount(accountID string) error {
	if a.accounts == nil {
		return fmt.Errorf("accounts subsystem not initialized")
	}
	return a.accounts.Delete(a.ctx, accountID)
}

// AccountSecretsDTO is returned to the UI only when the user explicitly
// asks to auto-fill credentials. Never serialized in lists.
type AccountSecretsDTO struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	APIKey   string `json:"apiKey,omitempty"`
}

// GetAccountSecrets decrypts and returns the account's credentials. Used by
// the Inspector to auto-fill login-endpoint bodies.
func (a *App) GetAccountSecrets(accountID string) (*AccountSecretsDTO, error) {
	if a.accounts == nil || a.vault == nil {
		return nil, fmt.Errorf("accounts subsystem not initialized")
	}
	acc, err := a.accounts.Get(a.ctx, accountID)
	if err != nil || acc == nil {
		return nil, err
	}
	out := &AccountSecretsDTO{Username: acc.Username}
	if acc.PasswordEnc != "" {
		if pwd, err := a.vault.Decrypt(acc.PasswordEnc); err == nil {
			out.Password = pwd
		}
	}
	if acc.TokenEnc != "" {
		if tok, err := a.vault.Decrypt(acc.TokenEnc); err == nil {
			out.Token = tok
		}
	}
	if acc.APIKeyEnc != "" {
		if k, err := a.vault.Decrypt(acc.APIKeyEnc); err == nil {
			out.APIKey = k
		}
	}
	return out, nil
}

// GenerateAccountTOTP returns the current 6-digit code for the account's
// TOTP secret. Useful for showing the code in the UI.
func (a *App) GenerateAccountTOTP(accountID string) (string, error) {
	if a.accounts == nil || a.vault == nil {
		return "", fmt.Errorf("accounts subsystem not initialized")
	}
	acc, err := a.accounts.Get(a.ctx, accountID)
	if err != nil || acc == nil {
		return "", err
	}
	secret, err := a.vault.Decrypt(acc.TOTPSecretEnc)
	if err != nil {
		return "", err
	}
	return authpkg.GenerateTOTP(secret)
}

// RefreshAccountToken forces an OAuth2 token refresh for the account.
func (a *App) RefreshAccountToken(accountID string) (*ProjectAccountDTO, error) {
	if a.accounts == nil || a.authResolve == nil {
		return nil, fmt.Errorf("accounts subsystem not initialized")
	}
	acc, err := a.accounts.Get(a.ctx, accountID)
	if err != nil || acc == nil {
		return nil, err
	}
	if _, err := a.authResolve.Resolve(a.ctx, acc); err != nil {
		return nil, err
	}
	saved, err := a.accounts.Get(a.ctx, accountID)
	if err != nil || saved == nil {
		return nil, err
	}
	dto := a.accountToDTO(*saved)
	return &dto, nil
}

// injectAccountVars adds the active account's decrypted credentials to the
// request's variable map so request bodies and headers can reference them
// via {{account.username}}, {{account.password}}, {{account.token}},
// {{account.totp}}.
func (a *App) injectAccountVars(vars map[string]string, projectID, accountID string) {
	if a.accounts == nil || a.vault == nil {
		return
	}
	acc, err := a.resolveActiveAccount(projectID, accountID)
	if err != nil || acc == nil {
		return
	}
	if acc.Username != "" {
		vars["account.username"] = acc.Username
	}
	if acc.PasswordEnc != "" {
		if pwd, err := a.vault.Decrypt(acc.PasswordEnc); err == nil {
			vars["account.password"] = pwd
		}
	}
	if acc.TokenEnc != "" {
		if tok, err := a.vault.Decrypt(acc.TokenEnc); err == nil {
			vars["account.token"] = tok
		}
	}
	if acc.APIKeyEnc != "" {
		if k, err := a.vault.Decrypt(acc.APIKeyEnc); err == nil {
			vars["account.apiKey"] = k
		}
	}
	if acc.TOTPSecretEnc != "" {
		if secret, err := a.vault.Decrypt(acc.TOTPSecretEnc); err == nil {
			if code, err := authpkg.GenerateTOTP(secret); err == nil {
				vars["account.totp"] = code
			}
		}
	}
}

// migrateLegacyAuth converts a pre-accounts project_auth row into a default
// "Default" account. Returns true when a migration was performed.
func (a *App) migrateLegacyAuth(projectID string) (bool, error) {
	rec, err := a.auth.Get(a.ctx, projectID)
	if err != nil || rec == nil {
		return false, err
	}
	if rec.Token == "" && rec.HeadersJSON == "" && rec.CookiesJSON == "" {
		return false, nil
	}
	tokenEnc, err := a.vault.Encrypt(rec.Token)
	if err != nil {
		return false, err
	}
	scheme := rec.Scheme
	if scheme == "" {
		scheme = "Bearer"
	}
	acc := domain.ProjectAccount{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Label:       "Default",
		Kind:        domain.AccountKindBearer,
		Scheme:      scheme,
		TokenEnc:    tokenEnc,
		ExpiresAt:   rec.ExpiresAt,
		TokenPath:   rec.TokenPath,
		UserJSON:    rec.UserJSON,
		CookiesJSON: rec.CookiesJSON,
		HeadersJSON: rec.HeadersJSON,
		IsDefault:   true,
		SortOrder:   0,
		CreatedAt:   rec.CapturedAt,
	}
	if err := a.accounts.Save(a.ctx, acc); err != nil {
		return false, err
	}
	return true, nil
}

func (a *App) accountToDTO(acc domain.ProjectAccount) ProjectAccountDTO {
	dto := ProjectAccountDTO{
		ID:                acc.ID,
		ProjectID:         acc.ProjectID,
		Label:             acc.Label,
		Kind:              string(acc.Kind),
		Scheme:            acc.Scheme,
		Username:          acc.Username,
		HasPassword:       acc.PasswordEnc != "",
		HasAPIKey:         acc.APIKeyEnc != "",
		APIKeyHeader:      acc.APIKeyHeader,
		APIKeyIn:          string(acc.APIKeyIn),
		HasToken:          acc.TokenEnc != "",
		HasRefreshToken:   acc.RefreshTokenEnc != "",
		ExpiresAt:         acc.ExpiresAt,
		HasTOTP:           acc.TOTPSecretEnc != "",
		TOTPParam:         acc.TOTPParam,
		LoginEndpointID:   acc.LoginEndpointID,
		LoginBodyTemplate: acc.LoginBodyTemplate,
		TokenPath:         acc.TokenPath,
		HasCookies:        acc.CookiesJSON != "" && acc.CookiesJSON != "[]",
		IsDefault:         acc.IsDefault,
		SortOrder:         acc.SortOrder,
		CreatedAt:         acc.CreatedAt,
		UpdatedAt:         acc.UpdatedAt,
	}
	if acc.TokenEnc != "" && a.vault != nil {
		if token, err := a.vault.Decrypt(acc.TokenEnc); err == nil {
			dto.TokenPreview = previewToken(token)
		}
	}
	if acc.UserJSON != "" {
		var user map[string]any
		if err := json.Unmarshal([]byte(acc.UserJSON), &user); err == nil {
			dto.User = user
		}
	}
	if acc.HeadersJSON != "" {
		var extra map[string]string
		if err := json.Unmarshal([]byte(acc.HeadersJSON), &extra); err == nil {
			dto.ExtraHeaders = extra
		}
	}
	if acc.OAuthConfigJSON != "" {
		var cfg authpkg.OAuth2Config
		if err := json.Unmarshal([]byte(acc.OAuthConfigJSON), &cfg); err == nil {
			dto.OAuth = &OAuthConfigDTO{
				GrantType: cfg.GrantType,
				TokenURL:  cfg.TokenURL,
				ClientID:  cfg.ClientID,
				HasSecret: cfg.ClientSecret != "",
				Scopes:    cfg.Scopes,
				Audience:  cfg.Audience,
				Username:  cfg.Username,
			}
		}
	}
	return dto
}

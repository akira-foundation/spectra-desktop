package app

import (
	"fmt"

	"github.com/google/uuid"

	authpkg "spectra-desktop/internal/auth"
	"spectra-desktop/internal/domain"
)

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
func (a *App) SetDefaultProjectAccount(projectID, accountID string) error {
	if a.accounts == nil {
		return fmt.Errorf("accounts subsystem not initialized")
	}
	return a.accounts.SetDefault(a.ctx, projectID, accountID)
}

func (a *App) DeleteProjectAccount(accountID string) error {
	if a.accounts == nil {
		return fmt.Errorf("accounts subsystem not initialized")
	}
	return a.accounts.Delete(a.ctx, accountID)
}

type AccountSecretsDTO struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Token    string `json:"token,omitempty"`
	APIKey   string `json:"apiKey,omitempty"`
}

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

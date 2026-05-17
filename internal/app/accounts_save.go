package app

import (
	"encoding/json"
	"fmt"
	authpkg "spectra-desktop/internal/auth"
	"spectra-desktop/internal/domain"
	"strings"

	"github.com/google/uuid"
)

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
	if input.ID == "" && a.billingGate != nil {
		existingCount := 0
		if list, err := a.accounts.List(a.ctx, input.ProjectID); err == nil {
			existingCount = len(list)
		}
		if existingCount >= 1 {
			if err := a.billingGate.Require(a.ctx, "multi_account"); err != nil {
				a.emitUpsell("multi_account", err)
				return nil, err
			}
		}
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

package app

import (
	"encoding/json"
	"time"

	authpkg "spectra-desktop/internal/auth"
	"spectra-desktop/internal/domain"
)

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

type OAuthConfigDTO struct {
	GrantType string   `json:"grantType"`
	TokenURL  string   `json:"tokenUrl"`
	ClientID  string   `json:"clientId"`
	HasSecret bool     `json:"hasSecret"`
	Scopes    []string `json:"scopes,omitempty"`
	Audience  string   `json:"audience,omitempty"`
	Username  string   `json:"username,omitempty"`
}

type SaveAccountInput struct {
	ID                string            `json:"id,omitempty"`
	ProjectID         string            `json:"projectID"`
	Label             string            `json:"label"`
	Kind              string            `json:"kind"`
	Scheme            string            `json:"scheme,omitempty"`
	Username          string            `json:"username,omitempty"`
	Password          *string           `json:"password,omitempty"`
	APIKey            *string           `json:"apiKey,omitempty"`
	APIKeyHeader      string            `json:"apiKeyHeader,omitempty"`
	APIKeyIn          string            `json:"apiKeyIn,omitempty"`
	Token             *string           `json:"token,omitempty"`
	RefreshToken      *string           `json:"refreshToken,omitempty"`
	OAuth             *SaveOAuthInput   `json:"oauth,omitempty"`
	TOTPSecret        *string           `json:"totpSecret,omitempty"`
	TOTPParam         string            `json:"totpParam,omitempty"`
	LoginEndpointID   string            `json:"loginEndpointId,omitempty"`
	LoginBodyTemplate string            `json:"loginBodyTemplate,omitempty"`
	TokenPath         string            `json:"tokenPath,omitempty"`
	ExtraHeaders      map[string]string `json:"extraHeaders,omitempty"`
	IsDefault         bool              `json:"isDefault,omitempty"`
	SortOrder         int               `json:"sortOrder,omitempty"`
	ExpiresAt         *time.Time        `json:"expiresAt,omitempty"`
	User              map[string]any    `json:"user,omitempty"`
}

type SaveOAuthInput struct {
	GrantType    string   `json:"grantType"`
	TokenURL     string   `json:"tokenUrl"`
	ClientID     string   `json:"clientId"`
	ClientSecret *string  `json:"clientSecret,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
	Audience     string   `json:"audience,omitempty"`
	Username     string   `json:"username,omitempty"`
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

package app

import (
	"encoding/json"
	"fmt"
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"strings"
	"time"
)

type ProjectAuthState struct {
	ProjectID            string         `json:"projectID"`
	Scheme               string         `json:"scheme"`
	HasToken             bool           `json:"hasToken"`
	TokenPreview         string         `json:"tokenPreview,omitempty"`
	TokenPath            string         `json:"tokenPath,omitempty"`
	User                 *core.AuthUser `json:"user,omitempty"`
	HasCookies           bool           `json:"hasCookies"`
	ExpiresAt            *time.Time     `json:"expiresAt,omitempty"`
	CapturedFromEndpoint string         `json:"capturedFromEndpoint,omitempty"`
	CapturedAt           time.Time      `json:"capturedAt"`
}

func (a *App) GetProjectAuth(projectID string) (*ProjectAuthState, error) {
	rec, err := a.auth.Get(a.ctx, projectID)
	if err != nil || rec == nil {
		return nil, err
	}
	state := &ProjectAuthState{
		ProjectID:            rec.ProjectID,
		Scheme:               rec.Scheme,
		HasToken:             rec.Token != "",
		TokenPath:            rec.TokenPath,
		HasCookies:           rec.CookiesJSON != "" && rec.CookiesJSON != "[]",
		ExpiresAt:            rec.ExpiresAt,
		CapturedFromEndpoint: rec.CapturedFromEndpoint,
		CapturedAt:           rec.CapturedAt,
	}
	if rec.Token != "" {
		state.TokenPreview = previewToken(rec.Token)
	}
	if rec.UserJSON != "" {
		var user core.AuthUser
		if err := json.Unmarshal([]byte(rec.UserJSON), &user); err == nil {
			state.User = &user
		}
	}
	return state, nil
}

func (a *App) ClearProjectAuth(projectID string) error {
	return a.auth.Clear(a.ctx, projectID)
}

type SetProjectAuthInput struct {
	ProjectID string `json:"projectID"`
	Scheme    string `json:"scheme"`
	Token     string `json:"token"`
}

func (a *App) SetProjectAuthManual(input SetProjectAuthInput) error {
	if input.ProjectID == "" {
		return fmt.Errorf("project id required")
	}
	scheme := input.Scheme
	if scheme == "" {
		scheme = string(core.AuthSchemeBearer)
	}
	rec := domain.ProjectAuth{
		ProjectID: input.ProjectID,
		Scheme:    scheme,
		Token:     strings.TrimSpace(input.Token),
	}
	return a.auth.Save(a.ctx, rec)
}

func previewToken(token string) string {
	if len(token) <= 12 {
		return token
	}
	return token[:6] + "…" + token[len(token)-4:]
}

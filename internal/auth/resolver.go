package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/secrets"
)

type HeaderInjection struct {
	Header     string
	Value      string
	QueryKey   string
	QueryValue string
}

type Resolver struct {
	repo  domain.AccountRepository
	vault *secrets.Vault
}

func NewResolver(repo domain.AccountRepository, vault *secrets.Vault) *Resolver {
	return &Resolver{repo: repo, vault: vault}
}

func (r *Resolver) Resolve(ctx context.Context, acc *domain.ProjectAccount) (HeaderInjection, error) {
	if acc == nil {
		return HeaderInjection{}, nil
	}
	switch acc.Kind {
	case domain.AccountKindBearer, domain.AccountKindLogin:
		return r.bearer(acc)
	case domain.AccountKindBasic:
		return r.basic(acc)
	case domain.AccountKindAPIKey:
		return r.apikey(acc)
	case domain.AccountKindOAuth2:
		return r.oauth2(ctx, acc)
	}
	return HeaderInjection{}, nil
}

func (r *Resolver) bearer(acc *domain.ProjectAccount) (HeaderInjection, error) {
	token, err := r.vault.Decrypt(acc.TokenEnc)
	if err != nil {
		return HeaderInjection{}, err
	}
	if token == "" {
		return HeaderInjection{}, nil
	}
	scheme := acc.Scheme
	if scheme == "" {
		scheme = "Bearer"
	}
	return HeaderInjection{
		Header: "Authorization",
		Value:  fmt.Sprintf("%s %s", scheme, token),
	}, nil
}

func (r *Resolver) basic(acc *domain.ProjectAccount) (HeaderInjection, error) {
	password, err := r.vault.Decrypt(acc.PasswordEnc)
	if err != nil {
		return HeaderInjection{}, err
	}
	if acc.Username == "" && password == "" {
		return HeaderInjection{}, nil
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(acc.Username + ":" + password))
	return HeaderInjection{
		Header: "Authorization",
		Value:  "Basic " + encoded,
	}, nil
}

func (r *Resolver) apikey(acc *domain.ProjectAccount) (HeaderInjection, error) {
	key, err := r.vault.Decrypt(acc.APIKeyEnc)
	if err != nil {
		return HeaderInjection{}, err
	}
	if key == "" {
		return HeaderInjection{}, nil
	}
	if acc.APIKeyIn == domain.APIKeyInQuery {
		name := acc.APIKeyHeader
		if name == "" {
			name = "api_key"
		}
		return HeaderInjection{QueryKey: name, QueryValue: key}, nil
	}
	header := acc.APIKeyHeader
	if header == "" {
		header = "X-API-Key"
	}
	return HeaderInjection{Header: header, Value: key}, nil
}

func (r *Resolver) oauth2(ctx context.Context, acc *domain.ProjectAccount) (HeaderInjection, error) {
	token, err := r.vault.Decrypt(acc.TokenEnc)
	if err != nil {
		return HeaderInjection{}, err
	}
	expired := acc.ExpiresAt != nil && time.Now().UTC().Add(30*time.Second).After(*acc.ExpiresAt)
	if token == "" || expired {
		refreshed, refreshErr := r.refreshOAuth2(ctx, acc)
		if refreshErr != nil {
			if token == "" {
				return HeaderInjection{}, refreshErr
			}
		} else {
			token = refreshed
		}
	}
	if token == "" {
		return HeaderInjection{}, nil
	}
	scheme := acc.Scheme
	if scheme == "" {
		scheme = "Bearer"
	}
	return HeaderInjection{
		Header: "Authorization",
		Value:  fmt.Sprintf("%s %s", scheme, token),
	}, nil
}

func (r *Resolver) refreshOAuth2(ctx context.Context, acc *domain.ProjectAccount) (string, error) {
	if acc.OAuthConfigJSON == "" {
		return "", fmt.Errorf("oauth2: account %q missing config", acc.Label)
	}
	var cfg OAuth2Config
	if err := json.Unmarshal([]byte(acc.OAuthConfigJSON), &cfg); err != nil {
		return "", fmt.Errorf("oauth2: parse config: %w", err)
	}
	refresh, err := r.vault.Decrypt(acc.RefreshTokenEnc)
	if err != nil {
		return "", err
	}
	password, err := r.vault.Decrypt(acc.PasswordEnc)
	if err != nil {
		return "", err
	}

	token, err := FetchOAuth2Token(ctx, cfg, refresh, password)
	if err != nil {
		return "", err
	}

	tokenEnc, err := r.vault.Encrypt(token.AccessToken)
	if err != nil {
		return "", err
	}
	acc.TokenEnc = tokenEnc
	if token.RefreshToken != "" {
		refreshEnc, err := r.vault.Encrypt(token.RefreshToken)
		if err != nil {
			return "", err
		}
		acc.RefreshTokenEnc = refreshEnc
	}
	if !token.ExpiresAt.IsZero() {
		t := token.ExpiresAt
		acc.ExpiresAt = &t
	}
	if r.repo != nil {
		if err := r.repo.Save(ctx, *acc); err != nil {
			return "", err
		}
	}
	return token.AccessToken, nil
}

func (r *Resolver) MergeTOTP(acc *domain.ProjectAccount) (HeaderInjection, error) {
	if acc == nil || acc.TOTPSecretEnc == "" {
		return HeaderInjection{}, nil
	}
	secret, err := r.vault.Decrypt(acc.TOTPSecretEnc)
	if err != nil {
		return HeaderInjection{}, err
	}
	code, err := GenerateTOTP(secret)
	if err != nil || code == "" {
		return HeaderInjection{}, err
	}
	param := strings.TrimSpace(acc.TOTPParam)
	if param == "" {
		param = "X-OTP"
	}
	if strings.HasPrefix(param, "?") {
		return HeaderInjection{QueryKey: strings.TrimPrefix(param, "?"), QueryValue: code}, nil
	}
	return HeaderInjection{Header: param, Value: code}, nil
}

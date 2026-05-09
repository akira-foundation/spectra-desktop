package domain

import (
	"context"
	"time"
)

type AccountKind string

const (
	AccountKindBearer AccountKind = "bearer"
	AccountKindBasic  AccountKind = "basic"
	AccountKindAPIKey AccountKind = "apikey"
	AccountKindOAuth2 AccountKind = "oauth2"
	AccountKindLogin  AccountKind = "login"
)

type APIKeyLocation string

const (
	APIKeyInHeader APIKeyLocation = "header"
	APIKeyInQuery  APIKeyLocation = "query"
)

type ProjectAccount struct {
	ID                string
	ProjectID         string
	Label             string
	Kind              AccountKind
	Scheme            string
	Username          string
	PasswordEnc       string
	APIKeyEnc         string
	APIKeyHeader      string
	APIKeyIn          APIKeyLocation
	TokenEnc          string
	RefreshTokenEnc   string
	ExpiresAt         *time.Time
	OAuthConfigJSON   string
	TOTPSecretEnc     string
	TOTPParam         string
	LoginEndpointID   string
	LoginBodyTemplate string
	TokenPath         string
	UserJSON          string
	CookiesJSON       string
	HeadersJSON       string
	IsDefault         bool
	SortOrder         int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type AccountRepository interface {
	List(ctx context.Context, projectID string) ([]ProjectAccount, error)
	Get(ctx context.Context, id string) (*ProjectAccount, error)
	GetDefault(ctx context.Context, projectID string) (*ProjectAccount, error)
	Save(ctx context.Context, acc ProjectAccount) error
	SetDefault(ctx context.Context, projectID, accountID string) error
	Delete(ctx context.Context, id string) error
}

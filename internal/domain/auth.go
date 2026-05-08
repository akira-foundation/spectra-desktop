package domain

import (
	"context"
	"time"
)

type ProjectAuth struct {
	ProjectID            string
	Scheme               string
	Token                string
	TokenPath            string
	UserJSON             string
	CookiesJSON          string
	HeadersJSON          string
	ExpiresAt            *time.Time
	CapturedFromEndpoint string
	CapturedAt           time.Time
	UpdatedAt            time.Time
}

type AuthRepository interface {
	Get(ctx context.Context, projectID string) (*ProjectAuth, error)
	Save(ctx context.Context, auth ProjectAuth) error
	Clear(ctx context.Context, projectID string) error
}

package domain

import (
	"context"
	"time"
)

type EndpointSnapshot struct {
	ID            string
	ProjectID     string
	Hash          string
	PayloadJSON   string
	EndpointCount int
	ScannedAt     time.Time
	CreatedAt     time.Time
}

type SnapshotRepository interface {
	Save(ctx context.Context, snapshot EndpointSnapshot) error
	List(ctx context.Context, projectID string, limit int) ([]EndpointSnapshot, error)
	GetByID(ctx context.Context, id string) (*EndpointSnapshot, error)
	Latest(ctx context.Context, projectID string) (*EndpointSnapshot, error)
	Predecessor(ctx context.Context, projectID string, scannedAt time.Time) (*EndpointSnapshot, error)
	TrimOldest(ctx context.Context, projectID string, keep int) error
}

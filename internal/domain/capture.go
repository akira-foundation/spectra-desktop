package domain

import (
	"context"
	"time"
)

type EndpointCapture struct {
	ID          string
	ProjectID   string
	EndpointKey string
	Name        string
	Source      string
	Path        string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CaptureRepository interface {
	List(ctx context.Context, projectID, endpointKey string) ([]EndpointCapture, error)
	Replace(ctx context.Context, projectID, endpointKey string, captures []EndpointCapture) error
	DeleteByEndpoint(ctx context.Context, projectID, endpointKey string) error
}

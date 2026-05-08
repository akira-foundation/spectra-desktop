package domain

import (
	"context"
	"time"

	"spectra-desktop/internal/core"
)

type ProjectStats struct {
	Routes        int        `json:"routes"`
	Models        int        `json:"models"`
	Middleware    int        `json:"middleware"`
	Controllers   int        `json:"controllers"`
	Errors        int        `json:"errors"`
	LastScannedAt *time.Time `json:"lastScannedAt,omitempty"`
}

type EndpointRepository interface {
	List(ctx context.Context, projectID string) ([]core.Endpoint, error)
	GetByID(ctx context.Context, id string) (*core.Endpoint, error)
	Replace(ctx context.Context, projectID string, endpoints []core.Endpoint) error
	DeleteByProject(ctx context.Context, projectID string) error
	UpdateAuthOverride(ctx context.Context, endpointID string, role core.AuthRole, tokenPath string) error
	Stats(ctx context.Context, projectID string) (ProjectStats, error)
}

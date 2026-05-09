package domain

import (
	"context"
	"time"
)

type MockSource string

const (
	MockSourceAuto      MockSource = "auto"
	MockSourceHistory   MockSource = "history"
	MockSourceCustom    MockSource = "custom"
	MockSourceGenerated MockSource = "generated"
	MockSourceNoMatch   MockSource = "no-match"
)

type MockOverride struct {
	ID          string
	ProjectID   string
	EndpointID  string
	Enabled     bool
	Status      int
	LatencyMs   int
	Body        string
	HeadersJSON string
	Source      MockSource
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MockRepository interface {
	List(ctx context.Context, projectID string) ([]MockOverride, error)
	Get(ctx context.Context, projectID, endpointID string) (*MockOverride, error)
	Save(ctx context.Context, override MockOverride) error
	Delete(ctx context.Context, id string) error
	DeleteByProject(ctx context.Context, projectID string) error
}

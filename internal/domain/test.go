package domain

import (
	"context"
	"time"
)

type EndpointTest struct {
	ID          string
	ProjectID   string
	EndpointKey string
	Name        string
	Kind        string
	JSONPath    string
	Op          string
	Expected    string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TestRepository interface {
	List(ctx context.Context, projectID, endpointKey string) ([]EndpointTest, error)
	Replace(ctx context.Context, projectID, endpointKey string, tests []EndpointTest) error
	DeleteByEndpoint(ctx context.Context, projectID, endpointKey string) error
}

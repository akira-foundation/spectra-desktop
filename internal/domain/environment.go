package domain

import (
	"context"
	"time"
)

type Environment struct {
	ID        string
	ProjectID string
	Name      string
	Vars      map[string]string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type EnvironmentInput struct {
	ID        string
	ProjectID string
	Name      string
	Vars      map[string]string
	SortOrder int
}

type EnvironmentRepository interface {
	List(ctx context.Context, projectID string) ([]Environment, error)
	GetByID(ctx context.Context, id string) (*Environment, error)
	Save(ctx context.Context, input EnvironmentInput) (*Environment, error)
	Delete(ctx context.Context, id string) error
}

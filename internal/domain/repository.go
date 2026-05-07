package domain

import (
	"context"
	"errors"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrProjectExists   = errors.New("project already exists")
)

type ProjectRepository interface {
	List(ctx context.Context) ([]Project, error)
	GetByID(ctx context.Context, id string) (*Project, error)
	GetByPath(ctx context.Context, path string) (*Project, error)
	Save(ctx context.Context, input ProjectInput) (*Project, error)
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status ProjectStatus) error
	MarkSynced(ctx context.Context, id string) error
}

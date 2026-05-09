package domain

import (
	"context"
	"time"
)

type Collection struct {
	ID          string
	ProjectID   string
	Name        string
	Description string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Items       []CollectionItem
}

type CollectionItem struct {
	ID              string
	CollectionID    string
	EndpointID      string
	SortOrder       int
	BodyOverride    string
	HeadersOverride string
	SkipOnFailure   bool
	IterateDataset  bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CollectionRepository interface {
	List(ctx context.Context, projectID string) ([]Collection, error)
	Get(ctx context.Context, id string) (*Collection, error)
	Create(ctx context.Context, c Collection) (*Collection, error)
	Update(ctx context.Context, c Collection) error
	Delete(ctx context.Context, id string) error
	ReplaceItems(ctx context.Context, collectionID string, items []CollectionItem) error
}

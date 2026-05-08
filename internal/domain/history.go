package domain

import (
	"context"
	"time"
)

type HistoryEntry struct {
	ID              string
	ProjectID       string
	EndpointID      string
	Method          string
	URL             string
	RequestHeaders  string
	RequestBody     string
	ResponseStatus  int
	ResponseHeaders string
	ResponseBody    string
	DurationMs      int
	SizeBytes       int
	Error           string
	CreatedAt       time.Time
}

type HistoryRepository interface {
	Save(ctx context.Context, entry HistoryEntry) error
	List(ctx context.Context, projectID string, limit int) ([]HistoryEntry, error)
	GetByID(ctx context.Context, id string) (*HistoryEntry, error)
	Clear(ctx context.Context, projectID string) error
	TrimOldest(ctx context.Context, projectID string, keep int) error
}

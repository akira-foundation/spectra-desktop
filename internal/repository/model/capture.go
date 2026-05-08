package model

import (
	"time"

	"github.com/uptrace/bun"
)

type EndpointCapture struct {
	bun.BaseModel `bun:"table:endpoint_captures"`

	ID          string    `bun:"id,pk"`
	ProjectID   string    `bun:"project_id,notnull"`
	EndpointKey string    `bun:"endpoint_key,notnull"`
	Name        string    `bun:"name,notnull"`
	Source      string    `bun:"source,notnull"`
	Path        string    `bun:"path,notnull"`
	SortOrder   int       `bun:"sort_order,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
	UpdatedAt   time.Time `bun:"updated_at,notnull"`
}

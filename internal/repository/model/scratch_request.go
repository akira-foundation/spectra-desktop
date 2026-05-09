package model

import (
	"time"

	"github.com/uptrace/bun"
)

type ScratchRequest struct {
	bun.BaseModel `bun:"table:scratch_requests"`

	ID           string    `bun:"id,pk"`
	ProjectID    string    `bun:"project_id,notnull"`
	Name         string    `bun:"name,notnull"`
	Method       string    `bun:"method,notnull"`
	URL          string    `bun:"url,notnull"`
	HeadersJSON  string    `bun:"headers_json,notnull"`
	Body         string    `bun:"body,notnull"`
	ResponseJSON string    `bun:"response_json,nullzero"`
	SortOrder    int       `bun:"sort_order,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull"`
	UpdatedAt    time.Time `bun:"updated_at,notnull"`
}

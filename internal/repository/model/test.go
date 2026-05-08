package model

import (
	"time"

	"github.com/uptrace/bun"
)

type EndpointTest struct {
	bun.BaseModel `bun:"table:endpoint_tests"`

	ID          string    `bun:"id,pk"`
	ProjectID   string    `bun:"project_id,notnull"`
	EndpointKey string    `bun:"endpoint_key,notnull"`
	Name        string    `bun:"name,notnull"`
	Kind        string    `bun:"kind,notnull"`
	JSONPath    string    `bun:"json_path,notnull"`
	Op          string    `bun:"op,notnull"`
	Expected    string    `bun:"expected,notnull"`
	SortOrder   int       `bun:"sort_order,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
	UpdatedAt   time.Time `bun:"updated_at,notnull"`
}

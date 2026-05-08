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
	Name        string    `bun:"name,notnull,default:''"`
	Kind        string    `bun:"kind,notnull"`
	JSONPath    string    `bun:"json_path,notnull,default:''"`
	Op          string    `bun:"op,notnull,default:''"`
	Expected    string    `bun:"expected,notnull,default:''"`
	SortOrder   int       `bun:"sort_order,notnull,default:0"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
	UpdatedAt   time.Time `bun:"updated_at,notnull"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Environment struct {
	bun.BaseModel `bun:"table:environments"`

	ID        string    `bun:"id,pk"`
	ProjectID string    `bun:"project_id,notnull"`
	Name      string    `bun:"name,notnull"`
	VarsJSON  string    `bun:"vars_json,notnull,default:'{}'"`
	SortOrder int       `bun:"sort_order,notnull,default:0"`
	CreatedAt time.Time `bun:"created_at,notnull"`
	UpdatedAt time.Time `bun:"updated_at,notnull"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

type EndpointDataset struct {
	bun.BaseModel `bun:"table:endpoint_datasets"`

	ProjectID   string    `bun:"project_id,pk"`
	EndpointKey string    `bun:"endpoint_key,pk"`
	RowsJSON    string    `bun:"rows_json,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
	UpdatedAt   time.Time `bun:"updated_at,notnull"`
}

package model

import (
	"time"

	"github.com/uptrace/bun"
)

type CapturedValue struct {
	bun.BaseModel `bun:"table:captured_values"`

	ProjectID   string    `bun:"project_id,pk"`
	Name        string    `bun:"name,pk"`
	Value       string    `bun:"value,notnull"`
	EndpointKey string    `bun:"endpoint_key,notnull"`
	CapturedAt  time.Time `bun:"captured_at,notnull"`
}

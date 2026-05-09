package model

import (
	"time"

	"github.com/uptrace/bun"
)

type MockOverride struct {
	bun.BaseModel `bun:"table:mock_overrides"`

	ID          string    `bun:"id,pk"`
	ProjectID   string    `bun:"project_id,notnull"`
	EndpointID  string    `bun:"endpoint_id,notnull"`
	Enabled     bool      `bun:"enabled,notnull,default:true"`
	Status      int       `bun:"status,notnull,default:200"`
	LatencyMs   int       `bun:"latency_ms,notnull,default:0"`
	Body        string    `bun:"body,notnull,default:''"`
	HeadersJSON string    `bun:"headers_json,notnull,default:''"`
	Source      string    `bun:"source,notnull,default:'auto'"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
	UpdatedAt   time.Time `bun:"updated_at,notnull"`
}

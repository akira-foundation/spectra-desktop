package model

import (
	"time"

	"github.com/uptrace/bun"
)

type EndpointSnapshot struct {
	bun.BaseModel `bun:"table:endpoint_snapshots"`

	ID            string    `bun:"id,pk"`
	ProjectID     string    `bun:"project_id,notnull"`
	Hash          string    `bun:"hash,notnull"`
	PayloadJSON   string    `bun:"payload_json,notnull"`
	EndpointCount int       `bun:"endpoint_count,notnull,default:0"`
	ScannedAt     time.Time `bun:"scanned_at,notnull"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
}

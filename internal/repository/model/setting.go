package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Setting struct {
	bun.BaseModel `bun:"table:settings"`

	Key       string    `bun:"key,pk"`
	Value     string    `bun:"value,notnull"`
	UpdatedAt time.Time `bun:"updated_at,notnull"`
}

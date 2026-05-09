package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Collection struct {
	bun.BaseModel `bun:"table:collections"`

	ID          string    `bun:"id,pk"`
	ProjectID   string    `bun:"project_id,notnull"`
	Name        string    `bun:"name,notnull"`
	Description string    `bun:"description,notnull"`
	SortOrder   int       `bun:"sort_order,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull"`
	UpdatedAt   time.Time `bun:"updated_at,notnull"`
}

type CollectionItem struct {
	bun.BaseModel `bun:"table:collection_items"`

	ID              string    `bun:"id,pk"`
	CollectionID    string    `bun:"collection_id,notnull"`
	EndpointID      string    `bun:"endpoint_id,notnull"`
	SortOrder       int       `bun:"sort_order,notnull"`
	BodyOverride    string    `bun:"body_override,notnull"`
	HeadersOverride string    `bun:"headers_override,notnull"`
	SkipOnFailure   int       `bun:"skip_on_failure,notnull"`
	CreatedAt       time.Time `bun:"created_at,notnull"`
	UpdatedAt       time.Time `bun:"updated_at,notnull"`
}

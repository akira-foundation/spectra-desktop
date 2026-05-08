package model

import (
	"time"

	"github.com/uptrace/bun"
)

type RequestHistory struct {
	bun.BaseModel `bun:"table:request_history"`

	ID              string    `bun:"id,pk"`
	ProjectID       string    `bun:"project_id,notnull"`
	EndpointID      string    `bun:"endpoint_id,notnull,default:''"`
	Method          string    `bun:"method,notnull"`
	URL             string    `bun:"url,notnull"`
	RequestHeaders  string    `bun:"request_headers,notnull,default:'{}'"`
	RequestBody     string    `bun:"request_body,notnull,default:''"`
	ResponseStatus  int       `bun:"response_status,notnull,default:0"`
	ResponseHeaders string    `bun:"response_headers,notnull,default:'{}'"`
	ResponseBody    string    `bun:"response_body,notnull,default:''"`
	DurationMs      int       `bun:"duration_ms,notnull,default:0"`
	SizeBytes       int       `bun:"size_bytes,notnull,default:0"`
	Error           string    `bun:"error,notnull,default:''"`
	CreatedAt       time.Time `bun:"created_at,notnull"`
}

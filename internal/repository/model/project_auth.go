package model

import (
	"time"

	"github.com/uptrace/bun"
)

type ProjectAuth struct {
	bun.BaseModel `bun:"table:project_auth"`

	ProjectID            string     `bun:"project_id,pk"`
	Scheme               string     `bun:"scheme,notnull,default:''"`
	Token                string     `bun:"token,notnull,default:''"`
	TokenPath            string     `bun:"token_path,notnull,default:''"`
	UserJSON             string     `bun:"user_json,notnull,default:''"`
	CookiesJSON          string     `bun:"cookies_json,notnull,default:''"`
	HeadersJSON          string     `bun:"headers_json,notnull,default:''"`
	ExpiresAt            *time.Time `bun:"expires_at"`
	CapturedFromEndpoint string     `bun:"captured_from_endpoint,notnull,default:''"`
	CapturedAt           time.Time  `bun:"captured_at,notnull"`
	UpdatedAt            time.Time  `bun:"updated_at,notnull"`
}

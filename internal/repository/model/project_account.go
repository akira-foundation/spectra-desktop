package model

import (
	"time"

	"github.com/uptrace/bun"
)

type ProjectAccount struct {
	bun.BaseModel `bun:"table:project_accounts"`

	ID                string     `bun:"id,pk"`
	ProjectID         string     `bun:"project_id,notnull"`
	Label             string     `bun:"label,notnull"`
	Kind              string     `bun:"kind,notnull,default:'bearer'"`
	Scheme            string     `bun:"scheme,notnull,default:''"`
	Username          string     `bun:"username,notnull,default:''"`
	PasswordEnc       string     `bun:"password_enc,notnull,default:''"`
	APIKeyEnc         string     `bun:"api_key_enc,notnull,default:''"`
	APIKeyHeader      string     `bun:"api_key_header,notnull,default:''"`
	APIKeyIn          string     `bun:"api_key_in,notnull,default:'header'"`
	TokenEnc          string     `bun:"token_enc,notnull,default:''"`
	RefreshTokenEnc   string     `bun:"refresh_token_enc,notnull,default:''"`
	ExpiresAt         *time.Time `bun:"expires_at"`
	OAuthConfigJSON   string     `bun:"oauth_config_json,notnull,default:''"`
	TOTPSecretEnc     string     `bun:"totp_secret_enc,notnull,default:''"`
	TOTPParam         string     `bun:"totp_param,notnull,default:''"`
	LoginEndpointID   string     `bun:"login_endpoint_id,notnull,default:''"`
	LoginBodyTemplate string     `bun:"login_body_template,notnull,default:''"`
	TokenPath         string     `bun:"token_path,notnull,default:''"`
	UserJSON          string     `bun:"user_json,notnull,default:''"`
	CookiesJSON       string     `bun:"cookies_json,notnull,default:''"`
	HeadersJSON       string     `bun:"headers_json,notnull,default:''"`
	IsDefault         bool       `bun:"is_default,notnull,default:false"`
	SortOrder         int        `bun:"sort_order,notnull,default:0"`
	CreatedAt         time.Time  `bun:"created_at,notnull"`
	UpdatedAt         time.Time  `bun:"updated_at,notnull"`
}

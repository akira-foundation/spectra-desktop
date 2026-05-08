package core

import (
	"net/http"
	"time"
)

type AuthRole string

const (
	AuthRoleNone    AuthRole = ""
	AuthRoleLogin   AuthRole = "login"
	AuthRoleLogout  AuthRole = "logout"
	AuthRoleRefresh AuthRole = "refresh"
	AuthRoleCSRF    AuthRole = "csrf"
)

type AuthScheme string

const (
	AuthSchemeNone   AuthScheme = ""
	AuthSchemeBearer AuthScheme = "bearer"
	AuthSchemeCookie AuthScheme = "cookie"
	AuthSchemeBasic  AuthScheme = "basic"
	AuthSchemeAPIKey AuthScheme = "api_key"
	AuthSchemeCustom AuthScheme = "custom"
)

type AuthConfidence string

const (
	AuthConfidenceHigh   AuthConfidence = "high"
	AuthConfidenceMedium AuthConfidence = "medium"
	AuthConfidenceLow    AuthConfidence = "low"
)

type AuthRoleHint struct {
	Role       AuthRole
	Confidence AuthConfidence
	Reason     string
}

type AuthContext struct {
	Scheme    AuthScheme
	Token     string
	TokenPath string
	Cookies   []http.Cookie
	Headers   map[string]string
	ExpiresAt *time.Time
}

type AuthExtraction struct {
	Token     string
	TokenPath string
	Cookies   []http.Cookie
	ExpiresAt *time.Time
}

type AuthResponse struct {
	Status  int
	Headers http.Header
	Body    []byte
}

type AuthCapable interface {
	DetectAuthRole(ep Endpoint) AuthRoleHint
	ExtractCredentials(resp AuthResponse) (*AuthExtraction, bool)
	ApplyAuth(req *http.Request, ctx AuthContext)
	DefaultScheme() AuthScheme
}

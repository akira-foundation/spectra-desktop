package laravel

import (
	"encoding/json"
	"net/http"
	"strings"

	"spectra-desktop/internal/core"
)

var tokenPaths = [][]string{
	{"data", "token"},
	{"data", "access_token"},
	{"data", "accessToken"},
	{"data", "auth", "token"},
	{"data", "user", "token"},
	{"token"},
	{"access_token"},
	{"accessToken"},
	{"auth", "token"},
	{"meta", "token"},
	{"result", "token"},
}

var userPaths = [][]string{
	{"data", "user"},
	{"data", "auth", "user"},
	{"user"},
	{"auth", "user"},
	{"data"},
	{"result", "user"},
}

func (AuthCapability) ExtractCredentials(resp core.AuthResponse) (*core.AuthExtraction, bool) {
	if len(resp.Body) == 0 {
		return extractFromHeaders(resp), false
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body, &payload); err != nil {
		return extractFromHeaders(resp), false
	}

	out := &core.AuthExtraction{}
	for _, p := range tokenPaths {
		if v, ok := lookupString(payload, p); ok && v != "" {
			out.Token = v
			out.TokenPath = strings.Join(p, ".")
			break
		}
	}

	for _, p := range userPaths {
		if u, ok := lookupObject(payload, p); ok {
			user := buildUser(u)
			if user != nil {
				out.User = user
				out.UserPath = strings.Join(p, ".")
				break
			}
		}
	}

	if cookies := parseSetCookies(resp.Headers); len(cookies) > 0 {
		out.Cookies = cookies
	}

	if out.Token == "" && out.User == nil && len(out.Cookies) == 0 {
		return nil, false
	}
	return out, true
}

func extractFromHeaders(resp core.AuthResponse) *core.AuthExtraction {
	cookies := parseSetCookies(resp.Headers)
	if len(cookies) == 0 {
		return nil
	}
	return &core.AuthExtraction{Cookies: cookies}
}

func parseSetCookies(h http.Header) []http.Cookie {
	if h == nil {
		return nil
	}
	resp := http.Response{Header: h}
	cs := resp.Cookies()
	out := make([]http.Cookie, 0, len(cs))
	for _, c := range cs {
		if c == nil {
			continue
		}
		out = append(out, *c)
	}
	return out
}

package laravel

import (
	"encoding/json"
	"net/http"
	"strings"

	"spectra-desktop/internal/core"
)

type AuthCapability struct{}

func (AuthCapability) DefaultScheme() core.AuthScheme {
	return core.AuthSchemeBearer
}

func (a AuthCapability) DetectAuthRole(ep core.Endpoint) core.AuthRoleHint {
	path := strings.ToLower(ep.Path)
	handler := strings.ToLower(ep.Handler)
	mw := lowerSlice(ep.Middleware)

	if hint := matchCSRF(path); hint.Role != core.AuthRoleNone {
		return hint
	}
	if hint := matchLogout(path, handler, mw); hint.Role != core.AuthRoleNone {
		return hint
	}
	if hint := matchRefresh(path, handler); hint.Role != core.AuthRoleNone {
		return hint
	}
	if hint := matchLogin(ep, path, handler, mw); hint.Role != core.AuthRoleNone {
		return hint
	}
	return core.AuthRoleHint{Role: core.AuthRoleNone}
}

func matchCSRF(path string) core.AuthRoleHint {
	if strings.Contains(path, "sanctum/csrf-cookie") {
		return core.AuthRoleHint{
			Role:       core.AuthRoleCSRF,
			Confidence: core.AuthConfidenceHigh,
			Reason:     "Sanctum CSRF cookie endpoint",
		}
	}
	return core.AuthRoleHint{}
}

func matchLogout(path, handler string, mw []string) core.AuthRoleHint {
	if !strings.Contains(path, "logout") && !strings.Contains(handler, "logout") {
		return core.AuthRoleHint{}
	}
	conf := core.AuthConfidenceMedium
	for _, m := range mw {
		if strings.HasPrefix(m, "auth") {
			conf = core.AuthConfidenceHigh
			break
		}
	}
	return core.AuthRoleHint{
		Role:       core.AuthRoleLogout,
		Confidence: conf,
		Reason:     "path or handler contains 'logout'",
	}
}

func matchRefresh(path, handler string) core.AuthRoleHint {
	if strings.Contains(path, "refresh") || strings.Contains(handler, "refresh") {
		return core.AuthRoleHint{
			Role:       core.AuthRoleRefresh,
			Confidence: core.AuthConfidenceMedium,
			Reason:     "path or handler contains 'refresh'",
		}
	}
	return core.AuthRoleHint{}
}

func matchLogin(ep core.Endpoint, path, handler string, mw []string) core.AuthRoleHint {
	if ep.Method != core.MethodPost {
		return core.AuthRoleHint{}
	}
	score := 0
	reasons := []string{}
	for _, kw := range []string{"login", "signin", "sign-in", "authenticate", "auth/token"} {
		if strings.Contains(path, kw) {
			score += 3
			reasons = append(reasons, "path contains '"+kw+"'")
			break
		}
	}
	for _, kw := range []string{"login", "signin", "authenticate"} {
		if strings.Contains(handler, kw) {
			score += 2
			reasons = append(reasons, "handler contains '"+kw+"'")
			break
		}
	}
	if strings.Contains(handler, "auth") {
		score++
	}
	hasGuest := false
	for _, m := range mw {
		if m == "guest" || strings.HasPrefix(m, "guest:") {
			hasGuest = true
			break
		}
	}
	if hasGuest {
		score += 2
		reasons = append(reasons, "guest middleware")
	}
	if score == 0 {
		return core.AuthRoleHint{}
	}
	conf := core.AuthConfidenceLow
	if score >= 5 {
		conf = core.AuthConfidenceHigh
	} else if score >= 3 {
		conf = core.AuthConfidenceMedium
	}
	return core.AuthRoleHint{
		Role:       core.AuthRoleLogin,
		Confidence: conf,
		Reason:     strings.Join(reasons, ", "),
	}
}

func (AuthCapability) ApplyAuth(req *http.Request, ctx core.AuthContext) {
	if ctx.Token != "" {
		switch ctx.Scheme {
		case core.AuthSchemeBearer, core.AuthSchemeNone:
			req.Header.Set("Authorization", "Bearer "+ctx.Token)
		}
	}
	for k, v := range ctx.Headers {
		if req.Header.Get(k) == "" {
			req.Header.Set(k, v)
		}
	}
	for _, c := range ctx.Cookies {
		req.AddCookie(&c)
	}
}

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

func lookupString(m map[string]any, path []string) (string, bool) {
	v, ok := lookup(m, path)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(s), s != ""
}

func lookupObject(m map[string]any, path []string) (map[string]any, bool) {
	v, ok := lookup(m, path)
	if !ok {
		return nil, false
	}
	obj, ok := v.(map[string]any)
	return obj, ok
}

func lookup(m map[string]any, path []string) (any, bool) {
	var cur any = m
	for _, key := range path {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, exists := obj[key]
		if !exists {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func buildUser(obj map[string]any) *core.AuthUser {
	if obj == nil {
		return nil
	}
	user := &core.AuthUser{}
	user.ID = firstString(obj, "id", "uuid", "sub")
	user.Name = firstString(obj, "name", "full_name", "fullname", "display_name", "displayName")
	user.Username = firstString(obj, "username", "user_name", "userName", "login", "handle")
	user.Email = firstString(obj, "email", "email_address", "mail")
	user.Role = firstString(obj, "role", "role_name", "type")
	if user.ID == "" && user.Name == "" && user.Username == "" && user.Email == "" {
		return nil
	}
	if raw, err := json.Marshal(obj); err == nil {
		user.Raw = string(raw)
	}
	return user
}

func firstString(obj map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

func lowerSlice(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(s)
	}
	return out
}

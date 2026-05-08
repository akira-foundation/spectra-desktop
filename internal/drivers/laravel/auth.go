package laravel

import (
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

func (AuthCapability) ExtractCredentials(resp core.AuthResponse) (*core.AuthExtraction, bool) {
	return nil, false
}

func lowerSlice(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(s)
	}
	return out
}

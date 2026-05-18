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

package app

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"spectra-desktop/internal/auth"
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/httpclient"
	"strings"
)

func (a *App) isLogoutEndpoint(projectID, endpointID string) bool {
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return false
	}
	return project.LogoutEndpointID != "" && project.LogoutEndpointID == endpointID
}

func (a *App) UpdateProjectAuthRoutes(projectID, loginID, logoutID, tokenPath string) error {
	return a.projects.UpdateAuthRoutes(a.ctx, projectID, loginID, logoutID, tokenPath)
}

func (a *App) applyProjectAuth(projectID, accountID string, base map[string]string) (map[string]string, []http.Cookie, map[string]string) {
	headers := map[string]string{}
	for k, v := range base {
		headers[k] = v
	}
	queryParams := map[string]string{}

	// New path: resolve via accounts when available.
	if a.authResolve != nil {
		acc, err := a.resolveActiveAccount(projectID, accountID)
		if err == nil && acc != nil {
			injection, err := a.authResolve.Resolve(a.ctx, acc)
			if err == nil {
				applyInjection(headers, queryParams, injection)
			}
			if totp, err := a.authResolve.MergeTOTP(acc); err == nil {
				applyInjection(headers, queryParams, totp)
			}
			if acc.HeadersJSON != "" {
				var extra map[string]string
				if err := json.Unmarshal([]byte(acc.HeadersJSON), &extra); err == nil {
					for k, v := range extra {
						if _, exists := headers[k]; !exists {
							headers[k] = v
						}
					}
				}
			}
			var cookies []http.Cookie
			if acc.CookiesJSON != "" {
				_ = json.Unmarshal([]byte(acc.CookiesJSON), &cookies)
			}
			return headers, cookies, queryParams
		}
	}

	// Legacy fallback: pre-accounts project_auth row.
	rec, err := a.auth.Get(a.ctx, projectID)
	if err != nil || rec == nil {
		return headers, nil, queryParams
	}
	if rec.Token != "" {
		if _, exists := headers["Authorization"]; !exists {
			a.applyAuthScheme(projectID, headers, rec)
		}
	}
	if rec.HeadersJSON != "" {
		var extra map[string]string
		if err := json.Unmarshal([]byte(rec.HeadersJSON), &extra); err == nil {
			for k, v := range extra {
				if _, exists := headers[k]; !exists {
					headers[k] = v
				}
			}
		}
	}
	var cookies []http.Cookie
	if rec.CookiesJSON != "" {
		_ = json.Unmarshal([]byte(rec.CookiesJSON), &cookies)
	}
	return headers, cookies, queryParams
}

// resolveActiveAccount picks the account for this request:
//  1. Explicit accountID (per-tab override).
//  2. Project default account.
//  3. nil (callers fall back to legacy auth).
func (a *App) resolveActiveAccount(projectID, accountID string) (*domain.ProjectAccount, error) {
	if a.accounts == nil {
		return nil, nil
	}
	if accountID != "" {
		acc, err := a.accounts.Get(a.ctx, accountID)
		if err != nil || acc == nil {
			return nil, err
		}
		if acc.ProjectID == projectID {
			return acc, nil
		}
	}
	return a.accounts.GetDefault(a.ctx, projectID)
}

func applyInjection(headers, query map[string]string, inj auth.HeaderInjection) {
	if inj.Header != "" && inj.Value != "" {
		if _, exists := headers[inj.Header]; !exists {
			headers[inj.Header] = inj.Value
		}
	}
	if inj.QueryKey != "" && inj.QueryValue != "" {
		query[inj.QueryKey] = inj.QueryValue
	}
}

func appendQueryParams(target string, params map[string]string) string {
	if len(params) == 0 {
		return target
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return target
	}
	q := parsed.Query()
	for k, v := range params {
		if !q.Has(k) {
			q.Set(k, v)
		}
	}
	parsed.RawQuery = q.Encode()
	return parsed.String()
}

func (a *App) captureAuthFromResponse(projectID, endpointID, accountID string, resp *httpclient.Response) {
	if resp == nil || resp.Status >= 400 {
		return
	}

	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		log.Printf("capture: project nil err=%v", err)
		return
	}
	if project.LoginEndpointID == "" || project.LoginEndpointID != endpointID {
		return
	}

	driver, err := a.scanner.ResolveByName(project.Framework)
	if err != nil {
		driver = nil
	}
	cap, ok := authCapabilityFor(driver)
	if !ok {
		return
	}

	authResp := core.AuthResponse{
		Status:  resp.Status,
		Headers: toHTTPHeader(resp.Headers),
		Body:    []byte(resp.Body),
	}
	extraction, ok := cap.ExtractCredentials(authResp)
	if !ok || extraction == nil {
		extraction = &core.AuthExtraction{}
	}
	if project.LoginTokenPath != "" {
		if token, path, found := extractTokenAtPath(authResp.Body, project.LoginTokenPath); found {
			extraction.Token = token
			extraction.TokenPath = path
		}
	}
	if extraction.Token == "" && extraction.User == nil && len(extraction.Cookies) == 0 {
		return
	}

	rec := domain.ProjectAuth{
		ProjectID:            projectID,
		Scheme:               string(cap.DefaultScheme()),
		Token:                extraction.Token,
		TokenPath:            extraction.TokenPath,
		ExpiresAt:            extraction.ExpiresAt,
		CapturedFromEndpoint: endpointID,
	}
	if extraction.User != nil {
		if raw, err := json.Marshal(extraction.User); err == nil {
			rec.UserJSON = string(raw)
		}
	}
	if len(extraction.Cookies) > 0 {
		if raw, err := json.Marshal(extraction.Cookies); err == nil {
			rec.CookiesJSON = string(raw)
		}
	}
	if err := a.auth.Save(a.ctx, rec); err != nil {
		log.Printf("save project auth: %v", err)
	} else {
		log.Printf("capture: saved token_len=%d user=%v", len(rec.Token), extraction.User)
	}

	// Mirror token into the active account so multi-account setups stay in sync.
	if a.accounts != nil && a.vault != nil && extraction.Token != "" {
		if acc, err := a.resolveActiveAccount(projectID, accountID); err == nil && acc != nil {
			if tokenEnc, err := a.vault.Encrypt(extraction.Token); err == nil {
				acc.TokenEnc = tokenEnc
				acc.Scheme = string(cap.DefaultScheme())
				acc.TokenPath = extraction.TokenPath
				acc.ExpiresAt = extraction.ExpiresAt
				acc.UserJSON = rec.UserJSON
				acc.CookiesJSON = rec.CookiesJSON
				if err := a.accounts.Save(a.ctx, *acc); err != nil {
					log.Printf("save account token: %v", err)
				}
			}
		}
	}
}

func (a *App) applyAuthScheme(projectID string, headers map[string]string, rec *domain.ProjectAuth) {
	driver := a.driverForProject(projectID)
	cap, ok := authCapabilityFor(driver)
	scheme := core.AuthScheme(rec.Scheme)
	if scheme == "" && ok {
		scheme = cap.DefaultScheme()
	}
	if !ok || cap == nil {
		// no driver support — fall back to bearer if scheme matches or empty
		if scheme == core.AuthSchemeBearer || scheme == core.AuthSchemeNone {
			headers["Authorization"] = "Bearer " + rec.Token
		}
		return
	}
	req, err := http.NewRequest("GET", "http://placeholder.local/", nil)
	if err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	cap.ApplyAuth(req, core.AuthContext{
		Scheme: scheme,
		Token:  rec.Token,
	})
	for k := range req.Header {
		headers[k] = req.Header.Get(k)
	}
}

func authCapabilityFor(driver core.FrameworkDriver) (core.AuthCapable, bool) {
	if driver == nil {
		return nil, false
	}
	if cap, ok := driver.(core.AuthCapable); ok {
		return cap, true
	}
	return nil, false
}

func extractTokenAtPath(body []byte, path string) (string, string, bool) {
	if len(body) == 0 || path == "" {
		return "", "", false
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", "", false
	}
	parts := strings.Split(path, ".")
	cur := payload
	for _, p := range parts {
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", "", false
		}
		v, exists := obj[p]
		if !exists {
			return "", "", false
		}
		cur = v
	}
	s, ok := cur.(string)
	if !ok || s == "" {
		return "", "", false
	}
	return s, path, true
}

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/drivers/laravel"
	"spectra-desktop/internal/httpclient"
	"spectra-desktop/internal/repository"
	"spectra-desktop/internal/storage"
	"spectra-desktop/internal/watcher"
	"spectra-desktop/internal/workspace"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	scanner   *core.Scanner
	workspace *workspace.Service
	storage   *storage.Storage
	projects  domain.ProjectRepository
	settings  domain.SettingsRepository
	endpoints domain.EndpointRepository
	auth      domain.AuthRepository
	http      *httpclient.Client
	watcher   *watcher.Watcher
}

func New() (*App, error) {
	scanner := core.NewScanner()
	scanner.Register(laravel.New())

	store := storage.New()
	if err := store.Open(""); err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("migrate storage: %w", err)
	}

	return &App{
		scanner:   scanner,
		workspace: workspace.NewService(),
		storage:   store,
		projects:  repository.NewProjectRepository(store.DB),
		settings:  repository.NewSettingsRepository(store.DB),
		endpoints: repository.NewEndpointRepository(store.DB),
		auth:      repository.NewAuthRepository(store.DB),
		http:      httpclient.New(),
		watcher:   watcher.New(),
	}, nil
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(_ context.Context) {
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			log.Printf("close storage: %v", err)
		}
	}
}

func (a *App) OpenProject(path string) (*workspace.Workspace, error) {
	return a.workspace.Open(path)
}

func (a *App) DetectFramework(path string) (string, core.DetectionResult, error) {
	driver, result, err := a.scanner.Resolve(path)
	if err != nil {
		return "", core.DetectionResult{}, err
	}
	return driver.Name(), result, nil
}

func (a *App) SelectProjectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select Project Folder",
		CanCreateDirectories: false,
	})
}

type ProjectInfo struct {
	Path             string               `json:"path"`
	Name             string               `json:"name"`
	Framework        string               `json:"framework"`
	FrameworkVersion string               `json:"frameworkVersion"`
	Detection        core.DetectionResult `json:"detection"`
	APIDetection     APIDetection         `json:"apiDetection"`
	DefaultBaseURL   string               `json:"defaultBaseUrl"`
	DefaultPorts     []int                `json:"defaultPorts,omitempty"`
}

type APIDetection struct {
	Mode       string `json:"mode"`
	Value      string `json:"value"`
	Count      int    `json:"count"`
	TotalCount int    `json:"totalCount"`
	ScanError  string `json:"scanError,omitempty"`
}

func (a *App) InspectProject(path string) (ProjectInfo, error) {
	ws, err := a.workspace.Open(path)
	if err != nil {
		return ProjectInfo{}, err
	}
	info := ProjectInfo{
		Path:      ws.Path,
		Name:      ws.Name,
		Framework: "other",
	}
	driver, det, err := a.scanner.Resolve(ws.Path)
	if err != nil {
		return info, nil
	}
	info.Framework = driver.Name()
	info.Detection = det
	defaults := driver.Defaults()
	info.DefaultBaseURL = defaults.BaseURL
	info.DefaultPorts = defaults.Ports

	endpoints, scanErr := driver.Scan(a.ctx, ws.Path)
	if scanErr != nil {
		info.APIDetection = APIDetection{
			Mode:      core.FilterModeAuto,
			ScanError: scanErr.Error(),
		}
		return info, nil
	}
	result := core.ApplyFilter(endpoints, core.FilterModeAuto, "")
	info.APIDetection = APIDetection{
		Mode:       result.Mode,
		Value:      result.Value,
		Count:      len(result.Endpoints),
		TotalCount: len(endpoints),
	}
	return info, nil
}

func (a *App) PreviewAPIRoutes(path, mode, value string) (APIDetection, error) {
	driver, _, err := a.scanner.Resolve(path)
	if err != nil {
		return APIDetection{}, err
	}
	endpoints, err := driver.Scan(a.ctx, path)
	if err != nil {
		return APIDetection{Mode: mode, Value: value, ScanError: err.Error()}, nil
	}
	result := core.ApplyFilter(endpoints, mode, value)
	return APIDetection{
		Mode:       result.Mode,
		Value:      result.Value,
		Count:      len(result.Endpoints),
		TotalCount: len(endpoints),
	}, nil
}

func (a *App) Drivers() []string {
	names := make([]string, 0)
	for _, d := range a.scanner.Drivers() {
		names = append(names, d.Name())
	}
	return names
}

func (a *App) ListProjects() ([]domain.Project, error) {
	return a.projects.List(a.ctx)
}

func (a *App) SaveProject(input domain.ProjectInput) (*domain.Project, error) {
	if strings.TrimSpace(input.BaseURL) == "" {
		if driver, _, err := a.scanner.Resolve(input.Path); err == nil {
			input.BaseURL = driver.Defaults().BaseURL
		}
	}
	return a.projects.Save(a.ctx, input)
}

func (a *App) UpdateProjectBaseURL(projectID, baseURL string) error {
	return a.projects.UpdateBaseURL(a.ctx, projectID, baseURL)
}

type ProjectAuthState struct {
	ProjectID            string         `json:"projectID"`
	Scheme               string         `json:"scheme"`
	HasToken             bool           `json:"hasToken"`
	TokenPreview         string         `json:"tokenPreview,omitempty"`
	TokenPath            string         `json:"tokenPath,omitempty"`
	User                 *core.AuthUser `json:"user,omitempty"`
	HasCookies           bool           `json:"hasCookies"`
	ExpiresAt            *time.Time     `json:"expiresAt,omitempty"`
	CapturedFromEndpoint string         `json:"capturedFromEndpoint,omitempty"`
	CapturedAt           time.Time      `json:"capturedAt"`
}

func (a *App) GetProjectAuth(projectID string) (*ProjectAuthState, error) {
	rec, err := a.auth.Get(a.ctx, projectID)
	if err != nil || rec == nil {
		return nil, err
	}
	state := &ProjectAuthState{
		ProjectID:            rec.ProjectID,
		Scheme:               rec.Scheme,
		HasToken:             rec.Token != "",
		TokenPath:            rec.TokenPath,
		HasCookies:           rec.CookiesJSON != "" && rec.CookiesJSON != "[]",
		ExpiresAt:            rec.ExpiresAt,
		CapturedFromEndpoint: rec.CapturedFromEndpoint,
		CapturedAt:           rec.CapturedAt,
	}
	if rec.Token != "" {
		state.TokenPreview = previewToken(rec.Token)
	}
	if rec.UserJSON != "" {
		var user core.AuthUser
		if err := json.Unmarshal([]byte(rec.UserJSON), &user); err == nil {
			state.User = &user
		}
	}
	return state, nil
}

func (a *App) ClearProjectAuth(projectID string) error {
	return a.auth.Clear(a.ctx, projectID)
}

type SetProjectAuthInput struct {
	ProjectID string `json:"projectID"`
	Scheme    string `json:"scheme"`
	Token     string `json:"token"`
}

func (a *App) SetProjectAuthManual(input SetProjectAuthInput) error {
	if input.ProjectID == "" {
		return fmt.Errorf("project id required")
	}
	scheme := input.Scheme
	if scheme == "" {
		scheme = string(core.AuthSchemeBearer)
	}
	rec := domain.ProjectAuth{
		ProjectID: input.ProjectID,
		Scheme:    scheme,
		Token:     strings.TrimSpace(input.Token),
	}
	return a.auth.Save(a.ctx, rec)
}

type SetEndpointAuthInput struct {
	EndpointID string `json:"endpointID"`
	Role       string `json:"role"`
	TokenPath  string `json:"tokenPath"`
}

func (a *App) SetEndpointAuth(input SetEndpointAuthInput) error {
	if input.EndpointID == "" {
		return fmt.Errorf("endpoint id required")
	}
	return a.endpoints.UpdateAuthOverride(a.ctx, input.EndpointID, core.AuthRole(input.Role), input.TokenPath)
}

func previewToken(token string) string {
	if len(token) <= 12 {
		return token
	}
	return token[:6] + "…" + token[len(token)-4:]
}

type ExecuteRequestInput struct {
	ProjectID  string            `json:"projectID"`
	EndpointID string            `json:"endpointID,omitempty"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	BaseURL    string            `json:"baseUrl,omitempty"`
	TimeoutMs  int               `json:"timeoutMs,omitempty"`
	SkipAuth   bool              `json:"skipAuth,omitempty"`
}

func (a *App) ExecuteRequest(input ExecuteRequestInput) (*httpclient.Response, error) {
	baseURL := strings.TrimSpace(input.BaseURL)
	if baseURL == "" && input.ProjectID != "" {
		project, err := a.projects.GetByID(a.ctx, input.ProjectID)
		if err != nil {
			return nil, err
		}
		baseURL = strings.TrimSpace(project.BaseURL)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("missing base url")
	}
	target, err := joinURL(baseURL, input.Path)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(input.TimeoutMs) * time.Millisecond

	headers := input.Headers
	var cookies []http.Cookie
	if !input.SkipAuth && input.ProjectID != "" {
		merged, ck := a.applyProjectAuth(input.ProjectID, headers)
		headers = merged
		cookies = ck
	}

	resp, err := a.http.Send(a.ctx, httpclient.Request{
		Method:  input.Method,
		URL:     target,
		Headers: headers,
		Body:    input.Body,
		Cookies: cookies,
		Timeout: timeout,
	})
	if err != nil {
		return resp, err
	}

	if input.ProjectID != "" && input.EndpointID != "" {
		a.captureAuthFromResponse(input.ProjectID, input.EndpointID, resp)
	}
	return resp, nil
}

func (a *App) applyProjectAuth(projectID string, base map[string]string) (map[string]string, []http.Cookie) {
	rec, err := a.auth.Get(a.ctx, projectID)
	if err != nil || rec == nil {
		return base, nil
	}
	headers := map[string]string{}
	for k, v := range base {
		headers[k] = v
	}
	if rec.Token != "" {
		if _, exists := headers["Authorization"]; !exists {
			scheme := rec.Scheme
			if scheme == "" || scheme == string(core.AuthSchemeBearer) {
				headers["Authorization"] = "Bearer " + rec.Token
			}
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
	return headers, cookies
}

func (a *App) captureAuthFromResponse(projectID, endpointID string, resp *httpclient.Response) {
	if resp == nil || resp.Status >= 400 {
		return
	}
	driver, err := a.scanner.ResolveByName(a.scannerNameFor(projectID))
	if err != nil || driver == nil {
		driver = nil
	}

	ep, err := a.endpoints.GetByID(a.ctx, endpointID)
	if err != nil || ep == nil {
		return
	}
	role := ep.EffectiveAuthRole()
	if role != core.AuthRoleLogin && role != core.AuthRoleRefresh {
		return
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
	if ep.TokenPathOverride != "" {
		if token, path, found := extractTokenAtPath(authResp.Body, ep.TokenPathOverride); found {
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
	}
}

func (a *App) scannerNameFor(projectID string) string {
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return ""
	}
	return project.Framework
}

func authCapabilityFor(driver core.FrameworkDriver) (core.AuthCapable, bool) {
	if driver == nil {
		return laravel.AuthCapability{}, true
	}
	if cap, ok := driver.(core.AuthCapable); ok {
		return cap, true
	}
	if driver.Name() == laravel.DriverName {
		return laravel.AuthCapability{}, true
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

func toHTTPHeader(in map[string][]string) http.Header {
	h := http.Header{}
	for k, vs := range in {
		for _, v := range vs {
			h.Add(k, v)
		}
	}
	return h
}

func joinURL(base, path string) (string, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	path = strings.TrimSpace(path)
	if base == "" {
		return "", fmt.Errorf("empty base url")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	if path == "" {
		return u.String(), nil
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	return u.String(), nil
}

func (a *App) DeleteProject(id string) error {
	return a.projects.Delete(a.ctx, id)
}

func (a *App) MarkProjectSynced(id string) error {
	return a.projects.MarkSynced(a.ctx, id)
}

func (a *App) GetActiveProjectID() (string, error) {
	return a.settings.Get(a.ctx, domain.SettingActiveProjectID)
}

func (a *App) SetActiveProjectID(id string) error {
	if id == "" {
		return a.settings.Delete(a.ctx, domain.SettingActiveProjectID)
	}
	return a.settings.Set(a.ctx, domain.SettingActiveProjectID, id)
}

func (a *App) ScanWorkspace(projectID string) ([]core.Endpoint, error) {
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	driver, _, err := a.scanner.Resolve(project.Path)
	if err != nil {
		return nil, err
	}
	all, err := driver.Scan(a.ctx, project.Path)
	if err != nil {
		return nil, err
	}
	filtered := core.ApplyFilter(all, project.APIFilterMode, project.APIFilterValue).Endpoints
	if err := a.endpoints.Replace(a.ctx, projectID, filtered); err != nil {
		return nil, fmt.Errorf("persist endpoints: %w", err)
	}
	return filtered, nil
}

func (a *App) ScanActiveProject() ([]core.Endpoint, error) {
	id, err := a.settings.Get(a.ctx, domain.SettingActiveProjectID)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return []core.Endpoint{}, nil
	}
	return a.ScanWorkspace(id)
}

func (a *App) ListEndpoints(projectID string) ([]core.Endpoint, error) {
	if projectID == "" {
		return []core.Endpoint{}, nil
	}
	return a.endpoints.List(a.ctx, projectID)
}

func (a *App) GetProjectStats(projectID string) (domain.ProjectStats, error) {
	if projectID == "" {
		return domain.ProjectStats{}, nil
	}
	return a.endpoints.Stats(a.ctx, projectID)
}

func (a *App) DetectProject(id string) (core.DetectionResult, error) {
	project, err := a.projects.GetByID(a.ctx, id)
	if err != nil {
		return core.DetectionResult{}, err
	}
	for _, d := range a.scanner.Drivers() {
		result := d.Detect(project.Path)
		if result.Detected {
			return result, nil
		}
	}
	return core.DetectionResult{}, nil
}

package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/drivers/laravel"
	"spectra-desktop/internal/assertions"
	"spectra-desktop/internal/exporter/openapi"
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
	history   domain.HistoryRepository
	envs      domain.EnvironmentRepository
	snapshots domain.SnapshotRepository
	tests     domain.TestRepository
	captures  domain.CaptureRepository
	captured  *capturedStore
	collections domain.CollectionRepository
	datasets  *repository.DatasetRepository
	metrics   *repository.MetricsRepository
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

	a := &App{
		scanner:   scanner,
		workspace: workspace.NewService(),
		storage:   store,
		projects:  repository.NewProjectRepository(store.DB),
		settings:  repository.NewSettingsRepository(store.DB),
		endpoints: repository.NewEndpointRepository(store.DB),
		auth:      repository.NewAuthRepository(store.DB),
		history:   repository.NewHistoryRepository(store.DB),
		envs:      repository.NewEnvironmentRepository(store.DB),
		snapshots: repository.NewSnapshotRepository(store.DB),
		tests:     repository.NewTestRepository(store.DB),
		captures:  repository.NewCaptureRepository(store.DB),
		collections: repository.NewCollectionRepository(store.DB),
		datasets:  repository.NewDatasetRepository(store.DB),
		metrics:   repository.NewMetricsRepository(store.DB),
		http:      httpclient.New(),
		watcher:   watcher.New(),
	}
	a.captured = newCapturedStore(repository.NewCapturedValuesRepository(store.DB), func() context.Context {
		if a.ctx != nil {
			return a.ctx
		}
		return context.Background()
	})
	return a, nil
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	go a.migrateAuthRolesIfNeeded()
}

func (a *App) migrateAuthRolesIfNeeded() {
	projects, err := a.projects.List(a.ctx)
	if err != nil {
		return
	}
	for _, p := range projects {
		eps, err := a.endpoints.List(a.ctx, p.ID)
		if err != nil || len(eps) == 0 {
			continue
		}
		hasRole := false
		for _, e := range eps {
			if e.AuthRole != "" {
				hasRole = true
				break
			}
		}
		if hasRole {
			continue
		}
		log.Printf("auth migrate: rescanning project %s (%s)", p.Name, p.ID)
		if _, err := a.ScanWorkspace(p.ID); err != nil {
			log.Printf("auth migrate scan failed %s: %v", p.ID, err)
		}
	}
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
	Mode        string `json:"mode"`
	Value       string `json:"value"`
	Count       int    `json:"count"`
	TotalCount  int    `json:"totalCount"`
	ScanError   string `json:"scanError,omitempty"`
	LoginRoute  string `json:"loginRoute,omitempty"`
	LogoutRoute string `json:"logoutRoute,omitempty"`
}

func summarizeAuth(endpoints []core.Endpoint) (string, string) {
	type cand struct {
		ep    core.Endpoint
		score int
	}
	var bestLogin, bestLogout cand
	for _, ep := range endpoints {
		if ep.AuthRole == core.AuthRoleLogin {
			score := scoreAuthPath(string(ep.Path), core.AuthRoleLogin) + 1
			if score > bestLogin.score {
				bestLogin = cand{ep: ep, score: score}
			}
		} else if ep.AuthRole == core.AuthRoleLogout {
			score := scoreAuthPath(string(ep.Path), core.AuthRoleLogout) + 1
			if score > bestLogout.score {
				bestLogout = cand{ep: ep, score: score}
			}
		}
	}
	loginRoute := ""
	if bestLogin.score > 0 {
		loginRoute = string(bestLogin.ep.Method) + " " + bestLogin.ep.Path
	}
	logoutRoute := ""
	if bestLogout.score > 0 {
		logoutRoute = string(bestLogout.ep.Method) + " " + bestLogout.ep.Path
	}
	return loginRoute, logoutRoute
}

func scoreAuthPath(path string, role core.AuthRole) int {
	p := strings.ToLower(path)
	switch role {
	case core.AuthRoleLogin:
		switch {
		case strings.HasSuffix(p, "/login"):
			return 5
		case strings.Contains(p, "/login"):
			return 3
		case strings.Contains(p, "/signin") || strings.Contains(p, "/sign-in"):
			return 2
		case strings.Contains(p, "/authenticate") || strings.Contains(p, "/auth/token"):
			return 2
		}
	case core.AuthRoleLogout:
		switch {
		case strings.HasSuffix(p, "/logout"):
			return 5
		case strings.Contains(p, "/logout"):
			return 3
		case strings.Contains(p, "/signout") || strings.Contains(p, "/sign-out"):
			return 2
		}
	}
	return 0
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
	if envURL := readDotenvAppURL(ws.Path); envURL != "" {
		info.DefaultBaseURL = envURL
	}
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
	loginRoute, logoutRoute := summarizeAuth(result.Endpoints)
	info.APIDetection = APIDetection{
		Mode:        result.Mode,
		Value:       result.Value,
		Count:       len(result.Endpoints),
		TotalCount:  len(endpoints),
		LoginRoute:  loginRoute,
		LogoutRoute: logoutRoute,
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
	loginRoute, logoutRoute := summarizeAuth(result.Endpoints)
	return APIDetection{
		Mode:        result.Mode,
		Value:       result.Value,
		Count:       len(result.Endpoints),
		TotalCount:  len(endpoints),
		LoginRoute:  loginRoute,
		LogoutRoute: logoutRoute,
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
	vars := a.resolveEnvVars(input.ProjectID)
	resolvedPath := substituteVars(input.Path, vars)
	target, err := joinURL(baseURL, resolvedPath)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(input.TimeoutMs) * time.Millisecond

	headers := substituteHeaderVars(input.Headers, vars)
	body := substituteVars(input.Body, vars)
	var cookies []http.Cookie
	if !input.SkipAuth && input.ProjectID != "" {
		merged, ck := a.applyProjectAuth(input.ProjectID, headers)
		headers = merged
		cookies = ck
	}

	resp, sendErr := a.http.Send(a.ctx, httpclient.Request{
		Method:  input.Method,
		URL:     target,
		Headers: headers,
		Body:    body,
		Cookies: cookies,
		Timeout: timeout,
	})

	var testResults []assertions.Result
	if input.ProjectID != "" && resp != nil && sendErr == nil {
		testResults = a.runTestsForRequest(input, resp)
		a.runCapturesForRequest(input, resp)
	}

	if input.ProjectID != "" {
		a.saveHistory(input, target, headers, resp, sendErr, testResults)
	}

	if sendErr != nil {
		return resp, sendErr
	}

	if input.ProjectID != "" && input.EndpointID != "" {
		if a.isLogoutEndpoint(input.ProjectID, input.EndpointID) && resp.Status < 400 {
			if err := a.auth.Clear(a.ctx, input.ProjectID); err != nil {
				log.Printf("clear auth on logout: %v", err)
			}
		} else {
			a.captureAuthFromResponse(input.ProjectID, input.EndpointID, resp)
		}
	}
	return resp, nil
}

const historyFreeLimit = 5000

func (a *App) saveHistory(input ExecuteRequestInput, url string, headers map[string]string, resp *httpclient.Response, sendErr error, testResults []assertions.Result) {
	headersJSON, _ := json.Marshal(headers)
	respHeaders := ""
	respBody := ""
	respStatus := 0
	durationMs := 0
	sizeBytes := 0
	errStr := ""
	if resp != nil {
		respStatus = resp.Status
		respBody = resp.Body
		durationMs = int(resp.DurationMs)
		sizeBytes = resp.SizeBytes
		if rh, err := json.Marshal(resp.Headers); err == nil {
			respHeaders = string(rh)
		}
	}
	if sendErr != nil {
		errStr = sendErr.Error()
	}
	testResultsJSON := ""
	if len(testResults) > 0 {
		if buf, err := json.Marshal(testResults); err == nil {
			testResultsJSON = string(buf)
		}
	}
	entry := domain.HistoryEntry{
		ProjectID:       input.ProjectID,
		EndpointID:      input.EndpointID,
		Method:          input.Method,
		URL:             url,
		RequestHeaders:  string(headersJSON),
		RequestBody:     input.Body,
		ResponseStatus:  respStatus,
		ResponseHeaders: respHeaders,
		ResponseBody:    respBody,
		DurationMs:      durationMs,
		SizeBytes:       sizeBytes,
		Error:           errStr,
		TestResultsJSON: testResultsJSON,
	}
	if err := a.history.Save(a.ctx, entry); err != nil {
		log.Printf("save history: %v", err)
		return
	}
	if err := a.history.TrimOldest(a.ctx, input.ProjectID, historyFreeLimit); err != nil {
		log.Printf("trim history: %v", err)
	}
}

type HistoryListItem struct {
	ID             string    `json:"id"`
	EndpointID     string    `json:"endpointID,omitempty"`
	Method         string    `json:"method"`
	URL            string    `json:"url"`
	ResponseStatus int       `json:"responseStatus"`
	DurationMs     int       `json:"durationMs"`
	SizeBytes      int       `json:"sizeBytes"`
	Error          string    `json:"error,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type HistoryEntryDetail struct {
	HistoryListItem
	RequestHeaders  string          `json:"requestHeaders"`
	RequestBody     string          `json:"requestBody"`
	ResponseHeaders string          `json:"responseHeaders"`
	ResponseBody    string          `json:"responseBody"`
	TestResults     []TestResultDTO `json:"testResults,omitempty"`
}

func (a *App) ListHistory(projectID string, limit int) ([]HistoryListItem, error) {
	if projectID == "" {
		return []HistoryListItem{}, nil
	}
	if limit <= 0 {
		limit = 100
	}
	entries, err := a.history.List(a.ctx, projectID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]HistoryListItem, 0, len(entries))
	for _, e := range entries {
		out = append(out, HistoryListItem{
			ID:             e.ID,
			EndpointID:     e.EndpointID,
			Method:         e.Method,
			URL:            e.URL,
			ResponseStatus: e.ResponseStatus,
			DurationMs:     e.DurationMs,
			SizeBytes:      e.SizeBytes,
			Error:          e.Error,
			CreatedAt:      e.CreatedAt,
		})
	}
	return out, nil
}

func (a *App) GetHistoryEntry(id string) (*HistoryEntryDetail, error) {
	entry, err := a.history.GetByID(a.ctx, id)
	if err != nil || entry == nil {
		return nil, err
	}
	return &HistoryEntryDetail{
		HistoryListItem: HistoryListItem{
			ID:             entry.ID,
			EndpointID:     entry.EndpointID,
			Method:         entry.Method,
			URL:            entry.URL,
			ResponseStatus: entry.ResponseStatus,
			DurationMs:     entry.DurationMs,
			SizeBytes:      entry.SizeBytes,
			Error:          entry.Error,
			CreatedAt:      entry.CreatedAt,
		},
		RequestHeaders:  entry.RequestHeaders,
		RequestBody:     entry.RequestBody,
		ResponseHeaders: entry.ResponseHeaders,
		ResponseBody:    entry.ResponseBody,
		TestResults:     parseTestResults(entry.TestResultsJSON),
	}, nil
}

func parseTestResults(raw string) []TestResultDTO {
	if raw == "" {
		return nil
	}
	var out []TestResultDTO
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func (a *App) ClearHistory(projectID string) error {
	return a.history.Clear(a.ctx, projectID)
}

type EnvironmentDTO struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"projectID"`
	Name      string            `json:"name"`
	Vars      map[string]string `json:"vars"`
	SortOrder int               `json:"sortOrder"`
}

type SaveEnvironmentInput struct {
	ID        string            `json:"id,omitempty"`
	ProjectID string            `json:"projectID"`
	Name      string            `json:"name"`
	Vars      map[string]string `json:"vars,omitempty"`
	SortOrder int               `json:"sortOrder,omitempty"`
}

func (a *App) ListEnvironments(projectID string) ([]EnvironmentDTO, error) {
	if projectID == "" {
		return []EnvironmentDTO{}, nil
	}
	envs, err := a.envs.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]EnvironmentDTO, 0, len(envs))
	for _, e := range envs {
		out = append(out, envToDTO(e))
	}
	return out, nil
}

func (a *App) SaveEnvironment(input SaveEnvironmentInput) (*EnvironmentDTO, error) {
	if input.ProjectID == "" {
		return nil, fmt.Errorf("project id required")
	}
	if input.Vars == nil {
		input.Vars = map[string]string{}
	}
	env, err := a.envs.Save(a.ctx, domain.EnvironmentInput{
		ID:        input.ID,
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Vars:      input.Vars,
		SortOrder: input.SortOrder,
	})
	if err != nil || env == nil {
		return nil, err
	}
	dto := envToDTO(*env)
	return &dto, nil
}

func (a *App) DeleteEnvironment(id string) error {
	return a.envs.Delete(a.ctx, id)
}

func (a *App) SetActiveEnvironment(projectID, envID string) error {
	return a.projects.UpdateActiveEnvironment(a.ctx, projectID, envID)
}

func (a *App) resolveEnvVars(projectID string) map[string]string {
	out := map[string]string{}
	if projectID == "" {
		return out
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err == nil && project != nil && project.ActiveEnvironmentID != "" {
		if env, err := a.envs.GetByID(a.ctx, project.ActiveEnvironmentID); err == nil && env != nil {
			for k, v := range env.Vars {
				out[k] = v
			}
		}
	}
	if a.captured != nil {
		a.captured.ensureLoaded(projectID)
		for k, v := range a.captured.values(projectID) {
			out[k] = v
		}
	}
	return out
}

var varPattern = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.\-]+)\s*\}\}`)

func substituteVars(input string, vars map[string]string) string {
	if input == "" || len(vars) == 0 {
		return input
	}
	return varPattern.ReplaceAllStringFunc(input, func(match string) string {
		groups := varPattern.FindStringSubmatch(match)
		if len(groups) < 2 {
			return match
		}
		key := groups[1]
		if v, ok := vars[key]; ok {
			return v
		}
		return match
	})
}

func substituteHeaderVars(headers map[string]string, vars map[string]string) map[string]string {
	if len(headers) == 0 {
		return headers
	}
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		resolvedKey := substituteVars(k, vars)
		if !isValidHeaderName(resolvedKey) {
			continue
		}
		out[resolvedKey] = substituteVars(v, vars)
	}
	return out
}

func isValidHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)):
		default:
			return false
		}
	}
	return true
}

type snapshotEndpoint struct {
	Method       string                  `json:"method"`
	Path         string                  `json:"path"`
	Handler      string                  `json:"handler,omitempty"`
	Middleware   []string                `json:"middleware,omitempty"`
	AuthRole     string                  `json:"authRole,omitempty"`
	SchemaHash   string                  `json:"schemaHash,omitempty"`
	SchemaFields []snapshotSchemaField   `json:"schemaFields,omitempty"`
}

type snapshotSchemaField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
}

func (a *App) recordSnapshot(projectID string, endpoints []core.Endpoint) error {
	items := make([]snapshotEndpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, snapshotEndpoint{
			Method:       string(ep.Method),
			Path:         ep.Path,
			Handler:      ep.Handler,
			Middleware:   ep.Middleware,
			AuthRole:     string(ep.AuthRole),
			SchemaHash:   stableSchemaHash(ep.RequestSchema),
			SchemaFields: extractSchemaFields(ep.RequestSchema),
		})
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	hash := hashString(string(payload))

	latest, _ := a.snapshots.Latest(a.ctx, projectID)
	if latest != nil && latest.Hash == hash {
		return nil
	}

	now := time.Now().UTC()
	snapshot := domain.EndpointSnapshot{
		ProjectID:     projectID,
		Hash:          hash,
		PayloadJSON:   string(payload),
		EndpointCount: len(endpoints),
		ScannedAt:     now,
	}
	if err := a.snapshots.Save(a.ctx, snapshot); err != nil {
		return err
	}
	return a.snapshots.TrimOldest(a.ctx, projectID, 50)
}

func hashString(s string) string {
	if s == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func extractSchemaFields(raw string) []snapshotSchemaField {
	if raw == "" {
		return nil
	}
	var s struct {
		Fields []struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil
	}
	out := make([]snapshotSchemaField, 0, len(s.Fields))
	for _, f := range s.Fields {
		out = append(out, snapshotSchemaField{Name: f.Name, Type: f.Type, Required: f.Required})
	}
	return out
}

// stableSchemaHash strips per-scan noise (gofakeit examples) so a
// schema with the same shape produces the same hash across scans.
func stableSchemaHash(raw string) string {
	if raw == "" {
		return ""
	}
	var s struct {
		Source     string `json:"source"`
		Confidence string `json:"confidence"`
		Fields     []struct {
			Name     string   `json:"name"`
			Type     string   `json:"type"`
			Required bool     `json:"required"`
			Rules    []string `json:"rules,omitempty"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return hashString(raw)
	}
	canonical, err := json.Marshal(s)
	if err != nil {
		return hashString(raw)
	}
	return hashString(string(canonical))
}

type SnapshotSummary struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"projectID"`
	EndpointCount int       `json:"endpointCount"`
	ScannedAt     time.Time `json:"scannedAt"`
	Added         int       `json:"added"`
	Removed       int       `json:"removed"`
	Changed       int       `json:"changed"`
}

type SnapshotDiffEntry struct {
	Method   string             `json:"method"`
	Path     string             `json:"path"`
	Kind     string             `json:"kind"`
	Changes  []string           `json:"changes,omitempty"`
	AuthRole string             `json:"authRole,omitempty"`
	Handler  string             `json:"handler,omitempty"`
	Previous *snapshotEndpoint  `json:"previous,omitempty"`
	Current  *snapshotEndpoint  `json:"current,omitempty"`
}

type SnapshotDiff struct {
	ID         string              `json:"id"`
	ScannedAt  time.Time           `json:"scannedAt"`
	PreviousID string              `json:"previousID,omitempty"`
	Added      []SnapshotDiffEntry `json:"added"`
	Removed    []SnapshotDiffEntry `json:"removed"`
	Changed    []SnapshotDiffEntry `json:"changed"`
}

func (a *App) ListSnapshots(projectID string, limit int) ([]SnapshotSummary, error) {
	if projectID == "" {
		return []SnapshotSummary{}, nil
	}
	if limit <= 0 {
		limit = 50
	}
	snaps, err := a.snapshots.List(a.ctx, projectID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SnapshotSummary, 0, len(snaps))
	for i, s := range snaps {
		summary := SnapshotSummary{
			ID:            s.ID,
			ProjectID:     s.ProjectID,
			EndpointCount: s.EndpointCount,
			ScannedAt:     s.ScannedAt,
		}
		if i+1 < len(snaps) {
			diff, err := computeDiff(snaps[i+1].PayloadJSON, getSnapshotPayload(a, s.ID))
			if err == nil {
				summary.Added = len(diff.Added)
				summary.Removed = len(diff.Removed)
				summary.Changed = len(diff.Changed)
			}
		}
		out = append(out, summary)
	}
	return out, nil
}

func getSnapshotPayload(a *App, id string) string {
	s, err := a.snapshots.GetByID(a.ctx, id)
	if err != nil || s == nil {
		return ""
	}
	return s.PayloadJSON
}

func (a *App) GetSnapshotDiff(snapshotID string) (*SnapshotDiff, error) {
	current, err := a.snapshots.GetByID(a.ctx, snapshotID)
	if err != nil || current == nil {
		return nil, err
	}
	previous, err := a.snapshots.Predecessor(a.ctx, current.ProjectID, current.ScannedAt)
	if err != nil {
		return nil, err
	}
	prevPayload := ""
	prevID := ""
	if previous != nil {
		prevPayload = previous.PayloadJSON
		prevID = previous.ID
	}
	diff, err := computeDiff(prevPayload, current.PayloadJSON)
	if err != nil {
		return nil, err
	}
	return &SnapshotDiff{
		ID:         current.ID,
		ScannedAt:  current.ScannedAt,
		PreviousID: prevID,
		Added:      ensureSlice(diff.Added),
		Removed:    ensureSlice(diff.Removed),
		Changed:    ensureSlice(diff.Changed),
	}, nil
}

func ensureSlice(in []SnapshotDiffEntry) []SnapshotDiffEntry {
	if in == nil {
		return []SnapshotDiffEntry{}
	}
	return in
}

func computeDiff(previousJSON, currentJSON string) (struct {
	Added   []SnapshotDiffEntry
	Removed []SnapshotDiffEntry
	Changed []SnapshotDiffEntry
}, error) {
	var out struct {
		Added   []SnapshotDiffEntry
		Removed []SnapshotDiffEntry
		Changed []SnapshotDiffEntry
	}
	prev, err := decodeSnapshotPayload(previousJSON)
	if err != nil {
		return out, err
	}
	cur, err := decodeSnapshotPayload(currentJSON)
	if err != nil {
		return out, err
	}
	prevMap := indexSnapshot(prev)
	curMap := indexSnapshot(cur)

	for key, ep := range curMap {
		old, exists := prevMap[key]
		if !exists {
			added := snapshotDiffEntry(ep, "added", nil)
			cur := ep
			added.Current = &cur
			out.Added = append(out.Added, added)
			continue
		}
		changes := compareEndpoint(old, ep)
		if len(changes) > 0 {
			entry := snapshotDiffEntry(ep, "changed", changes)
			prev := old
			cur := ep
			entry.Previous = &prev
			entry.Current = &cur
			out.Changed = append(out.Changed, entry)
		}
	}
	for key, ep := range prevMap {
		if _, exists := curMap[key]; !exists {
			removed := snapshotDiffEntry(ep, "removed", nil)
			prev := ep
			removed.Previous = &prev
			out.Removed = append(out.Removed, removed)
		}
	}
	return out, nil
}

func decodeSnapshotPayload(s string) ([]snapshotEndpoint, error) {
	if s == "" {
		return nil, nil
	}
	var items []snapshotEndpoint
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func indexSnapshot(items []snapshotEndpoint) map[string]snapshotEndpoint {
	out := make(map[string]snapshotEndpoint, len(items))
	for _, ep := range items {
		key := ep.Method + " " + ep.Path
		out[key] = ep
	}
	return out
}

func compareEndpoint(a, b snapshotEndpoint) []string {
	changes := []string{}
	if a.Handler != b.Handler {
		changes = append(changes, "handler")
	}
	if a.AuthRole != b.AuthRole {
		changes = append(changes, "authRole")
	}
	if a.SchemaHash != b.SchemaHash {
		changes = append(changes, "schema")
	}
	if !stringSliceEqual(a.Middleware, b.Middleware) {
		changes = append(changes, "middleware")
	}
	return changes
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func snapshotDiffEntry(ep snapshotEndpoint, kind string, changes []string) SnapshotDiffEntry {
	return SnapshotDiffEntry{
		Method:   ep.Method,
		Path:     ep.Path,
		Kind:     kind,
		Changes:  changes,
		AuthRole: ep.AuthRole,
		Handler:  ep.Handler,
	}
}

func (a *App) ExportOpenAPI(projectID string) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("project id required")
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return "", err
	}
	endpoints, err := a.endpoints.List(a.ctx, projectID)
	if err != nil {
		return "", err
	}
	spec := openapi.Build(project, endpoints)
	return openapi.ToJSON(spec)
}

func (a *App) SaveOpenAPIToFile(projectID string) (string, error) {
	content, err := a.ExportOpenAPI(projectID)
	if err != nil {
		return "", err
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return "", err
	}
	defaultName := strings.ReplaceAll(strings.ToLower(project.Name), " ", "-") + "-openapi.json"
	target, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save OpenAPI spec",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON files", Pattern: "*.json"},
		},
	})
	if err != nil || target == "" {
		return "", err
	}
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		return "", err
	}
	return target, nil
}

type RegenerateFieldInput struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Rules []string `json:"rules,omitempty"`
}

type RegenerateBodyInput struct {
	Body   string                 `json:"body"`
	Fields []RegenerateFieldInput `json:"fields,omitempty"`
}

func (a *App) RegenerateExampleBody(endpointID string) (string, error) {
	if endpointID == "" {
		return "{}", nil
	}
	ep, err := a.endpoints.GetByID(a.ctx, endpointID)
	if err != nil || ep == nil || ep.RequestSchema == "" {
		return "{}", err
	}
	var raw struct {
		Fields []RegenerateFieldInput `json:"fields"`
	}
	if err := json.Unmarshal([]byte(ep.RequestSchema), &raw); err != nil {
		return "{}", err
	}
	return regenerateFromFields(raw.Fields)
}

func (a *App) RegenerateBodyValues(input RegenerateBodyInput) (string, error) {
	body := strings.TrimSpace(input.Body)
	if body == "" || body == "{}" {
		return regenerateFromFields(input.Fields)
	}
	current := orderedMap{}
	if err := json.Unmarshal([]byte(body), &current); err != nil {
		return regenerateFromFields(input.Fields)
	}
	if len(current.Keys) == 0 {
		return regenerateFromFields(input.Fields)
	}
	fieldByName := map[string]RegenerateFieldInput{}
	for _, f := range input.Fields {
		fieldByName[f.Name] = f
	}
	out := orderedMap{}
	for _, key := range current.Keys {
		oldVal := current.Values[key]
		var inferredType string
		var rules []string
		if f, ok := fieldByName[key]; ok {
			inferredType = f.Type
			rules = f.Rules
		} else {
			inferredType = inferTypeFromValue(oldVal)
		}
		out.Set(key, laravel.RegenerateValue(key, inferredType, rules))
	}
	return marshalOrdered(out)
}

func regenerateFromFields(fields []RegenerateFieldInput) (string, error) {
	if len(fields) == 0 {
		return "{}", nil
	}
	out := orderedMap{}
	for _, f := range fields {
		out.Set(f.Name, laravel.RegenerateValue(f.Name, f.Type, f.Rules))
	}
	return marshalOrdered(out)
}

type orderedMap struct {
	Keys   []string
	Values map[string]any
}

func (m *orderedMap) Set(k string, v any) {
	if m.Values == nil {
		m.Values = map[string]any{}
	}
	if _, exists := m.Values[k]; !exists {
		m.Keys = append(m.Keys, k)
	}
	m.Values[k] = v
}

func (m *orderedMap) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected object")
	}
	m.Values = map[string]any{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("expected string key")
		}
		var v any
		if err := dec.Decode(&v); err != nil {
			return err
		}
		m.Keys = append(m.Keys, key)
		m.Values[key] = v
	}
	return nil
}

func marshalOrdered(m orderedMap) (string, error) {
	var b strings.Builder
	b.WriteString("{\n")
	for i, key := range m.Keys {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString("  ")
		keyJSON, _ := json.Marshal(key)
		b.Write(keyJSON)
		b.WriteString(": ")
		valJSON, err := json.MarshalIndent(m.Values[key], "  ", "  ")
		if err != nil {
			return "", err
		}
		b.Write(valJSON)
	}
	b.WriteString("\n}")
	return b.String(), nil
}

func inferTypeFromValue(v any) string {
	switch t := v.(type) {
	case bool:
		return "boolean"
	case json.Number:
		if _, err := t.Int64(); err == nil {
			return "integer"
		}
		return "numeric"
	case float64:
		return "numeric"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case nil:
		return "string"
	default:
		return "string"
	}
}

type DashboardMetrics struct {
	StatusBuckets []StatusBucketDTO   `json:"statusBuckets"`
	Latency       LatencyDTO          `json:"latency"`
	Volume        []VolumePoint       `json:"volume"`
	TotalRuns     int                 `json:"totalRuns"`
	ErrorRate     float64             `json:"errorRate"`
	TopSlow       []EndpointMetricDTO `json:"topSlow"`
	TopFailing    []EndpointMetricDTO `json:"topFailing"`
	TopUsed       []EndpointMetricDTO `json:"topUsed"`
}

type StatusBucketDTO struct {
	Bucket string `json:"bucket"`
	Count  int    `json:"count"`
}

type LatencyDTO struct {
	Count int `json:"count"`
	Avg   int `json:"avg"`
	Min   int `json:"min"`
	Max   int `json:"max"`
	P50   int `json:"p50"`
	P95   int `json:"p95"`
	P99   int `json:"p99"`
}

type VolumePoint struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type EndpointMetricDTO struct {
	EndpointID string  `json:"endpointID"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Count      int     `json:"count"`
	Errors     int     `json:"errors"`
	AvgMs      int     `json:"avgMs"`
	ErrorRate  float64 `json:"errorRate"`
}

func (a *App) GetDashboardMetrics(projectID string, volumeDays int) (*DashboardMetrics, error) {
	if projectID == "" {
		return nil, nil
	}
	if volumeDays <= 0 {
		volumeDays = 7
	}
	out := &DashboardMetrics{
		StatusBuckets: []StatusBucketDTO{},
		Volume:        []VolumePoint{},
		TopSlow:       []EndpointMetricDTO{},
		TopFailing:    []EndpointMetricDTO{},
		TopUsed:       []EndpointMetricDTO{},
	}

	buckets, err := a.metrics.StatusBuckets(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	total := 0
	errs := 0
	for _, b := range buckets {
		out.StatusBuckets = append(out.StatusBuckets, StatusBucketDTO{Bucket: b.Bucket, Count: b.Count})
		total += b.Count
		if b.Bucket == "4xx" || b.Bucket == "5xx" || b.Bucket == "err" {
			errs += b.Count
		}
	}
	out.TotalRuns = total
	if total > 0 {
		out.ErrorRate = float64(errs) / float64(total)
	}

	lat, err := a.metrics.LatencyStats(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out.Latency = LatencyDTO{
		Count: lat.Count,
		Avg:   lat.Avg,
		Min:   lat.Min,
		Max:   lat.Max,
		P50:   lat.P50,
		P95:   lat.P95,
		P99:   lat.P99,
	}

	volume, err := a.metrics.DailyVolume(a.ctx, projectID, volumeDays)
	if err != nil {
		return nil, err
	}
	for _, v := range volume {
		out.Volume = append(out.Volume, VolumePoint{Day: v.Day.Format("2006-01-02"), Count: v.Count})
	}

	endpointStats, err := a.metrics.EndpointMetrics(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	all := make([]EndpointMetricDTO, 0, len(endpointStats))
	for _, e := range endpointStats {
		rate := 0.0
		if e.Count > 0 {
			rate = float64(e.Errors) / float64(e.Count)
		}
		all = append(all, EndpointMetricDTO{
			EndpointID: e.EndpointID,
			Method:     e.Method,
			Path:       e.Path,
			Count:      e.Count,
			Errors:     e.Errors,
			AvgMs:      e.AvgMs,
			ErrorRate:  rate,
		})
	}
	out.TopSlow = topNBy(all, func(a, b EndpointMetricDTO) bool { return a.AvgMs > b.AvgMs }, 5)
	out.TopUsed = topNBy(all, func(a, b EndpointMetricDTO) bool { return a.Count > b.Count }, 5)
	failing := []EndpointMetricDTO{}
	for _, e := range all {
		if e.Errors > 0 {
			failing = append(failing, e)
		}
	}
	out.TopFailing = topNBy(failing, func(a, b EndpointMetricDTO) bool { return a.ErrorRate > b.ErrorRate }, 5)

	return out, nil
}

type LatencyPointDTO struct {
	Day   string `json:"day"`
	AvgMs int    `json:"avgMs"`
	Count int    `json:"count"`
}

type EndpointLatencySeriesDTO struct {
	EndpointID string            `json:"endpointID"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	AvgMs      int               `json:"avgMs"`
	Points     []LatencyPointDTO `json:"points"`
}

type HourlyCellDTO struct {
	Day   int `json:"day"`
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type FlakyEndpointDTO struct {
	EndpointID string  `json:"endpointID"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Total      int     `json:"total"`
	Successes  int     `json:"successes"`
	Failures   int     `json:"failures"`
	FlakeScore float64 `json:"flakeScore"`
}

type EndpointUsageSeriesDTO struct {
	EndpointID string            `json:"endpointID"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Total      int               `json:"total"`
	Points     []LatencyPointDTO `json:"points"`
}

type EndpointFailureSeriesDTO struct {
	EndpointID string            `json:"endpointID"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Failures   int               `json:"failures"`
	Points     []LatencyPointDTO `json:"points"`
}

type InsightsDTO struct {
	LatencyOverTime  []EndpointLatencySeriesDTO `json:"latencyOverTime"`
	UsageOverTime    []EndpointUsageSeriesDTO   `json:"usageOverTime"`
	FailuresOverTime []EndpointFailureSeriesDTO `json:"failuresOverTime"`
	HourlyHeatmap    []HourlyCellDTO            `json:"hourlyHeatmap"`
	Flaky            []FlakyEndpointDTO         `json:"flaky"`
}

func (a *App) GetInsights(projectID string, days int) (*InsightsDTO, error) {
	if projectID == "" {
		return &InsightsDTO{LatencyOverTime: []EndpointLatencySeriesDTO{}, HourlyHeatmap: []HourlyCellDTO{}, Flaky: []FlakyEndpointDTO{}}, nil
	}
	if days <= 0 {
		days = 7
	}
	out := &InsightsDTO{
		LatencyOverTime:  []EndpointLatencySeriesDTO{},
		UsageOverTime:    []EndpointUsageSeriesDTO{},
		FailuresOverTime: []EndpointFailureSeriesDTO{},
		HourlyHeatmap:    []HourlyCellDTO{},
		Flaky:            []FlakyEndpointDTO{},
	}
	if series, err := a.metrics.LatencyOverTime(a.ctx, projectID, days, 5); err == nil {
		for _, s := range series {
			pts := make([]LatencyPointDTO, 0, len(s.Points))
			for _, p := range s.Points {
				pts = append(pts, LatencyPointDTO{Day: p.Day.Format("2006-01-02"), AvgMs: p.AvgMs, Count: p.Count})
			}
			out.LatencyOverTime = append(out.LatencyOverTime, EndpointLatencySeriesDTO{
				EndpointID: s.EndpointID, Method: s.Method, Path: s.Path, AvgMs: s.AvgMs, Points: pts,
			})
		}
	}
	if series, err := a.metrics.UsageOverTime(a.ctx, projectID, days, 5); err == nil {
		for _, s := range series {
			pts := make([]LatencyPointDTO, 0, len(s.Points))
			for _, p := range s.Points {
				pts = append(pts, LatencyPointDTO{Day: p.Day.Format("2006-01-02"), Count: p.Count})
			}
			out.UsageOverTime = append(out.UsageOverTime, EndpointUsageSeriesDTO{
				EndpointID: s.EndpointID, Method: s.Method, Path: s.Path, Total: s.Total, Points: pts,
			})
		}
	}
	if series, err := a.metrics.FailuresOverTime(a.ctx, projectID, days, 5); err == nil {
		for _, s := range series {
			pts := make([]LatencyPointDTO, 0, len(s.Points))
			for _, p := range s.Points {
				pts = append(pts, LatencyPointDTO{Day: p.Day.Format("2006-01-02"), Count: p.Count})
			}
			out.FailuresOverTime = append(out.FailuresOverTime, EndpointFailureSeriesDTO{
				EndpointID: s.EndpointID, Method: s.Method, Path: s.Path, Failures: s.Failures, Points: pts,
			})
		}
	}
	if cells, err := a.metrics.HourlyHeatmap(a.ctx, projectID, days); err == nil {
		for _, c := range cells {
			out.HourlyHeatmap = append(out.HourlyHeatmap, HourlyCellDTO{Day: c.Day, Hour: c.Hour, Count: c.Count})
		}
	}
	if flaky, err := a.metrics.FlakyEndpoints(a.ctx, projectID, 3); err == nil {
		for _, f := range flaky {
			out.Flaky = append(out.Flaky, FlakyEndpointDTO{
				EndpointID: f.EndpointID, Method: f.Method, Path: f.Path,
				Total: f.Total, Successes: f.Successes, Failures: f.Failures, FlakeScore: f.FlakeScore,
			})
		}
	}
	return out, nil
}

func topNBy(items []EndpointMetricDTO, less func(a, b EndpointMetricDTO) bool, n int) []EndpointMetricDTO {
	out := append([]EndpointMetricDTO(nil), items...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if less(out[j], out[i]) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}

func endpointTestKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func (a *App) runTestsForRequest(input ExecuteRequestInput, resp *httpclient.Response) []assertions.Result {
	if input.ProjectID == "" || input.Method == "" || input.Path == "" {
		return nil
	}
	key := endpointTestKey(input.Method, input.Path)
	tests, err := a.tests.List(a.ctx, input.ProjectID, key)
	if err != nil || len(tests) == 0 {
		return nil
	}
	domainTests := make([]assertions.Test, 0, len(tests))
	for _, t := range tests {
		domainTests = append(domainTests, assertions.Test{
			ID:       t.ID,
			Name:     t.Name,
			Kind:     t.Kind,
			JSONPath: t.JSONPath,
			Op:       t.Op,
			Expected: t.Expected,
		})
	}
	return assertions.Run(domainTests, assertions.ResponseSnapshot{
		Status:     resp.Status,
		Headers:    toHTTPHeader(resp.Headers),
		Body:       resp.Body,
		DurationMs: int(resp.DurationMs),
	})
}

type EndpointTestDTO struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Kind     string `json:"kind"`
	JSONPath string `json:"jsonPath,omitempty"`
	Op       string `json:"op,omitempty"`
	Expected string `json:"expected,omitempty"`
}

type TestResultDTO struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Pass    bool   `json:"pass"`
	Message string `json:"message,omitempty"`
}

type SaveTestsInput struct {
	ProjectID   string            `json:"projectID"`
	EndpointKey string            `json:"endpointKey"`
	Tests       []EndpointTestDTO `json:"tests"`
}

func (a *App) ListEndpointTests(projectID, endpointKey string) ([]EndpointTestDTO, error) {
	if projectID == "" || endpointKey == "" {
		return []EndpointTestDTO{}, nil
	}
	rows, err := a.tests.List(a.ctx, projectID, endpointKey)
	if err != nil {
		return nil, err
	}
	out := make([]EndpointTestDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, EndpointTestDTO{
			ID:       r.ID,
			Name:     r.Name,
			Kind:     r.Kind,
			JSONPath: r.JSONPath,
			Op:       r.Op,
			Expected: r.Expected,
		})
	}
	return out, nil
}

func (a *App) SaveEndpointTests(input SaveTestsInput) error {
	if input.ProjectID == "" || input.EndpointKey == "" {
		return fmt.Errorf("project id and endpoint key required")
	}
	tests := make([]domain.EndpointTest, 0, len(input.Tests))
	for i, t := range input.Tests {
		tests = append(tests, domain.EndpointTest{
			ID:        t.ID,
			Name:      t.Name,
			Kind:      t.Kind,
			JSONPath:  t.JSONPath,
			Op:        t.Op,
			Expected:  t.Expected,
			SortOrder: i,
		})
	}
	return a.tests.Replace(a.ctx, input.ProjectID, input.EndpointKey, tests)
}

type EndpointCaptureDTO struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name"`
	Source string `json:"source"`
	Path   string `json:"path"`
}

type SaveCapturesInput struct {
	ProjectID   string               `json:"projectID"`
	EndpointKey string               `json:"endpointKey"`
	Captures    []EndpointCaptureDTO `json:"captures"`
}

type CapturedValueDTO struct {
	Name           string `json:"name"`
	Value          string `json:"value"`
	EndpointKey    string `json:"endpointKey,omitempty"`
	CapturedAt     int64  `json:"capturedAt,omitempty"`
}

func (a *App) ListEndpointCaptures(projectID, endpointKey string) ([]EndpointCaptureDTO, error) {
	if projectID == "" || endpointKey == "" {
		return []EndpointCaptureDTO{}, nil
	}
	rows, err := a.captures.List(a.ctx, projectID, endpointKey)
	if err != nil {
		return nil, err
	}
	out := make([]EndpointCaptureDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, EndpointCaptureDTO{
			ID:     r.ID,
			Name:   r.Name,
			Source: r.Source,
			Path:   r.Path,
		})
	}
	return out, nil
}

func (a *App) SaveEndpointCaptures(input SaveCapturesInput) error {
	if input.ProjectID == "" || input.EndpointKey == "" {
		return fmt.Errorf("project id and endpoint key required")
	}
	captures := make([]domain.EndpointCapture, 0, len(input.Captures))
	keep := make(map[string]bool, len(input.Captures))
	for i, c := range input.Captures {
		captures = append(captures, domain.EndpointCapture{
			ID:        c.ID,
			Name:      c.Name,
			Source:    c.Source,
			Path:      c.Path,
			SortOrder: i,
		})
		if c.Name != "" {
			keep[c.Name] = true
		}
	}
	if err := a.captures.Replace(a.ctx, input.ProjectID, input.EndpointKey, captures); err != nil {
		return err
	}
	if a.captured != nil {
		a.captured.pruneByEndpoint(input.ProjectID, input.EndpointKey, keep)
	}
	return nil
}

func (a *App) ListCapturedValues(projectID string) []CapturedValueDTO {
	if projectID == "" || a.captured == nil {
		return []CapturedValueDTO{}
	}
	a.captured.ensureLoaded(projectID)
	return a.captured.list(projectID)
}

func (a *App) ClearCapturedValues(projectID string) {
	if projectID == "" || a.captured == nil {
		return
	}
	a.captured.clear(projectID)
}

func (a *App) runCapturesForRequest(input ExecuteRequestInput, resp *httpclient.Response) {
	if input.ProjectID == "" || a.captured == nil {
		return
	}
	key := endpointTestKey(input.Method, input.Path)
	rows, err := a.captures.List(a.ctx, input.ProjectID, key)
	if err != nil || len(rows) == 0 {
		return
	}
	var bodyValue any
	if resp != nil && resp.Body != "" {
		_ = json.Unmarshal([]byte(resp.Body), &bodyValue)
	}
	headers := toHTTPHeader(resp.Headers)
	for _, c := range rows {
		if c.Name == "" {
			continue
		}
		val, ok := extractCaptureValue(c.Source, c.Path, bodyValue, headers)
		if !ok {
			continue
		}
		a.captured.set(input.ProjectID, c.Name, val, key)
	}
}

type CollectionItemDTO struct {
	ID              string `json:"id,omitempty"`
	EndpointID      string `json:"endpointID"`
	BodyOverride    string `json:"bodyOverride,omitempty"`
	HeadersOverride string `json:"headersOverride,omitempty"`
	SkipOnFailure   bool   `json:"skipOnFailure,omitempty"`
	IterateDataset  bool   `json:"iterateDataset,omitempty"`
}

type CollectionDTO struct {
	ID          string              `json:"id"`
	ProjectID   string              `json:"projectID"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	SortOrder   int                 `json:"sortOrder"`
	Items       []CollectionItemDTO `json:"items"`
}

type SaveCollectionInput struct {
	ID          string              `json:"id,omitempty"`
	ProjectID   string              `json:"projectID"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	SortOrder   int                 `json:"sortOrder,omitempty"`
	Items       []CollectionItemDTO `json:"items"`
}

type CollectionRunItemDTO struct {
	EndpointID string            `json:"endpointID"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Status     int               `json:"status"`
	DurationMs int               `json:"durationMs"`
	Pass       bool              `json:"pass"`
	Skipped    bool              `json:"skipped,omitempty"`
	Error      string            `json:"error,omitempty"`
	TestResults []TestResultDTO  `json:"testResults,omitempty"`
}

type CollectionRunDTO struct {
	CollectionID string                 `json:"collectionID"`
	StartedAt    int64                  `json:"startedAt"`
	DurationMs   int                    `json:"durationMs"`
	PassCount    int                    `json:"passCount"`
	FailCount    int                    `json:"failCount"`
	SkipCount    int                    `json:"skipCount"`
	Items        []CollectionRunItemDTO `json:"items"`
}

func (a *App) ListCollections(projectID string) ([]CollectionDTO, error) {
	if projectID == "" {
		return []CollectionDTO{}, nil
	}
	rows, err := a.collections.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]CollectionDTO, 0, len(rows))
	for _, c := range rows {
		out = append(out, collectionToDTO(c))
	}
	return out, nil
}

func (a *App) SaveCollection(input SaveCollectionInput) (*CollectionDTO, error) {
	if input.ProjectID == "" || input.Name == "" {
		return nil, fmt.Errorf("project id and name required")
	}
	c := domain.Collection{
		ID:          input.ID,
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if c.ID == "" {
		created, err := a.collections.Create(a.ctx, c)
		if err != nil {
			return nil, err
		}
		c = *created
	} else {
		if err := a.collections.Update(a.ctx, c); err != nil {
			return nil, err
		}
	}
	items := make([]domain.CollectionItem, 0, len(input.Items))
	for i, it := range input.Items {
		items = append(items, domain.CollectionItem{
			ID:              it.ID,
			CollectionID:    c.ID,
			EndpointID:      it.EndpointID,
			SortOrder:       i,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
			SkipOnFailure:   it.SkipOnFailure,
			IterateDataset:  it.IterateDataset,
		})
	}
	if err := a.collections.ReplaceItems(a.ctx, c.ID, items); err != nil {
		return nil, err
	}
	full, err := a.collections.Get(a.ctx, c.ID)
	if err != nil || full == nil {
		return nil, err
	}
	dto := collectionToDTO(*full)
	return &dto, nil
}

type ExportedTest struct {
	Name     string `json:"name,omitempty"`
	Kind     string `json:"kind"`
	JSONPath string `json:"jsonPath,omitempty"`
	Op       string `json:"op,omitempty"`
	Expected string `json:"expected,omitempty"`
}

type ExportedCapture struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Path   string `json:"path"`
}

type ExportedItem struct {
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	IterateDataset  bool              `json:"iterateDataset,omitempty"`
	SkipOnFailure   bool              `json:"skipOnFailure,omitempty"`
	BodyOverride    string            `json:"bodyOverride,omitempty"`
	HeadersOverride string            `json:"headersOverride,omitempty"`
	Tests           []ExportedTest    `json:"tests,omitempty"`
	Captures        []ExportedCapture `json:"captures,omitempty"`
	Dataset         []json.RawMessage `json:"dataset,omitempty"`
}

type ExportedCollection struct {
	SpectraVersion string         `json:"spectraVersion"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	ExportedAt     int64          `json:"exportedAt"`
	Items          []ExportedItem `json:"items"`
}

func (a *App) ExportCollectionToFile(id string) (string, error) {
	json, err := a.ExportCollection(id)
	if err != nil {
		return "", err
	}
	c, _ := a.collections.Get(a.ctx, id)
	defaultName := "collection.spectra.json"
	if c != nil && c.Name != "" {
		defaultName = strings.ToLower(strings.ReplaceAll(c.Name, " ", "_")) + ".spectra.json"
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export collection",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON", Pattern: "*.json"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, []byte(json), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ExportCollection(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("id required")
	}
	c, err := a.collections.Get(a.ctx, id)
	if err != nil || c == nil {
		return "", fmt.Errorf("collection not found")
	}
	endpoints, err := a.endpoints.List(a.ctx, c.ProjectID)
	if err != nil {
		return "", err
	}
	byID := map[string]core.Endpoint{}
	for _, e := range endpoints {
		byID[e.ID] = e
	}
	out := ExportedCollection{
		SpectraVersion: "1",
		Name:           c.Name,
		Description:    c.Description,
		ExportedAt:     time.Now().UTC().Unix(),
		Items:          make([]ExportedItem, 0, len(c.Items)),
	}
	for _, it := range c.Items {
		ep, ok := byID[it.EndpointID]
		if !ok {
			continue
		}
		key := endpointTestKey(string(ep.Method), ep.Path)
		exported := ExportedItem{
			Method:          string(ep.Method),
			Path:            ep.Path,
			IterateDataset:  it.IterateDataset,
			SkipOnFailure:   it.SkipOnFailure,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
		}
		if tests, err := a.tests.List(a.ctx, c.ProjectID, key); err == nil {
			for _, t := range tests {
				exported.Tests = append(exported.Tests, ExportedTest{
					Name:     t.Name,
					Kind:     t.Kind,
					JSONPath: t.JSONPath,
					Op:       t.Op,
					Expected: t.Expected,
				})
			}
		}
		if caps, err := a.captures.List(a.ctx, c.ProjectID, key); err == nil {
			for _, cap := range caps {
				exported.Captures = append(exported.Captures, ExportedCapture{
					Name:   cap.Name,
					Source: cap.Source,
					Path:   cap.Path,
				})
			}
		}
		if it.IterateDataset {
			if rowsJSON, err := a.datasets.Get(a.ctx, c.ProjectID, key); err == nil {
				var rows []json.RawMessage
				if json.Unmarshal([]byte(rowsJSON), &rows) == nil {
					exported.Dataset = rows
				}
			}
		}
		out.Items = append(out.Items, exported)
	}
	buf, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

type ImportCollectionResult struct {
	Collection      CollectionDTO `json:"collection"`
	MissingEndpoints []string     `json:"missingEndpoints,omitempty"`
}

func (a *App) ImportCollection(projectID, payload string) (*ImportCollectionResult, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project id required")
	}
	if payload == "" {
		return nil, fmt.Errorf("payload required")
	}
	var imported ExportedCollection
	if err := json.Unmarshal([]byte(payload), &imported); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if imported.Name == "" {
		imported.Name = "Imported collection"
	}
	endpoints, err := a.endpoints.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	byKey := map[string]core.Endpoint{}
	for _, e := range endpoints {
		byKey[endpointTestKey(string(e.Method), e.Path)] = e
	}
	c := domain.Collection{
		ProjectID:   projectID,
		Name:        imported.Name,
		Description: imported.Description,
	}
	created, err := a.collections.Create(a.ctx, c)
	if err != nil {
		return nil, err
	}
	items := make([]domain.CollectionItem, 0, len(imported.Items))
	missing := []string{}
	for _, it := range imported.Items {
		key := endpointTestKey(it.Method, it.Path)
		ep, ok := byKey[key]
		if !ok {
			missing = append(missing, key)
			continue
		}
		items = append(items, domain.CollectionItem{
			CollectionID:    created.ID,
			EndpointID:      ep.ID,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
			SkipOnFailure:   it.SkipOnFailure,
			IterateDataset:  it.IterateDataset,
		})
		if len(it.Tests) > 0 {
			tests := make([]domain.EndpointTest, 0, len(it.Tests))
			for _, t := range it.Tests {
				tests = append(tests, domain.EndpointTest{
					Name:     t.Name,
					Kind:     t.Kind,
					JSONPath: t.JSONPath,
					Op:       t.Op,
					Expected: t.Expected,
				})
			}
			_ = a.tests.Replace(a.ctx, projectID, key, tests)
		}
		if len(it.Captures) > 0 {
			caps := make([]domain.EndpointCapture, 0, len(it.Captures))
			for _, cp := range it.Captures {
				caps = append(caps, domain.EndpointCapture{
					Name:   cp.Name,
					Source: cp.Source,
					Path:   cp.Path,
				})
			}
			_ = a.captures.Replace(a.ctx, projectID, key, caps)
		}
		if len(it.Dataset) > 0 {
			if buf, err := json.Marshal(it.Dataset); err == nil {
				_ = a.datasets.Save(a.ctx, projectID, key, string(buf))
			}
		}
	}
	if err := a.collections.ReplaceItems(a.ctx, created.ID, items); err != nil {
		return nil, err
	}
	full, err := a.collections.Get(a.ctx, created.ID)
	if err != nil || full == nil {
		return nil, err
	}
	dto := collectionToDTO(*full)
	return &ImportCollectionResult{Collection: dto, MissingEndpoints: missing}, nil
}

func (a *App) DeleteCollection(id string) error {
	if id == "" {
		return fmt.Errorf("id required")
	}
	return a.collections.Delete(a.ctx, id)
}

func (a *App) RunCollection(id string) (*CollectionRunDTO, error) {
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	c, err := a.collections.Get(a.ctx, id)
	if err != nil || c == nil {
		return nil, fmt.Errorf("collection not found")
	}
	endpoints, err := a.endpoints.List(a.ctx, c.ProjectID)
	if err != nil {
		return nil, err
	}
	byID := map[string]core.Endpoint{}
	for _, e := range endpoints {
		byID[e.ID] = e
	}
	run := &CollectionRunDTO{
		CollectionID: id,
		StartedAt:    time.Now().UTC().Unix(),
		Items:        make([]CollectionRunItemDTO, 0, len(c.Items)),
	}
	start := time.Now()
	skipRest := false
	total := len(c.Items)
	runtime.EventsEmit(a.ctx, "collection:run:start", map[string]any{
		"collectionID": id,
		"total":        total,
	})
	for idx, it := range c.Items {
		ep, ok := byID[it.EndpointID]
		if !ok {
			missing := CollectionRunItemDTO{
				EndpointID: it.EndpointID,
				Skipped:    true,
				Error:      "endpoint not found",
			}
			run.Items = append(run.Items, missing)
			run.SkipCount++
			runtime.EventsEmit(a.ctx, "collection:run:progress", map[string]any{
				"collectionID": id, "index": idx, "total": total, "item": missing,
			})
			continue
		}
		if skipRest {
			skipped := CollectionRunItemDTO{
				EndpointID: it.EndpointID,
				Method:     string(ep.Method),
				Path:       ep.Path,
				Skipped:    true,
				Error:      "skipped due to previous failure",
			}
			run.Items = append(run.Items, skipped)
			run.SkipCount++
			runtime.EventsEmit(a.ctx, "collection:run:progress", map[string]any{
				"collectionID": id, "index": idx, "total": total, "item": skipped,
			})
			continue
		}
		headers := map[string]string{}
		if it.HeadersOverride != "" {
			_ = json.Unmarshal([]byte(it.HeadersOverride), &headers)
		}
		bodies := []string{it.BodyOverride}
		if it.IterateDataset {
			rowsJSON, derr := a.datasets.Get(a.ctx, c.ProjectID, endpointTestKey(string(ep.Method), ep.Path))
			if derr == nil {
				var rows []json.RawMessage
				if json.Unmarshal([]byte(rowsJSON), &rows) == nil && len(rows) > 0 {
					bodies = make([]string, 0, len(rows))
					for _, r := range rows {
						bodies = append(bodies, string(r))
					}
				}
			}
		}
		anyFailed := false
		for bIdx, body := range bodies {
			input := ExecuteRequestInput{
				ProjectID:  c.ProjectID,
				EndpointID: ep.ID,
				Method:     string(ep.Method),
				Path:       ep.Path,
				Headers:    headers,
				Body:       body,
			}
			resp, sendErr := a.ExecuteRequest(input)
			item := CollectionRunItemDTO{
				EndpointID: ep.ID,
				Method:     string(ep.Method),
				Path:       ep.Path,
			}
			if len(bodies) > 1 {
				item.Path = fmt.Sprintf("%s [#%d]", ep.Path, bIdx+1)
			}
			if sendErr != nil {
				item.Error = sendErr.Error()
				item.Pass = false
			} else if resp != nil {
				item.Status = resp.Status
				item.DurationMs = int(resp.DurationMs)
				results := a.runTestsForRequest(input, resp)
				passed := true
				for _, r := range results {
					item.TestResults = append(item.TestResults, TestResultDTO{
						ID:      r.ID,
						Name:    r.Name,
						Kind:    r.Kind,
						Pass:    r.Pass,
						Message: r.Message,
					})
					if !r.Pass {
						passed = false
					}
				}
				item.Pass = passed && resp.Status < 400
			}
			if item.Pass {
				run.PassCount++
			} else {
				run.FailCount++
				anyFailed = true
			}
			run.Items = append(run.Items, item)
			runtime.EventsEmit(a.ctx, "collection:run:progress", map[string]any{
				"collectionID": id,
				"index":        idx,
				"total":        total,
				"item":         item,
			})
		}
		if anyFailed && it.SkipOnFailure {
			skipRest = true
		}
	}
	run.DurationMs = int(time.Since(start).Milliseconds())
	a.persistCollectionRun(id, run)
	runtime.EventsEmit(a.ctx, "collection:run:done", run)
	return run, nil
}

func (a *App) persistCollectionRun(collectionID string, run *CollectionRunDTO) {
	buf, err := json.Marshal(run)
	if err != nil {
		log.Printf("collection run marshal: %v", err)
		return
	}
	now := time.Now().UTC()
	_, err = a.storage.DB.NewRaw(
		`INSERT INTO collection_runs (collection_id, run_json, started_at, duration_ms, pass_count, fail_count, skip_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(collection_id) DO UPDATE SET
		   run_json = excluded.run_json,
		   started_at = excluded.started_at,
		   duration_ms = excluded.duration_ms,
		   pass_count = excluded.pass_count,
		   fail_count = excluded.fail_count,
		   skip_count = excluded.skip_count,
		   updated_at = excluded.updated_at`,
		collectionID, string(buf), time.Unix(run.StartedAt, 0).UTC(), run.DurationMs,
		run.PassCount, run.FailCount, run.SkipCount, now, now,
	).Exec(a.ctx)
	if err != nil {
		log.Printf("collection run persist: %v", err)
	}
}

func (a *App) GetLastCollectionRun(collectionID string) (*CollectionRunDTO, error) {
	if collectionID == "" {
		return nil, nil
	}
	var runJSON string
	err := a.storage.DB.NewRaw(
		`SELECT run_json FROM collection_runs WHERE collection_id = ?`,
		collectionID,
	).Scan(a.ctx, &runJSON)
	if err != nil {
		return nil, nil
	}
	if runJSON == "" {
		return nil, nil
	}
	var run CollectionRunDTO
	if err := json.Unmarshal([]byte(runJSON), &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (a *App) ListLastCollectionRuns(projectID string) (map[string]*CollectionRunDTO, error) {
	if projectID == "" {
		return map[string]*CollectionRunDTO{}, nil
	}
	type row struct {
		CollectionID string `bun:"collection_id"`
		RunJSON      string `bun:"run_json"`
	}
	var rows []row
	err := a.storage.DB.NewRaw(
		`SELECT cr.collection_id, cr.run_json FROM collection_runs cr
		 JOIN collections c ON c.id = cr.collection_id
		 WHERE c.project_id = ?`,
		projectID,
	).Scan(a.ctx, &rows)
	if err != nil {
		return map[string]*CollectionRunDTO{}, nil
	}
	out := map[string]*CollectionRunDTO{}
	for _, r := range rows {
		var run CollectionRunDTO
		if err := json.Unmarshal([]byte(r.RunJSON), &run); err == nil {
			out[r.CollectionID] = &run
		}
	}
	return out, nil
}

func collectionToDTO(c domain.Collection) CollectionDTO {
	items := make([]CollectionItemDTO, 0, len(c.Items))
	for _, it := range c.Items {
		items = append(items, CollectionItemDTO{
			ID:              it.ID,
			EndpointID:      it.EndpointID,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
			SkipOnFailure:   it.SkipOnFailure,
			IterateDataset:  it.IterateDataset,
		})
	}
	return CollectionDTO{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		Name:        c.Name,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		Items:       items,
	}
}

type DatasetRowResultDTO struct {
	Index      int    `json:"index"`
	Status     int    `json:"status"`
	DurationMs int    `json:"durationMs"`
	Pass       bool   `json:"pass"`
	Error      string `json:"error,omitempty"`
}

type DatasetRunDTO struct {
	EndpointKey string                `json:"endpointKey"`
	Total       int                   `json:"total"`
	PassCount   int                   `json:"passCount"`
	FailCount   int                   `json:"failCount"`
	DurationMs  int                   `json:"durationMs"`
	Rows        []DatasetRowResultDTO `json:"rows"`
}

func (a *App) GetEndpointDataset(projectID, endpointKey string) (string, error) {
	if projectID == "" || endpointKey == "" {
		return "[]", nil
	}
	return a.datasets.Get(a.ctx, projectID, endpointKey)
}

func (a *App) SaveEndpointDataset(projectID, endpointKey, rowsJSON string) error {
	if projectID == "" || endpointKey == "" {
		return fmt.Errorf("project id and endpoint key required")
	}
	if rowsJSON == "" {
		rowsJSON = "[]"
	}
	if !json.Valid([]byte(rowsJSON)) {
		return fmt.Errorf("invalid rows json")
	}
	return a.datasets.Save(a.ctx, projectID, endpointKey, rowsJSON)
}

func (a *App) GenerateDatasetRows(endpointID string, count int) (string, error) {
	if endpointID == "" {
		return "[]", nil
	}
	if count <= 0 {
		count = 1
	}
	if count > 500 {
		count = 500
	}
	ep, err := a.endpoints.GetByID(a.ctx, endpointID)
	if err != nil || ep == nil {
		return "[]", err
	}
	var raw struct {
		Fields []RegenerateFieldInput `json:"fields"`
	}
	if ep.RequestSchema != "" {
		_ = json.Unmarshal([]byte(ep.RequestSchema), &raw)
	}
	out := make([]json.RawMessage, 0, count)
	for i := 0; i < count; i++ {
		body, err := regenerateFromFields(raw.Fields)
		if err != nil {
			return "[]", err
		}
		out = append(out, json.RawMessage(body))
	}
	buf, err := json.Marshal(out)
	if err != nil {
		return "[]", err
	}
	return string(buf), nil
}

func (a *App) RunEndpointDataset(projectID, endpointID string) (*DatasetRunDTO, error) {
	if projectID == "" || endpointID == "" {
		return nil, fmt.Errorf("project id and endpoint id required")
	}
	ep, err := a.endpoints.GetByID(a.ctx, endpointID)
	if err != nil || ep == nil {
		return nil, fmt.Errorf("endpoint not found")
	}
	key := endpointTestKey(string(ep.Method), ep.Path)
	rowsJSON, err := a.datasets.Get(a.ctx, projectID, key)
	if err != nil {
		return nil, err
	}
	var rows []json.RawMessage
	if err := json.Unmarshal([]byte(rowsJSON), &rows); err != nil {
		return nil, fmt.Errorf("invalid dataset rows: %w", err)
	}
	run := &DatasetRunDTO{
		EndpointKey: key,
		Total:       len(rows),
		Rows:        make([]DatasetRowResultDTO, 0, len(rows)),
	}
	runtime.EventsEmit(a.ctx, "dataset:run:start", map[string]any{
		"endpointID": endpointID,
		"total":      len(rows),
	})
	start := time.Now()
	for i, row := range rows {
		input := ExecuteRequestInput{
			ProjectID:  projectID,
			EndpointID: endpointID,
			Method:     string(ep.Method),
			Path:       ep.Path,
			Body:       string(row),
		}
		resp, sendErr := a.ExecuteRequest(input)
		result := DatasetRowResultDTO{Index: i}
		if sendErr != nil {
			result.Error = sendErr.Error()
		} else if resp != nil {
			result.Status = resp.Status
			result.DurationMs = int(resp.DurationMs)
			result.Pass = resp.Status < 400
		}
		if result.Pass {
			run.PassCount++
		} else {
			run.FailCount++
		}
		run.Rows = append(run.Rows, result)
		runtime.EventsEmit(a.ctx, "dataset:run:progress", map[string]any{
			"endpointID": endpointID,
			"index":      i,
			"total":      len(rows),
			"row":        result,
		})
	}
	run.DurationMs = int(time.Since(start).Milliseconds())
	runtime.EventsEmit(a.ctx, "dataset:run:done", run)
	return run, nil
}

func extractCaptureValue(source, path string, body any, headers http.Header) (string, bool) {
	switch source {
	case "header":
		if v := headers.Get(path); v != "" {
			return v, true
		}
		return "", false
	case "body", "":
		v, ok := lookupJSONPath(body, path)
		if !ok {
			return "", false
		}
		return formatCapturedValue(v), true
	}
	return "", false
}

func formatCapturedValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	default:
		buf, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprintf("%v", x)
		}
		return string(buf)
	}
}

func lookupJSONPath(root any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" {
		return root, true
	}
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")
	cur := root
	tokens := splitJSONPath(path)
	for _, p := range tokens {
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[p]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, false
			}
			cur = v[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}

func splitJSONPath(path string) []string {
	if path == "" {
		return nil
	}
	out := []string{}
	cur := []byte{}
	flush := func() {
		if len(cur) > 0 {
			out = append(out, strings.Trim(string(cur), `"'`))
			cur = cur[:0]
		}
	}
	for i := 0; i < len(path); i++ {
		c := path[i]
		switch c {
		case '.', '[':
			flush()
		case ']':
			flush()
		default:
			cur = append(cur, c)
		}
	}
	flush()
	return out
}

func envToDTO(e domain.Environment) EnvironmentDTO {
	return EnvironmentDTO{
		ID:        e.ID,
		ProjectID: e.ProjectID,
		Name:      e.Name,
		Vars:      e.Vars,
		SortOrder: e.SortOrder,
	}
}

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

func (a *App) applyProjectAuth(projectID string, base map[string]string) (map[string]string, []http.Cookie) {
	rec, err := a.auth.Get(a.ctx, projectID)
	log.Printf("applyProjectAuth project=%s err=%v rec_nil=%t token_len=%d", projectID, err, rec == nil, func() int {
		if rec == nil {
			return 0
		}
		return len(rec.Token)
	}())
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

	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		log.Printf("capture: project nil err=%v", err)
		return
	}
	log.Printf("capture: project_login=%q endpoint=%q match=%t", project.LoginEndpointID, endpointID, project.LoginEndpointID == endpointID)
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

func readDotenvAppURL(projectPath string) string {
	candidates := []string{".env", ".env.local", ".env.example"}
	for _, name := range candidates {
		path := filepath.Join(projectPath, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if url := extractDotenvKey(string(data), "APP_URL"); url != "" {
			return url
		}
	}
	return ""
}

func extractDotenvKey(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.Index(trimmed, "=")
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(trimmed[:idx])
		if k != key {
			continue
		}
		v := strings.TrimSpace(trimmed[idx+1:])
		v = strings.Trim(v, `"'`)
		if v == "" {
			continue
		}
		return v
	}
	return ""
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
	if err := a.recordSnapshot(projectID, filtered); err != nil {
		log.Printf("record snapshot: %v", err)
	}
	if err := a.projects.MarkSynced(a.ctx, projectID); err != nil {
		log.Printf("mark synced: %v", err)
	}
	if project.LoginEndpointID == "" || project.LogoutEndpointID == "" {
		stored, err := a.endpoints.List(a.ctx, projectID)
		if err == nil {
			loginID := project.LoginEndpointID
			logoutID := project.LogoutEndpointID
			if loginID == "" {
				loginID = pickAuthEndpoint(stored, core.AuthRoleLogin)
			}
			if logoutID == "" {
				logoutID = pickAuthEndpoint(stored, core.AuthRoleLogout)
			}
			if loginID != project.LoginEndpointID || logoutID != project.LogoutEndpointID {
				if err := a.projects.UpdateAuthRoutes(a.ctx, projectID, loginID, logoutID, project.LoginTokenPath); err != nil {
					log.Printf("auto-set auth routes: %v", err)
				}
			}
		}
	}
	return filtered, nil
}

func pickAuthEndpoint(endpoints []core.Endpoint, target core.AuthRole) string {
	type candidate struct {
		id    string
		score int
	}
	var best candidate
	for _, ep := range endpoints {
		if ep.AuthRole != target {
			continue
		}
		score := 1
		path := strings.ToLower(ep.Path)
		switch target {
		case core.AuthRoleLogin:
			switch {
			case strings.HasSuffix(path, "/login"):
				score += 5
			case strings.Contains(path, "/login"):
				score += 3
			case strings.Contains(path, "/signin") || strings.Contains(path, "/sign-in"):
				score += 2
			case strings.Contains(path, "/authenticate") || strings.Contains(path, "/auth/token"):
				score += 2
			}
		case core.AuthRoleLogout:
			switch {
			case strings.HasSuffix(path, "/logout"):
				score += 5
			case strings.Contains(path, "/logout"):
				score += 3
			case strings.Contains(path, "/signout") || strings.Contains(path, "/sign-out"):
				score += 2
			}
		}
		if score > best.score {
			best = candidate{id: ep.ID, score: score}
		}
	}
	return best.id
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

func (a *App) GetProjectStatsReport(projectID string) (core.StatsReport, error) {
	if projectID == "" {
		return core.StatsReport{}, nil
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return core.StatsReport{}, err
	}
	driver, err := a.scanner.ResolveByName(project.Framework)
	if err != nil {
		return core.StatsReport{}, nil
	}
	cap, ok := driver.(core.StatsCapable)
	if !ok {
		return a.fallbackStatsReport(projectID)
	}
	endpoints, err := a.endpoints.List(a.ctx, projectID)
	if err != nil {
		return core.StatsReport{}, err
	}
	return cap.Stats(a.ctx, project.Path, endpoints)
}

func (a *App) fallbackStatsReport(projectID string) (core.StatsReport, error) {
	stats, err := a.endpoints.Stats(a.ctx, projectID)
	if err != nil {
		return core.StatsReport{}, err
	}
	return core.StatsReport{
		Cards: []core.StatCard{
			{Key: "routes", Kind: core.StatRoutes, Label: "Routes", Value: stats.Routes},
			{Key: "controllers", Kind: core.StatControllers, Label: "Controllers", Value: stats.Controllers},
			{Key: "middleware", Kind: core.StatMiddleware, Label: "Middleware", Value: stats.Middleware},
		},
	}, nil
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

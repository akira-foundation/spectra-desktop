package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
)

type LogEvent struct {
	Timestamp  time.Time `json:"timestamp"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	Status     int       `json:"status"`
	DurationMs int       `json:"durationMs"`
	Source     string    `json:"source"`
	EndpointID string    `json:"endpointId,omitempty"`
	BodySize   int       `json:"bodySize"`
}

type EventEmitter func(name string, data ...any)

type Manager struct {
	mu        sync.Mutex
	server    *http.Server
	listener  net.Listener
	projectID string
	port      int
	startedAt time.Time
	requests  atomic.Int64

	endpoints domain.EndpointRepository
	history   domain.HistoryRepository
	overrides domain.MockRepository
	emit      EventEmitter
}

func NewManager(
	endpoints domain.EndpointRepository,
	history domain.HistoryRepository,
	overrides domain.MockRepository,
	emit EventEmitter,
) *Manager {
	return &Manager{
		endpoints: endpoints,
		history:   history,
		overrides: overrides,
		emit:      emit,
	}
}

type Status struct {
	Running      bool      `json:"running"`
	ProjectID    string    `json:"projectId,omitempty"`
	Port         int       `json:"port,omitempty"`
	URL          string    `json:"url,omitempty"`
	StartedAt    time.Time `json:"startedAt,omitempty"`
	RequestCount int64     `json:"requestCount"`
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.server == nil {
		return Status{Running: false}
	}
	return Status{
		Running:      true,
		ProjectID:    m.projectID,
		Port:         m.port,
		URL:          fmt.Sprintf("http://localhost:%d", m.port),
		StartedAt:    m.startedAt,
		RequestCount: m.requests.Load(),
	}
}

func (m *Manager) Start(ctx context.Context, projectID string, port int) (Status, error) {
	if projectID == "" {
		return Status{}, errors.New("mock: project id required")
	}
	if err := m.Stop(); err != nil {
		return Status{}, err
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return Status{}, fmt.Errorf("mock: bind %s: %w", addr, err)
	}
	resolvedPort := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/", m.buildRequestHandler(ctx, projectID))

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	m.mu.Lock()
	m.server = server
	m.listener = listener
	m.projectID = projectID
	m.port = resolvedPort
	m.startedAt = time.Now().UTC()
	m.requests.Store(0)
	m.mu.Unlock()

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("mock: server error: %v\n", err)
		}
	}()

	return m.Status(), nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	server := m.server
	m.server = nil
	m.listener = nil
	m.projectID = ""
	m.port = 0
	m.mu.Unlock()
	if server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}

func (m *Manager) buildRequestHandler(ctx context.Context, projectID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		defer m.requests.Add(1)

		eps, err := m.endpoints.List(ctx, projectID)
		if err != nil {
			http.Error(w, "mock: load endpoints", http.StatusInternalServerError)
			return
		}

		match := findEndpointMatchingRequest(r.Method, r.URL.Path, eps)
		if match == nil {
			m.respondWithNoRouteMatch(w, r, started)
			return
		}

		response := m.resolveResponseForEndpoint(ctx, projectID, match)
		if response.LatencyMs > 0 {
			time.Sleep(time.Duration(response.LatencyMs) * time.Millisecond)
		}

		writeResolvedResponse(w, response)
		m.emitRequestLog(r, response, match, started)
	}
}

func (m *Manager) respondWithNoRouteMatch(w http.ResponseWriter, r *http.Request, started time.Time) {
	body := `{"error":"no matching mock route"}`
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_, _ = io.WriteString(w, body)
	if m.emit != nil {
		m.emit("mock:request", LogEvent{
			Timestamp:  started,
			Method:     r.Method,
			Path:       r.URL.Path,
			Status:     http.StatusNotFound,
			DurationMs: int(time.Since(started).Milliseconds()),
			Source:     string(domain.MockSourceNoMatch),
			BodySize:   len(body),
		})
	}
}

type resolvedResponse struct {
	Status    int
	Body      string
	Headers   map[string]string
	Source    domain.MockSource
	LatencyMs int
}

func (m *Manager) resolveResponseForEndpoint(ctx context.Context, projectID string, match *matchResult) resolvedResponse {
	override, _ := m.overrides.Get(ctx, projectID, match.endpoint.ID)
	if override != nil && !override.Enabled {
		return resolvedResponse{
			Status: http.StatusServiceUnavailable,
			Body:   `{"error":"endpoint disabled in mock"}`,
			Source: domain.MockSourceCustom,
		}
	}

	desiredSource := domain.MockSourceAuto
	customStatus := 0
	customLatency := 0
	customBody := ""
	customHeaders := map[string]string{}
	if override != nil {
		desiredSource = override.Source
		customStatus = override.Status
		customLatency = override.LatencyMs
		customBody = override.Body
		if override.HeadersJSON != "" {
			_ = json.Unmarshal([]byte(override.HeadersJSON), &customHeaders)
		}
	}

	if desiredSource == domain.MockSourceCustom || (customBody != "" && desiredSource != domain.MockSourceHistory && desiredSource != domain.MockSourceGenerated) {
		return resolvedResponse{
			Status:    statusOrDefault(customStatus, http.StatusOK),
			Body:      customBody,
			Headers:   withJSONContentType(customHeaders),
			Source:    domain.MockSourceCustom,
			LatencyMs: customLatency,
		}
	}

	if desiredSource == domain.MockSourceAuto || desiredSource == domain.MockSourceHistory {
		if entry, err := m.history.LatestSuccessByEndpoint(ctx, projectID, match.endpoint.ID); err == nil && entry != nil {
			headers := parseStoredResponseHeaders(entry.ResponseHeaders)
			return resolvedResponse{
				Status:    statusOrDefault(customStatus, entry.ResponseStatus),
				Body:      entry.ResponseBody,
				Headers:   withJSONContentType(headers),
				Source:    domain.MockSourceHistory,
				LatencyMs: customLatency,
			}
		}
	}

	fields := endpointSchemaFieldNames(match.endpoint)
	body := GenerateBody(string(match.endpoint.Method), match.endpoint.Path, fields, match.params)
	return resolvedResponse{
		Status:    statusOrDefault(customStatus, http.StatusOK),
		Body:      body,
		Headers:   withJSONContentType(customHeaders),
		Source:    domain.MockSourceGenerated,
		LatencyMs: customLatency,
	}
}

func (m *Manager) emitRequestLog(r *http.Request, response resolvedResponse, match *matchResult, started time.Time) {
	if m.emit == nil {
		return
	}
	m.emit("mock:request", LogEvent{
		Timestamp:  started,
		Method:     r.Method,
		Path:       r.URL.Path,
		Status:     response.Status,
		DurationMs: int(time.Since(started).Milliseconds()),
		Source:     string(response.Source),
		EndpointID: match.endpoint.ID,
		BodySize:   len(response.Body),
	})
}

func writeResolvedResponse(w http.ResponseWriter, response resolvedResponse) {
	for k, v := range response.Headers {
		w.Header().Set(k, v)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(response.Status)
	_, _ = io.WriteString(w, response.Body)
}

func withJSONContentType(h map[string]string) map[string]string {
	if h == nil {
		h = map[string]string{}
	}
	hasContentType := false
	for k := range h {
		if strings.EqualFold(k, "Content-Type") {
			hasContentType = true
			break
		}
	}
	if !hasContentType {
		h["Content-Type"] = "application/json"
	}
	return h
}

func parseStoredResponseHeaders(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var multi map[string][]string
	if err := json.Unmarshal([]byte(raw), &multi); err == nil {
		out := make(map[string]string, len(multi))
		for k, v := range multi {
			if len(v) > 0 {
				out[k] = v[0]
			}
		}
		return out
	}
	var single map[string]string
	if err := json.Unmarshal([]byte(raw), &single); err == nil {
		return single
	}
	return nil
}

func statusOrDefault(value, fallback int) int {
	if value > 0 {
		return value
	}
	if fallback > 0 {
		return fallback
	}
	return http.StatusOK
}

type matchResult struct {
	endpoint core.Endpoint
	params   map[string]string
}

var pathParamPattern = regexp.MustCompile(`\{([^}]+)\}`)

func findEndpointMatchingRequest(method, path string, eps []core.Endpoint) *matchResult {
	method = strings.ToUpper(method)
	for _, ep := range eps {
		if !strings.EqualFold(string(ep.Method), method) {
			continue
		}
		params, ok := matchPathTemplate(ep.Path, path)
		if ok {
			return &matchResult{endpoint: ep, params: params}
		}
	}
	return nil
}

func matchPathTemplate(template, actual string) (map[string]string, bool) {
	tParts := splitPathSegments(template)
	aParts := splitPathSegments(actual)
	if len(tParts) != len(aParts) {
		return nil, false
	}
	params := map[string]string{}
	for i, tp := range tParts {
		ap := aParts[i]
		if matches := pathParamPattern.FindStringSubmatch(tp); len(matches) == 2 {
			name := strings.TrimSuffix(matches[1], "?")
			params[name] = ap
			continue
		}
		if !strings.EqualFold(tp, ap) {
			return nil, false
		}
	}
	return params, true
}

func splitPathSegments(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func endpointSchemaFieldNames(ep core.Endpoint) []string {
	if ep.RequestSchema == "" {
		return nil
	}
	var generic struct {
		Fields []struct {
			Name string `json:"name"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(ep.RequestSchema), &generic); err != nil {
		return nil
	}
	out := make([]string, 0, len(generic.Fields))
	for _, f := range generic.Fields {
		if f.Name != "" {
			out = append(out, f.Name)
		}
	}
	return out
}

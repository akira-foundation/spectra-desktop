package mock

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"spectra-desktop/internal/domain"
)

type resolvedResponse struct {
	Status    int
	Body      string
	Headers   map[string]string
	Source    domain.MockSource
	LatencyMs int
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

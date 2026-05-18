package mock

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
)

func TestManager_Handler_NoMatchReturns404(t *testing.T) {
	events := make(chan capturedEvent, 4)
	m := newTestManager(nil, nil, nil, captureEmitter(events))
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	resp := doRequest(t, http.MethodGet, status.URL+"/missing")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	ev := waitForEvent(t, events)
	if ev.name != "mock:request" {
		t.Fatalf("event name = %q, want mock:request", ev.name)
	}
	log := ev.payload.(LogEvent)
	if log.Status != http.StatusNotFound || log.Path != "/missing" || log.Method != http.MethodGet {
		t.Fatalf("unexpected log event: %+v", log)
	}
	if log.Source != string(domain.MockSourceNoMatch) {
		t.Fatalf("source = %q, want %q", log.Source, domain.MockSourceNoMatch)
	}
}

func TestManager_Handler_CustomOverrideReturnsBodyAndHeaders(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodGet, Path: "/users"}}
	headers, _ := json.Marshal(map[string]string{"X-Custom": "yes", "Content-Type": "application/json"})
	overrides := map[string]*domain.MockOverride{
		"ep1": {
			ID:          "o1",
			EndpointID:  "ep1",
			Enabled:     true,
			Status:      http.StatusTeapot,
			Body:        `{"hello":"world"}`,
			HeadersJSON: string(headers),
			Source:      domain.MockSourceCustom,
		},
	}
	events := make(chan capturedEvent, 4)
	m := newTestManager(eps, overrides, nil, captureEmitter(events))
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	resp := doRequest(t, http.MethodGet, status.URL+"/users")
	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", resp.StatusCode)
	}
	if got := resp.Header.Get("X-Custom"); got != "yes" {
		t.Fatalf("X-Custom = %q, want yes", got)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"hello":"world"}` {
		t.Fatalf("body = %q", string(body))
	}

	ev := waitForEvent(t, events)
	log := ev.payload.(LogEvent)
	if log.Source != string(domain.MockSourceCustom) {
		t.Fatalf("source = %q, want %q", log.Source, domain.MockSourceCustom)
	}
	if log.EndpointID != "ep1" {
		t.Fatalf("EndpointID = %q, want ep1", log.EndpointID)
	}
}

func TestManager_Handler_DisabledOverrideReturns503(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodGet, Path: "/users"}}
	overrides := map[string]*domain.MockOverride{
		"ep1": {ID: "o1", EndpointID: "ep1", Enabled: false, Source: domain.MockSourceCustom},
	}
	m := newTestManager(eps, overrides, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	resp := doRequest(t, http.MethodGet, status.URL+"/users")
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", resp.StatusCode)
	}
}

func TestManager_Handler_LatencyMsRespected(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodGet, Path: "/slow"}}
	const latency = 120
	overrides := map[string]*domain.MockOverride{
		"ep1": {
			ID:         "o1",
			EndpointID: "ep1",
			Enabled:    true,
			Status:     http.StatusOK,
			Body:       `{"ok":true}`,
			LatencyMs:  latency,
			Source:     domain.MockSourceCustom,
		},
	}
	m := newTestManager(eps, overrides, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	start := time.Now()
	doRequest(t, http.MethodGet, status.URL+"/slow")
	elapsed := time.Since(start)

	if elapsed < time.Duration(latency)*time.Millisecond {
		t.Fatalf("elapsed = %v, want >= %dms", elapsed, latency)
	}
	if elapsed > time.Duration(latency+800)*time.Millisecond {
		t.Fatalf("elapsed = %v, exceeds jitter ceiling", elapsed)
	}
}

func TestManager_Handler_AutoSourceUsesHistoryWhenAvailable(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodGet, Path: "/users"}}
	history := map[string]*domain.HistoryEntry{
		"ep1": {
			ID:              "h1",
			EndpointID:      "ep1",
			ResponseStatus:  http.StatusCreated,
			ResponseBody:    `{"from":"history"}`,
			ResponseHeaders: `{"Content-Type":["application/json"]}`,
		},
	}
	events := make(chan capturedEvent, 4)
	m := newTestManager(eps, nil, history, captureEmitter(events))
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	resp := doRequest(t, http.MethodGet, status.URL+"/users")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"from":"history"}` {
		t.Fatalf("body = %q", string(body))
	}

	ev := waitForEvent(t, events)
	log := ev.payload.(LogEvent)
	if log.Source != string(domain.MockSourceHistory) {
		t.Fatalf("source = %q, want %q", log.Source, domain.MockSourceHistory)
	}
}

func TestManager_Handler_AutoSourceFallsBackToGenerated(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodGet, Path: "/users"}}
	events := make(chan capturedEvent, 4)
	m := newTestManager(eps, nil, nil, captureEmitter(events))
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	resp := doRequest(t, http.MethodGet, status.URL+"/users")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	ev := waitForEvent(t, events)
	log := ev.payload.(LogEvent)
	if log.Source != string(domain.MockSourceGenerated) {
		t.Fatalf("source = %q, want %q", log.Source, domain.MockSourceGenerated)
	}
}

func TestManager_Handler_PathTemplateMatches(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodGet, Path: "/users/{id}"}}
	m := newTestManager(eps, nil, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	resp := doRequest(t, http.MethodGet, status.URL+"/users/42")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestManager_Handler_MethodMismatchReturns404(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodPost, Path: "/users"}}
	m := newTestManager(eps, nil, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	resp := doRequest(t, http.MethodGet, status.URL+"/users")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

type capturedEvent struct {
	name    string
	payload any
}

func captureEmitter(ch chan capturedEvent) EventEmitter {
	return func(name string, data ...any) {
		var payload any
		if len(data) > 0 {
			payload = data[0]
		}
		select {
		case ch <- capturedEvent{name: name, payload: payload}:
		default:
		}
	}
}

func waitForEvent(t *testing.T, ch chan capturedEvent) capturedEvent {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	select {
	case ev := <-ch:
		return ev
	case <-ctx.Done():
		t.Fatalf("timed out waiting for event")
		return capturedEvent{}
	}
}

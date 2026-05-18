package mock

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
)

func newTestManager(eps []core.Endpoint, overrides map[string]*domain.MockOverride, history map[string]*domain.HistoryEntry, emit EventEmitter) *Manager {
	return NewManager(
		&stubEndpointRepo{endpoints: eps},
		&stubHistoryRepo{entries: history},
		&stubMockRepo{overrides: overrides},
		emit,
	)
}

func TestManager_Status_InitialStateNotRunning(t *testing.T) {
	m := newTestManager(nil, nil, nil, nil)

	got := m.Status()

	if got.Running {
		t.Fatalf("Running = true, want false")
	}
	if got.Port != 0 || got.URL != "" || got.ProjectID != "" {
		t.Fatalf("unexpected non-zero fields: %+v", got)
	}
}

func TestManager_Start_AssignsEphemeralPort(t *testing.T) {
	m := newTestManager(nil, nil, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	if !status.Running {
		t.Fatalf("Running = false, want true")
	}
	if status.Port == 0 {
		t.Fatalf("Port = 0, want ephemeral assignment")
	}
	if status.URL != fmt.Sprintf("http://localhost:%d", status.Port) {
		t.Fatalf("URL = %q, want http://localhost:%d", status.URL, status.Port)
	}
	if status.ProjectID != "proj-1" {
		t.Fatalf("ProjectID = %q, want proj-1", status.ProjectID)
	}
}

func TestManager_Start_EmptyProjectIDReturnsError(t *testing.T) {
	m := newTestManager(nil, nil, nil, nil)

	_, err := m.Start(context.Background(), "", 0)
	if err == nil {
		t.Fatalf("Start error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "project id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestManager_Stop_IdempotentWhenNotRunning(t *testing.T) {
	m := newTestManager(nil, nil, nil, nil)

	if err := m.Stop(); err != nil {
		t.Fatalf("first Stop error: %v", err)
	}
	if err := m.Stop(); err != nil {
		t.Fatalf("second Stop error: %v", err)
	}
	if m.Status().Running {
		t.Fatalf("Running = true after Stop, want false")
	}
}

func TestManager_Stop_AfterStartStopsServer(t *testing.T) {
	m := newTestManager(nil, nil, nil, nil)

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	if err := m.Stop(); err != nil {
		t.Fatalf("Stop error: %v", err)
	}

	if m.Status().Running {
		t.Fatalf("Running = true after Stop, want false")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, status.URL+"/anything", nil)
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		_ = resp.Body.Close()
		t.Fatalf("request succeeded after Stop, want failure")
	}
}

func TestManager_Start_RestartsCleanlyWhenAlreadyRunning(t *testing.T) {
	m := newTestManager(nil, nil, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	first, err := m.Start(context.Background(), "proj-a", 0)
	if err != nil {
		t.Fatalf("first Start error: %v", err)
	}

	second, err := m.Start(context.Background(), "proj-b", 0)
	if err != nil {
		t.Fatalf("second Start error: %v", err)
	}

	if second.ProjectID != "proj-b" {
		t.Fatalf("ProjectID = %q, want proj-b", second.ProjectID)
	}
	if first.Port == second.Port {
		t.Logf("note: ports matched (allowed but uncommon)")
	}
}

func TestManager_Start_PortAlreadyInUseReturnsError(t *testing.T) {
	holder, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = holder.Close() })

	port := holder.Addr().(*net.TCPAddr).Port

	m := newTestManager(nil, nil, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	_, err = m.Start(context.Background(), "proj-1", port)
	if err == nil {
		t.Fatalf("Start error = nil, want bind failure")
	}
	if !strings.Contains(err.Error(), "bind") {
		t.Fatalf("error = %v, want bind error", err)
	}
	if m.Status().Running {
		t.Fatalf("Running = true after failed Start, want false")
	}
}

func TestManager_Start_RequestIncrementsCount(t *testing.T) {
	eps := []core.Endpoint{{ID: "ep1", Method: core.MethodGet, Path: "/ping"}}
	m := newTestManager(eps, nil, nil, nil)
	t.Cleanup(func() { _ = m.Stop() })

	status, err := m.Start(context.Background(), "proj-1", 0)
	if err != nil {
		t.Fatalf("Start error: %v", err)
	}

	doRequest(t, http.MethodGet, status.URL+"/ping")
	doRequest(t, http.MethodGet, status.URL+"/ping")

	waitFor(t, 1*time.Second, func() bool {
		return m.Status().RequestCount >= 2
	})

	if got := m.Status().RequestCount; got < 2 {
		t.Fatalf("RequestCount = %d, want >= 2", got)
	}
}

func waitFor(t *testing.T, max time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(max)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func doRequest(t *testing.T, method, url string) *http.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

type stubEndpointRepo struct {
	endpoints []core.Endpoint
	err       error
}

func (s *stubEndpointRepo) List(ctx context.Context, projectID string) ([]core.Endpoint, error) {
	return s.endpoints, s.err
}

func (s *stubEndpointRepo) GetByID(ctx context.Context, id string) (*core.Endpoint, error) {
	return nil, errors.New("not implemented")
}

func (s *stubEndpointRepo) ProjectIDOf(ctx context.Context, endpointID string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *stubEndpointRepo) Replace(ctx context.Context, projectID string, endpoints []core.Endpoint) error {
	return nil
}

func (s *stubEndpointRepo) DeleteByProject(ctx context.Context, projectID string) error {
	return nil
}

func (s *stubEndpointRepo) UpdateAuthOverride(ctx context.Context, endpointID string, role core.AuthRole, tokenPath string) error {
	return nil
}

func (s *stubEndpointRepo) Stats(ctx context.Context, projectID string) (domain.ProjectStats, error) {
	return domain.ProjectStats{}, nil
}

type stubHistoryRepo struct {
	entries map[string]*domain.HistoryEntry
}

func (s *stubHistoryRepo) Save(ctx context.Context, entry domain.HistoryEntry) error {
	return nil
}

func (s *stubHistoryRepo) List(ctx context.Context, projectID string, limit int) ([]domain.HistoryEntry, error) {
	return nil, nil
}

func (s *stubHistoryRepo) GetByID(ctx context.Context, id string) (*domain.HistoryEntry, error) {
	return nil, nil
}

func (s *stubHistoryRepo) Clear(ctx context.Context, projectID string) error { return nil }

func (s *stubHistoryRepo) TrimOldest(ctx context.Context, projectID string, keep int) error {
	return nil
}

func (s *stubHistoryRepo) LatestSuccessByEndpoint(ctx context.Context, projectID, endpointID string) (*domain.HistoryEntry, error) {
	if s.entries == nil {
		return nil, nil
	}
	return s.entries[endpointID], nil
}

type stubMockRepo struct {
	overrides map[string]*domain.MockOverride
}

func (s *stubMockRepo) List(ctx context.Context, projectID string) ([]domain.MockOverride, error) {
	return nil, nil
}

func (s *stubMockRepo) Get(ctx context.Context, projectID, endpointID string) (*domain.MockOverride, error) {
	if s.overrides == nil {
		return nil, nil
	}
	return s.overrides[endpointID], nil
}

func (s *stubMockRepo) Save(ctx context.Context, override domain.MockOverride) error { return nil }

func (s *stubMockRepo) Delete(ctx context.Context, id string) error { return nil }

func (s *stubMockRepo) DeleteByProject(ctx context.Context, projectID string) error { return nil }

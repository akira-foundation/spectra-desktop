package mock

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

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

package billing

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
)

type recordingUsageRepo struct {
	mu        sync.Mutex
	rows      []domain.UsageBufferEntry
	pending   []domain.UsageBufferEntry
	pendErr   error
	markErr   error
	flushed   []string
	markCalls int32
}

func (r *recordingUsageRepo) Append(ctx context.Context, entry domain.UsageBufferEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows = append(r.rows, entry)
	return nil
}

func (r *recordingUsageRepo) PendingBatch(ctx context.Context, limit int) ([]domain.UsageBufferEntry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.pendErr != nil {
		return nil, r.pendErr
	}
	return append([]domain.UsageBufferEntry(nil), r.pending...), nil
}

func (r *recordingUsageRepo) MarkFlushed(ctx context.Context, ids []string) error {
	atomic.AddInt32(&r.markCalls, 1)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.markErr != nil {
		return r.markErr
	}
	r.flushed = append(r.flushed, ids...)
	remaining := r.pending[:0]
	for _, row := range r.pending {
		drop := false
		for _, id := range ids {
			if row.ID == id {
				drop = true
				break
			}
		}
		if !drop {
			remaining = append(remaining, row)
		}
	}
	r.pending = remaining
	return nil
}

func (r *recordingUsageRepo) snapshotFlushed() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.flushed...)
}

func newUsageClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	sdk := billingsdk.NewClient(srv.URL, ProductSlug, "test-secret")
	sdk.SetCustomerToken("customer-token")
	return &Client{sdk: sdk}, srv
}

func TestUsageTrackerTrack_AppendsToBuffer(t *testing.T) {
	repo := &recordingUsageRepo{}
	tracker := NewUsageTracker(nil, repo, nil, "fp")

	if err := tracker.Track(context.Background(), "export", 3); err != nil {
		t.Fatalf("track: %v", err)
	}
	if err := tracker.Track(context.Background(), "export", 0); err != nil {
		t.Fatalf("track zero: %v", err)
	}
	if err := tracker.Track(context.Background(), "export", -5); err != nil {
		t.Fatalf("track negative: %v", err)
	}
	if len(repo.rows) != 1 {
		t.Fatalf("expected exactly 1 row buffered, got %d", len(repo.rows))
	}
	if repo.rows[0].Feature != "export" || repo.rows[0].Amount != 3 {
		t.Fatalf("unexpected row %+v", repo.rows[0])
	}
	if repo.rows[0].ID == "" {
		t.Fatalf("expected entry to be assigned an ID")
	}
}

func TestUsageTrackerFlush_PostsDeltasAndMarksFlushed(t *testing.T) {
	var got billingsdk.LicenseSyncUsagePayload
	client, _ := newUsageClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/licenses/sync-usage" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(billingsdk.LicenseSyncUsageResponse{
			Applied: map[string]uint64{"export": 8, "requests": 2},
			Serial:  42,
		})
	})

	repo := &recordingUsageRepo{
		pending: []domain.UsageBufferEntry{
			{ID: "a", Feature: "export", Amount: 5},
			{ID: "b", Feature: "export", Amount: 3},
			{ID: "c", Feature: "requests", Amount: 2},
			{ID: "d", Feature: "ignored", Amount: 0},
		},
	}
	licenseRepo := &stubLicenseRepo{}
	tracker := NewUsageTracker(client, repo, licenseRepo, "fingerprint-x")

	if err := tracker.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if got.Product != ProductSlug || got.Fingerprint != "fingerprint-x" {
		t.Fatalf("unexpected payload identity: %+v", got)
	}
	if got.Serial != 0 {
		t.Fatalf("expected serial 0 when no stored license, got %d", got.Serial)
	}
	if got.Deltas["export"] != 8 || got.Deltas["requests"] != 2 {
		t.Fatalf("unexpected deltas %+v", got.Deltas)
	}
	if _, present := got.Deltas["ignored"]; present {
		t.Fatalf("zero-amount feature should not appear in deltas: %+v", got.Deltas)
	}
	flushed := repo.snapshotFlushed()
	if len(flushed) != 4 {
		t.Fatalf("expected 4 IDs marked flushed, got %d (%v)", len(flushed), flushed)
	}
}

func TestUsageTrackerFlush_KeepsBufferOnSyncError(t *testing.T) {
	client, _ := newUsageClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"server_down"}`, http.StatusServiceUnavailable)
	})
	repo := &recordingUsageRepo{
		pending: []domain.UsageBufferEntry{
			{ID: "a", Feature: "export", Amount: 5},
		},
	}
	tracker := NewUsageTracker(client, repo, &stubLicenseRepo{}, "fingerprint-x")

	err := tracker.Flush(context.Background())
	if err == nil {
		t.Fatalf("expected error from failing sync")
	}
	if atomic.LoadInt32(&repo.markCalls) != 0 {
		t.Fatalf("MarkFlushed must not be called on sync error")
	}
	if len(repo.pending) != 1 {
		t.Fatalf("buffer should retain rows on failure, have %d", len(repo.pending))
	}
}

func TestUsageTrackerFlush_NoopWithoutCustomerToken(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	t.Cleanup(srv.Close)
	client := &Client{sdk: billingsdk.NewClient(srv.URL, ProductSlug, "test-secret")}
	repo := &recordingUsageRepo{
		pending: []domain.UsageBufferEntry{{ID: "a", Feature: "export", Amount: 5}},
	}
	tracker := NewUsageTracker(client, repo, &stubLicenseRepo{}, "fp")

	if err := tracker.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if called {
		t.Fatalf("HTTP must not be called without customer token")
	}
}

func TestUsageTrackerFlush_NoopWhenEmpty(t *testing.T) {
	called := false
	client, _ := newUsageClient(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	repo := &recordingUsageRepo{}
	tracker := NewUsageTracker(client, repo, &stubLicenseRepo{}, "fp")

	if err := tracker.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}
	if called {
		t.Fatalf("must not call API with empty buffer")
	}
}

func TestUsageTrackerFlush_PropagatesPendingError(t *testing.T) {
	client, _ := newUsageClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("HTTP should not be hit when PendingBatch errors")
	})
	pendErr := errors.New("db down")
	repo := &recordingUsageRepo{pendErr: pendErr}
	tracker := NewUsageTracker(client, repo, &stubLicenseRepo{}, "fp")

	err := tracker.Flush(context.Background())
	if !errors.Is(err, pendErr) {
		t.Fatalf("expected wrapped pendErr, got %v", err)
	}
}

func TestUsageTrackerStartFlusher_TicksAndExitsOnCtxCancel(t *testing.T) {
	flushed := make(chan struct{}, 4)
	client, _ := newUsageClient(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case flushed <- struct{}{}:
		default:
		}
		_ = json.NewEncoder(w).Encode(billingsdk.LicenseSyncUsageResponse{
			Applied: map[string]uint64{"export": 1},
			Serial:  1,
		})
	})

	repo := &recordingUsageRepo{
		pending: []domain.UsageBufferEntry{{ID: "a", Feature: "export", Amount: 1}},
	}
	tracker := NewUsageTracker(client, repo, &stubLicenseRepo{}, "fp")

	ctx, cancel := context.WithCancel(context.Background())
	tracker.StartFlusher(ctx, 10*time.Millisecond)

	select {
	case <-flushed:
	case <-time.After(2 * time.Second):
		t.Fatalf("expected at least one flush within 2s")
	}

	cancel()
	tracker.Stop()
}

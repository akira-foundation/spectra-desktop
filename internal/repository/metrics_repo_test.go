package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/domain"
)

func TestMetricsRepository_StatusBuckets(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 204, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 301, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 404, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 500, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 0, Error: "dial fail", CreatedAt: now},
	})

	buckets, err := repo.StatusBuckets(ctx, p.ID, time.Time{})
	if err != nil {
		t.Fatalf("buckets: %v", err)
	}
	if len(buckets) != 5 {
		t.Fatalf("expected 5 buckets, got %d", len(buckets))
	}
	want := map[string]int{"2xx": 2, "3xx": 1, "4xx": 1, "5xx": 1, "err": 1}
	for _, b := range buckets {
		if want[b.Bucket] != b.Count {
			t.Fatalf("bucket %s expected %d, got %d", b.Bucket, want[b.Bucket], b.Count)
		}
	}
}

func TestMetricsRepository_StatusBuckets_SinceFilter(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 200, CreatedAt: now.Add(-48 * time.Hour)},
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 200, CreatedAt: now.Add(-1 * time.Hour)},
	})
	buckets, err := repo.StatusBuckets(ctx, p.ID, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("buckets: %v", err)
	}
	for _, b := range buckets {
		if b.Bucket == "2xx" && b.Count != 1 {
			t.Fatalf("expected since filter, got %d", b.Count)
		}
	}
}

func TestMetricsRepository_LatencyStats(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	for _, d := range []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100} {
		seedMetrics(t, hist, []domain.HistoryEntry{
			{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 200, DurationMs: d, CreatedAt: now},
		})
	}

	stats, err := repo.LatencyStats(ctx, p.ID, time.Time{})
	if err != nil {
		t.Fatalf("latency: %v", err)
	}
	if stats.Count != 10 {
		t.Fatalf("count = %d", stats.Count)
	}
	if stats.Min != 10 || stats.Max != 100 {
		t.Fatalf("min=%d max=%d", stats.Min, stats.Max)
	}
	if stats.Avg != 55 {
		t.Fatalf("avg = %d", stats.Avg)
	}
}

func TestMetricsRepository_LatencyStats_EmptyReturnsZero(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	p := seedProject(t, projects, "m")

	stats, err := repo.LatencyStats(context.Background(), p.ID, time.Time{})
	if err != nil {
		t.Fatalf("latency: %v", err)
	}
	if stats.Count != 0 {
		t.Fatalf("expected zero, got %+v", stats)
	}
}

func TestPercentile(t *testing.T) {
	t.Parallel()
	if percentile(nil, 0.5) != 0 {
		t.Fatal("nil should be 0")
	}
	if got := percentile([]int{1, 2, 3, 4, 5}, 0.5); got != 3 {
		t.Fatalf("p50 = %d", got)
	}
	if got := percentile([]int{1, 2, 3, 4, 5}, 1.0); got != 5 {
		t.Fatalf("p100 = %d", got)
	}
}

func TestMetricsRepository_DailyVolume(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/a", ResponseStatus: 200, CreatedAt: now.Add(-24 * time.Hour)},
	})

	vol, err := repo.DailyVolume(ctx, p.ID, 7)
	if err != nil {
		t.Fatalf("daily: %v", err)
	}
	if len(vol) != 7 {
		t.Fatalf("expected 7 days, got %d", len(vol))
	}
	total := 0
	for _, v := range vol {
		total += v.Count
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
}

func TestMetricsRepository_DailyVolume_DefaultDays(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	p := seedProject(t, projects, "m")

	vol, err := repo.DailyVolume(context.Background(), p.ID, 0)
	if err != nil {
		t.Fatalf("daily: %v", err)
	}
	if len(vol) != 7 {
		t.Fatalf("default days = %d", len(vol))
	}
}

func TestMetricsRepository_EndpointMetrics(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, EndpointID: "ep-1", Method: "GET", URL: "http://x/users", ResponseStatus: 200, DurationMs: 10, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "ep-1", Method: "GET", URL: "http://x/users", ResponseStatus: 500, DurationMs: 30, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "ep-2", Method: "POST", URL: "http://x/login", ResponseStatus: 200, DurationMs: 50, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "", Method: "GET", URL: "http://x/none", ResponseStatus: 200, DurationMs: 5, CreatedAt: now},
	})

	got, err := repo.EndpointMetrics(ctx, p.ID, time.Time{})
	if err != nil {
		t.Fatalf("endpoints: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(got))
	}
	byID := map[string]EndpointMetric{}
	for _, m := range got {
		byID[m.EndpointID] = m
	}
	ep1 := byID["ep-1"]
	if ep1.Count != 2 || ep1.Errors != 1 || ep1.AvgMs != 20 || ep1.Path != "/users" {
		t.Fatalf("ep-1 = %+v", ep1)
	}
	ep2 := byID["ep-2"]
	if ep2.Count != 1 || ep2.Errors != 0 || ep2.AvgMs != 50 {
		t.Fatalf("ep-2 = %+v", ep2)
	}
}

func TestShortURL(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"http://x/users", "/users"},
		{"https://api.local/v1/x", "/v1/x"},
		{"http://only", "/"},
		{"/already", "/already"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := shortURL(tc.in); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestMetricsRepository_LatencyOverTime(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, EndpointID: "fast", Method: "GET", URL: "http://x/fast", DurationMs: 10, ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "slow", Method: "GET", URL: "http://x/slow", DurationMs: 500, ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "slow", Method: "GET", URL: "http://x/slow", DurationMs: 700, ResponseStatus: 200, CreatedAt: now.Add(-24 * time.Hour)},
	})

	series, err := repo.LatencyOverTime(ctx, p.ID, 7, 5)
	if err != nil {
		t.Fatalf("latency time: %v", err)
	}
	if len(series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(series))
	}
	if series[0].EndpointID != "slow" {
		t.Fatalf("expected slow first, got %s", series[0].EndpointID)
	}
	if len(series[0].Points) != 7 {
		t.Fatalf("expected 7 points, got %d", len(series[0].Points))
	}
}

func TestMetricsRepository_LatencyOverTime_TopNCap(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	for i, dur := range []int{10, 20, 30, 40, 50} {
		seedMetrics(t, hist, []domain.HistoryEntry{
			{ProjectID: p.ID, EndpointID: uuid.NewString(), Method: "GET", URL: "http://x/" + string(rune('a'+i)), DurationMs: dur, ResponseStatus: 200, CreatedAt: now},
		})
	}
	got, err := repo.LatencyOverTime(ctx, p.ID, 7, 2)
	if err != nil {
		t.Fatalf("latency: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected topN=2, got %d", len(got))
	}
}

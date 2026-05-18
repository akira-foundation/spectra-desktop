package repository

import (
	"context"
	"testing"
	"time"

	"spectra-desktop/internal/domain"
)

func TestMetricsRepository_HourlyHeatmap(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: now},
	})
	cells, err := repo.HourlyHeatmap(ctx, p.ID, 7)
	if err != nil {
		t.Fatalf("heatmap: %v", err)
	}
	if len(cells) != 7*24 {
		t.Fatalf("expected 168 cells, got %d", len(cells))
	}
	total := 0
	for _, c := range cells {
		total += c.Count
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
}

func TestMetricsRepository_FlakyEndpoints(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	for i := 0; i < 4; i++ {
		seedMetrics(t, hist, []domain.HistoryEntry{
			{ProjectID: p.ID, EndpointID: "flaky", Method: "GET", URL: "http://x/f", ResponseStatus: 200, CreatedAt: now},
			{ProjectID: p.ID, EndpointID: "flaky", Method: "GET", URL: "http://x/f", ResponseStatus: 500, CreatedAt: now},
		})
		_ = i
	}
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, EndpointID: "stable", Method: "GET", URL: "http://x/s", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "stable", Method: "GET", URL: "http://x/s", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "stable", Method: "GET", URL: "http://x/s", ResponseStatus: 200, CreatedAt: now},
	})

	got, err := repo.FlakyEndpoints(ctx, p.ID, 3)
	if err != nil {
		t.Fatalf("flaky: %v", err)
	}
	if len(got) != 1 || got[0].EndpointID != "flaky" {
		t.Fatalf("expected only flaky, got %+v", got)
	}
	if got[0].FlakeScore <= 0 {
		t.Fatalf("expected positive flake score, got %f", got[0].FlakeScore)
	}
}

func TestAbsFloat(t *testing.T) {
	t.Parallel()
	if absFloat(-1.5) != 1.5 || absFloat(2.0) != 2.0 {
		t.Fatal("absFloat broken")
	}
}

func TestMetricsRepository_UsageOverTime(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, EndpointID: "ep-a", Method: "GET", URL: "http://x/a", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "ep-a", Method: "GET", URL: "http://x/a", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "ep-b", Method: "GET", URL: "http://x/b", ResponseStatus: 200, CreatedAt: now},
	})

	series, err := repo.UsageOverTime(ctx, p.ID, 7, 5)
	if err != nil {
		t.Fatalf("usage: %v", err)
	}
	if len(series) != 2 {
		t.Fatalf("expected 2 series, got %d", len(series))
	}
	if series[0].EndpointID != "ep-a" || series[0].Total != 2 {
		t.Fatalf("ordering or total wrong: %+v", series[0])
	}
}

func TestMetricsRepository_LastSeenByEndpoint(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	old := time.Now().UTC().Add(-48 * time.Hour)
	recent := time.Now().UTC().Add(-1 * time.Hour)
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, EndpointID: "ep-1", Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: old},
		{ProjectID: p.ID, EndpointID: "ep-1", Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: recent},
	})

	got, err := repo.LastSeenByEndpoint(ctx, p.ID)
	if err != nil {
		t.Fatalf("last seen: %v", err)
	}
	ts, ok := got["ep-1"]
	if !ok {
		t.Fatal("expected ep-1 entry")
	}
	if ts.Before(recent.Add(-time.Second)) {
		t.Fatalf("expected recent ts, got %v", ts)
	}
}

func TestMetricsRepository_Coverage(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, EndpointID: "ep-1", Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "ep-2", Method: "GET", URL: "/y", ResponseStatus: 200, CreatedAt: now},
	})

	cov, err := repo.Coverage(ctx, p.ID, 30)
	if err != nil {
		t.Fatalf("coverage: %v", err)
	}
	if cov.UsedEndpoints != 2 {
		t.Fatalf("used = %d", cov.UsedEndpoints)
	}
}

func TestMetricsRepository_MethodShare(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, Method: "GET", URL: "/x", ResponseStatus: 200, CreatedAt: now},
		{ProjectID: p.ID, Method: "POST", URL: "/x", ResponseStatus: 200, CreatedAt: now},
	})

	got, err := repo.MethodShare(ctx, p.ID, time.Time{})
	if err != nil {
		t.Fatalf("method share: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 methods, got %d", len(got))
	}
	if got[0].Method != "GET" || got[0].Count != 3 {
		t.Fatalf("top method = %+v", got[0])
	}
	if got[0].Percent != 0.75 {
		t.Fatalf("get percent = %f", got[0].Percent)
	}
}

func TestMetricsRepository_FailuresOverTime(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	hist := NewHistoryRepository(s.DB)
	repo := &MetricsRepository{db: s.DB}
	ctx := context.Background()
	p := seedProject(t, projects, "m")

	now := time.Now().UTC()
	seedMetrics(t, hist, []domain.HistoryEntry{
		{ProjectID: p.ID, EndpointID: "bad", Method: "GET", URL: "http://x/b", ResponseStatus: 500, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "bad", Method: "GET", URL: "http://x/b", ResponseStatus: 503, CreatedAt: now},
		{ProjectID: p.ID, EndpointID: "good", Method: "GET", URL: "http://x/g", ResponseStatus: 200, CreatedAt: now},
	})
	got, err := repo.FailuresOverTime(ctx, p.ID, 7, 5)
	if err != nil {
		t.Fatalf("failures: %v", err)
	}
	if len(got) != 1 || got[0].EndpointID != "bad" || got[0].Failures != 2 {
		t.Fatalf("unexpected: %+v", got)
	}
}

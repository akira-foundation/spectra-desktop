package core

import "testing"

func TestStatCardKind_Values(t *testing.T) {
	want := map[StatCardKind]string{
		StatRoutes:       "routes",
		StatControllers:  "controllers",
		StatMiddleware:   "middleware",
		StatModels:       "models",
		StatFormRequests: "form_requests",
		StatJobs:         "jobs",
		StatMailers:      "mailers",
		StatServices:     "services",
		StatErrors:       "errors",
	}
	for k, exp := range want {
		if string(k) != exp {
			t.Fatalf("kind %q != %q", string(k), exp)
		}
	}
}

func TestStatsReport_ZeroValue(t *testing.T) {
	var r StatsReport
	if r.Cards != nil {
		t.Fatalf("expected nil cards")
	}
}

func TestStatCard_Construction(t *testing.T) {
	c := StatCard{Key: "k", Kind: StatRoutes, Label: "Routes", Value: 5}
	if c.Hint != "" {
		t.Fatalf("hint should default empty")
	}
}

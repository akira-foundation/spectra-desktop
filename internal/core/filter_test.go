package core

import "testing"

func sampleEndpoints() []Endpoint {
	return []Endpoint{
		{Method: MethodGet, Path: "/api/users", Middleware: []string{"api", "auth"}},
		{Method: MethodGet, Path: "/api/posts", Middleware: []string{"App\\Http\\Middleware\\Api:guest"}},
		{Method: MethodGet, Path: "/web/home", Middleware: []string{"web"}},
		{Method: MethodGet, Path: "/api", Middleware: []string{"throttle"}},
	}
}

func TestApplyFilter_All(t *testing.T) {
	eps := sampleEndpoints()
	r := ApplyFilter(eps, FilterModeAll, "")
	if r.Mode != FilterModeAll || len(r.Endpoints) != len(eps) {
		t.Fatalf("all mode wrong: %+v", r)
	}
}

func TestApplyFilter_Middleware_Exact(t *testing.T) {
	r := ApplyFilter(sampleEndpoints(), FilterModeMiddleware, "api")
	if r.Mode != FilterModeMiddleware {
		t.Fatalf("mode: %q", r.Mode)
	}
	if len(r.Endpoints) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(r.Endpoints), r.Endpoints)
	}
}

func TestApplyFilter_Middleware_EmptyValue(t *testing.T) {
	r := ApplyFilter(sampleEndpoints(), FilterModeMiddleware, "")
	if len(r.Endpoints) != 0 {
		t.Fatalf("expected 0, got %d", len(r.Endpoints))
	}
}

func TestApplyFilter_Prefix(t *testing.T) {
	r := ApplyFilter(sampleEndpoints(), FilterModePrefix, "/api")
	if r.Mode != FilterModePrefix {
		t.Fatalf("mode: %q", r.Mode)
	}
	if len(r.Endpoints) != 3 {
		t.Fatalf("expected 3 (/api, /api/users, /api/posts), got %d", len(r.Endpoints))
	}
}

func TestApplyFilter_Prefix_WithoutLeadingSlash(t *testing.T) {
	r := ApplyFilter(sampleEndpoints(), FilterModePrefix, "api")
	if len(r.Endpoints) != 3 {
		t.Fatalf("expected 3, got %d", len(r.Endpoints))
	}
}

func TestApplyFilter_Prefix_Empty(t *testing.T) {
	r := ApplyFilter(sampleEndpoints(), FilterModePrefix, "/")
	if len(r.Endpoints) != 0 {
		t.Fatalf("expected 0, got %d", len(r.Endpoints))
	}
}

func TestApplyFilter_Auto_MiddlewareWins(t *testing.T) {
	r := ApplyFilter(sampleEndpoints(), "", "")
	if r.Mode != FilterModeMiddleware || r.Value != "api" {
		t.Fatalf("expected middleware/api auto, got %+v", r)
	}
}

func TestApplyFilter_Auto_FallbackPrefix(t *testing.T) {
	eps := []Endpoint{
		{Method: MethodGet, Path: "/api/x", Middleware: []string{"web"}},
	}
	r := ApplyFilter(eps, "auto", "")
	if r.Mode != FilterModePrefix || r.Value != "api" {
		t.Fatalf("expected prefix/api auto, got %+v", r)
	}
}

func TestApplyFilter_Auto_NoMatch(t *testing.T) {
	eps := []Endpoint{
		{Method: MethodGet, Path: "/web/x", Middleware: []string{"web"}},
	}
	r := ApplyFilter(eps, "auto", "")
	if r.Mode != FilterModeAuto || len(r.Endpoints) != 0 {
		t.Fatalf("expected empty auto, got %+v", r)
	}
}

func TestApplyFilter_TrimsAndLowercases(t *testing.T) {
	r := ApplyFilter(sampleEndpoints(), "  PREFIX  ", "  /API  ")
	if r.Mode != FilterModePrefix {
		t.Fatalf("mode: %q", r.Mode)
	}
	if len(r.Endpoints) != 3 {
		t.Fatalf("expected 3, got %d", len(r.Endpoints))
	}
}

func TestApplyFilter_ReturnsCopyOnAll(t *testing.T) {
	eps := sampleEndpoints()
	r := ApplyFilter(eps, FilterModeAll, "")
	r.Endpoints[0].Path = "mutated"
	if eps[0].Path == "mutated" {
		t.Fatalf("ApplyFilter must return a copy on all-mode")
	}
}

package laravel

import (
	"encoding/json"
	"testing"

	"spectra-desktop/internal/core"
)

func TestNormalizePath_Defaults(t *testing.T) {
	if got := normalizePath(""); got != "/" {
		t.Fatalf("want /, got %s", got)
	}
	if got := normalizePath("api/users"); got != "/api/users" {
		t.Fatalf("want /api/users, got %s", got)
	}
	if got := normalizePath("/api/users"); got != "/api/users" {
		t.Fatalf("want /api/users unchanged, got %s", got)
	}
}

func TestSplitMethods_DropsHeadWhenGetPresent(t *testing.T) {
	got := splitMethods("GET|HEAD")
	if !equalStringSlice(got, []string{"GET"}) {
		t.Fatalf("want [GET], got %v", got)
	}
}

func TestSplitMethods_KeepsHeadWhenNoGet(t *testing.T) {
	got := splitMethods("HEAD")
	if !equalStringSlice(got, []string{"HEAD"}) {
		t.Fatalf("want [HEAD], got %v", got)
	}
}

func TestSplitMethods_Multiple(t *testing.T) {
	got := splitMethods("POST|PUT")
	if !equalStringSlice(got, []string{"POST", "PUT"}) {
		t.Fatalf("want [POST PUT], got %v", got)
	}
}

func TestDecodeMiddleware_FromArray(t *testing.T) {
	raw := json.RawMessage(`["auth","throttle:api"," "]`)
	got := decodeMiddleware(raw)
	if !equalStringSlice(got, []string{"auth", "throttle:api"}) {
		t.Fatalf("got %v", got)
	}
}

func TestDecodeMiddleware_FromString(t *testing.T) {
	raw := json.RawMessage(`"auth,throttle:api"`)
	got := decodeMiddleware(raw)
	if !equalStringSlice(got, []string{"auth", "throttle:api"}) {
		t.Fatalf("got %v", got)
	}
}

func TestDecodeMiddleware_EmptyVariants(t *testing.T) {
	if got := decodeMiddleware(nil); got != nil {
		t.Fatalf("want nil for nil raw, got %v", got)
	}
	if got := decodeMiddleware(json.RawMessage(`""`)); got != nil {
		t.Fatalf("want nil for empty string, got %v", got)
	}
	if got := decodeMiddleware(json.RawMessage(`123`)); got != nil {
		t.Fatalf("want nil for unknown type, got %v", got)
	}
}

func TestNormalize_MultiMethodAndDeduplication(t *testing.T) {
	raws := []rawRoute{
		{Method: "GET|HEAD", URI: "users", Name: "users.index", Action: "App\\Http\\Controllers\\UserController@index", Middleware: json.RawMessage(`["web"]`)},
		{Method: "POST", URI: "/login", Action: "Closure", Middleware: json.RawMessage(`"guest"`)},
		{Method: "GET", URI: "users", Name: "users.index", Action: "dup"},
	}
	got := normalize(raws)
	if len(got) != 3 {
		t.Fatalf("want 3 endpoints, got %d", len(got))
	}
	if got[0].Method != core.MethodGet || got[0].Path != "/users" {
		t.Fatalf("unexpected first: %+v", got[0])
	}
	if got[0].Framework != DriverName {
		t.Fatalf("want framework laravel, got %s", got[0].Framework)
	}
	if got[1].Method != core.MethodPost || got[1].Handler != "Closure" {
		t.Fatalf("unexpected second: %+v", got[1])
	}
	if got[2].ID == got[0].ID {
		t.Fatalf("duplicate IDs not disambiguated: %s == %s", got[0].ID, got[2].ID)
	}
	if got[0].Confidence != driverConfidence {
		t.Fatalf("want confidence %v, got %v", driverConfidence, got[0].Confidence)
	}
}

func TestNormalize_EmptyInput(t *testing.T) {
	got := normalize(nil)
	if len(got) != 0 {
		t.Fatalf("want empty, got %v", got)
	}
}

func TestCleanStrings_TrimsAndDropsEmpty(t *testing.T) {
	got := cleanStrings([]string{" a ", "", "b"})
	if !equalStringSlice(got, []string{"a", "b"}) {
		t.Fatalf("got %v", got)
	}
	if cleanStrings([]string{"", " "}) != nil {
		t.Fatal("want nil for all empty")
	}
}

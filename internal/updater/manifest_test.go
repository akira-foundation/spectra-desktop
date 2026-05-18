package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

func TestPlatformKey_ContainsGOOS(t *testing.T) {
	key := PlatformKey()
	if !strings.HasPrefix(key, runtime.GOOS+"-") {
		t.Fatalf("PlatformKey() = %q, want prefix %q-", key, runtime.GOOS)
	}
}

func TestPlatformKey_ArchMapping(t *testing.T) {
	key := PlatformKey()
	switch runtime.GOARCH {
	case "arm64":
		if !strings.HasSuffix(key, "-aarch64") {
			t.Fatalf("arm64 should map to aarch64; got %q", key)
		}
	case "amd64":
		if !strings.HasSuffix(key, "-x86_64") {
			t.Fatalf("amd64 should map to x86_64; got %q", key)
		}
	default:
		t.Skipf("unmapped arch %q, skipping precise suffix check", runtime.GOARCH)
	}
}

func TestPlatformKey_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skipf("darwin-only check; running on %s", runtime.GOOS)
	}
	if !strings.HasPrefix(PlatformKey(), "darwin-") {
		t.Fatalf("expected darwin prefix, got %q", PlatformKey())
	}
}

func TestManifest_JSONRoundtrip(t *testing.T) {
	in := Manifest{
		Version: "1.2.3",
		Notes:   "hello",
		PubDate: "2026-01-01T00:00:00Z",
		Platforms: map[string]Platform{
			"darwin-aarch64": {URL: "https://example.test/x.zip", Signature: "sig"},
			"darwin-x86_64":  {URL: "https://example.test/y.zip", Signature: "sig2"},
		},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Manifest
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Version != in.Version || out.Notes != in.Notes || out.PubDate != in.PubDate {
		t.Fatalf("scalar roundtrip mismatch: %+v vs %+v", out, in)
	}
	if len(out.Platforms) != len(in.Platforms) {
		t.Fatalf("platforms len = %d, want %d", len(out.Platforms), len(in.Platforms))
	}
	for k, v := range in.Platforms {
		got, ok := out.Platforms[k]
		if !ok {
			t.Fatalf("missing platform key %q", k)
		}
		if got != v {
			t.Fatalf("platform %q = %+v, want %+v", k, got, v)
		}
	}
}

func TestFetchManifest_ParsesAndSelectsCurrentPlatform(t *testing.T) {
	key := PlatformKey()
	body := fmt.Sprintf(`{
		"version": "9.9.9",
		"notes": "release notes",
		"pub_date": "2026-05-17T00:00:00Z",
		"platforms": {
			%q: {"url": "https://example.test/app.zip", "signature": "abc"}
		}
	}`, key)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)

	prev := ManifestURL
	ManifestURL = srv.URL
	t.Cleanup(func() { ManifestURL = prev })

	m, err := fetchManifest(context.Background())
	if err != nil {
		t.Fatalf("fetchManifest: %v", err)
	}
	if m.Version != "9.9.9" {
		t.Fatalf("version = %q, want 9.9.9", m.Version)
	}
	plat, ok := m.Platforms[key]
	if !ok {
		t.Fatalf("missing platform %q", key)
	}
	if plat.URL != "https://example.test/app.zip" {
		t.Fatalf("url = %q", plat.URL)
	}
}

func TestFetchManifest_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	prev := ManifestURL
	ManifestURL = srv.URL
	t.Cleanup(func() { ManifestURL = prev })

	if _, err := fetchManifest(context.Background()); err == nil {
		t.Fatal("expected error on non-200 response")
	}
}

func TestFetchManifest_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{not-json"))
	}))
	t.Cleanup(srv.Close)

	prev := ManifestURL
	ManifestURL = srv.URL
	t.Cleanup(func() { ManifestURL = prev })

	if _, err := fetchManifest(context.Background()); err == nil {
		t.Fatal("expected parse error")
	}
}

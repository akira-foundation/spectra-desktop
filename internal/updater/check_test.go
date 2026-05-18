package updater

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"spectra-desktop/internal/version"
)

func setVersion(t *testing.T, v string) {
	t.Helper()
	prev := version.Version
	version.Version = v
	t.Cleanup(func() { version.Version = prev })
}

func startManifestServer(t *testing.T, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	prev := ManifestURL
	ManifestURL = srv.URL
	t.Cleanup(func() { ManifestURL = prev })
}

func manifestBody(version, url string) string {
	return fmt.Sprintf(`{
		"version": %q,
		"notes": "release notes here",
		"pub_date": "2026-05-17T00:00:00Z",
		"platforms": {
			%q: {"url": %q, "signature": "sig"}
		}
	}`, version, PlatformKey(), url)
}

func TestCheck_DevReturnsNil(t *testing.T) {
	setVersion(t, "dev")
	info, err := Check(context.Background(), "1.0.0")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if info != nil {
		t.Fatalf("expected nil info in dev mode, got %+v", info)
	}
}

func TestCheck_EmptyCurrentVersionReturnsNil(t *testing.T) {
	setVersion(t, "1.0.0")
	info, err := Check(context.Background(), "")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if info != nil {
		t.Fatalf("expected nil info for empty current version")
	}
}

func TestCheck_UpdateAvailable(t *testing.T) {
	setVersion(t, "1.0.0")
	startManifestServer(t, manifestBody("2.0.0", "https://example.test/app.zip"))

	info, err := Check(context.Background(), "1.0.0")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if info == nil {
		t.Fatal("expected update info, got nil")
	}
	if info.Version != "2.0.0" {
		t.Fatalf("version = %q, want 2.0.0", info.Version)
	}
	if info.CurrentVersion != "1.0.0" {
		t.Fatalf("currentVersion = %q", info.CurrentVersion)
	}
	if info.URL != "https://example.test/app.zip" {
		t.Fatalf("url = %q", info.URL)
	}
	if info.Notes != "release notes here" {
		t.Fatalf("notes = %q", info.Notes)
	}
}

func TestCheck_NoUpdateWhenEqual(t *testing.T) {
	setVersion(t, "1.0.0")
	startManifestServer(t, manifestBody("1.0.0", "https://example.test/app.zip"))

	info, err := Check(context.Background(), "1.0.0")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if info != nil {
		t.Fatalf("expected nil, got %+v", info)
	}
}

func TestCheck_NoUpdateWhenCurrentNewer(t *testing.T) {
	setVersion(t, "3.0.0")
	startManifestServer(t, manifestBody("1.0.0", "https://example.test/app.zip"))

	info, err := Check(context.Background(), "3.0.0")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if info != nil {
		t.Fatalf("expected nil when local is newer, got %+v", info)
	}
}

func TestCheck_PreReleasePrecedence(t *testing.T) {
	setVersion(t, "1.0.0-beta.1")
	startManifestServer(t, manifestBody("1.0.0", "https://example.test/app.zip"))

	info, err := Check(context.Background(), "1.0.0-beta.1")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if info == nil {
		t.Fatal("expected 1.0.0 > 1.0.0-beta.1 to trigger update")
	}
}

func TestCheck_MissingPlatformReturnsError(t *testing.T) {
	setVersion(t, "1.0.0")
	body := `{
		"version": "2.0.0",
		"notes": "",
		"pub_date": "",
		"platforms": {
			"some-unknown-platform-xyz": {"url": "https://example.test/x.zip", "signature": "s"}
		}
	}`
	startManifestServer(t, body)

	_, err := Check(context.Background(), "1.0.0")
	if err == nil {
		t.Fatal("expected error for missing platform key")
	}
}

func TestCheck_EmptyPlatformURLReturnsError(t *testing.T) {
	setVersion(t, "1.0.0")
	startManifestServer(t, manifestBody("2.0.0", ""))

	_, err := Check(context.Background(), "1.0.0")
	if err == nil {
		t.Fatal("expected error for empty platform URL")
	}
}

func TestCheck_ManifestFetchError(t *testing.T) {
	setVersion(t, "1.0.0")
	prev := ManifestURL
	ManifestURL = "http://127.0.0.1:1/nope"
	t.Cleanup(func() { ManifestURL = prev })

	if _, err := Check(context.Background(), "1.0.0"); err == nil {
		t.Fatal("expected error when manifest fetch fails")
	}
}

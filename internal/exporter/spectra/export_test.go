package spectra

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestExport_ProducesArchiveWithManifestAndData(t *testing.T) {
	srcPath, db := newMigratedDB(t)
	fx := seedProject(t, db)

	out := filepath.Join(t.TempDir(), "out.spectra")
	if err := Export(context.Background(), srcPath, fx.ProjectID, out, ExportOptions{IncludeHistory: true, IncludeSecrets: true}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	manifest, entries := openArchive(t, out)
	if manifest.FormatVersion != formatVersion {
		t.Fatalf("formatVersion=%d want %d", manifest.FormatVersion, formatVersion)
	}
	if manifest.ProjectID != fx.ProjectID || manifest.ProjectName != "demo" {
		t.Fatalf("manifest project mismatch: %+v", manifest)
	}
	if manifest.Framework != "laravel" || manifest.BaseURL != "http://demo.test" {
		t.Fatalf("manifest metadata mismatch: %+v", manifest)
	}
	if _, ok := entries[dataFile]; !ok {
		t.Fatal("archive missing data.db")
	}
	if _, ok := entries[manifestFile]; !ok {
		t.Fatal("archive missing manifest.json")
	}
	shardPath := filepath.Join(jsonShardDir, "projects.json")
	if _, ok := entries[shardPath]; !ok {
		t.Fatalf("archive missing shard %s", shardPath)
	}

	hasHistoryTable := false
	for _, table := range manifest.Tables {
		if table == "request_history" {
			hasHistoryTable = true
		}
	}
	if !hasHistoryTable {
		t.Fatal("expected request_history in tables when IncludeHistory=true")
	}
}

func TestExport_IncludeHistoryFalse_OmitsHistoryTable(t *testing.T) {
	srcPath, db := newMigratedDB(t)
	fx := seedProject(t, db)

	out := filepath.Join(t.TempDir(), "out.spectra")
	if err := Export(context.Background(), srcPath, fx.ProjectID, out, ExportOptions{IncludeHistory: false, IncludeSecrets: true}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	manifest, entries := openArchive(t, out)
	for _, table := range manifest.Tables {
		if table == "request_history" {
			t.Fatal("request_history should be absent")
		}
	}
	if _, ok := entries[filepath.Join(jsonShardDir, "request_history.json")]; ok {
		t.Fatal("history shard should not be present")
	}
}

func TestExport_IncludeSecretsFalse_StripsSecretColumns(t *testing.T) {
	srcPath, db := newMigratedDB(t)
	fx := seedProject(t, db)

	out := filepath.Join(t.TempDir(), "out.spectra")
	if err := Export(context.Background(), srcPath, fx.ProjectID, out, ExportOptions{IncludeSecrets: false}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	extractedDB := filepath.Join(t.TempDir(), "extracted.db")
	_, entries := openArchive(t, out)
	writeFile(t, extractedDB, entries[dataFile])

	raw := openRawSQLite(t, extractedDB)
	token := queryString(t, raw, "SELECT token FROM project_auth WHERE project_id = ?", fx.ProjectID)
	if token != "" {
		t.Fatalf("expected stripped token, got %q", token)
	}
}

func TestExport_IncludeSecretsTrue_KeepsSecretColumns(t *testing.T) {
	srcPath, db := newMigratedDB(t)
	fx := seedProject(t, db)

	out := filepath.Join(t.TempDir(), "out.spectra")
	if err := Export(context.Background(), srcPath, fx.ProjectID, out, ExportOptions{IncludeSecrets: true}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	extractedDB := filepath.Join(t.TempDir(), "extracted.db")
	_, entries := openArchive(t, out)
	writeFile(t, extractedDB, entries[dataFile])

	raw := openRawSQLite(t, extractedDB)
	token := queryString(t, raw, "SELECT token FROM project_auth WHERE project_id = ?", fx.ProjectID)
	if token != "super-secret-token" {
		t.Fatalf("expected token preserved, got %q", token)
	}
}

func TestExport_HistoryHeadersAreRedacted(t *testing.T) {
	srcPath, db := newMigratedDB(t)
	fx := seedProject(t, db)

	out := filepath.Join(t.TempDir(), "out.spectra")
	if err := Export(context.Background(), srcPath, fx.ProjectID, out, ExportOptions{IncludeHistory: true, IncludeSecrets: true}); err != nil {
		t.Fatalf("Export: %v", err)
	}

	extractedDB := filepath.Join(t.TempDir(), "extracted.db")
	_, entries := openArchive(t, out)
	writeFile(t, extractedDB, entries[dataFile])

	raw := openRawSQLite(t, extractedDB)
	reqHeaders := queryString(t, raw, "SELECT request_headers FROM request_history WHERE id = ?", fx.HistoryID)
	respHeaders := queryString(t, raw, "SELECT response_headers FROM request_history WHERE id = ?", fx.HistoryID)

	if !strings.Contains(reqHeaders, "[redacted]") {
		t.Fatalf("expected request Authorization redacted, got %s", reqHeaders)
	}
	if strings.Contains(reqHeaders, "Bearer abc") {
		t.Fatalf("expected raw token gone, got %s", reqHeaders)
	}
	if !strings.Contains(respHeaders, "[redacted]") {
		t.Fatalf("expected response Set-Cookie redacted, got %s", respHeaders)
	}
	if !strings.Contains(reqHeaders, "X-Trace") {
		t.Fatalf("expected non-sensitive header preserved, got %s", reqHeaders)
	}
}

func TestExport_UnknownProjectID_ReturnsError(t *testing.T) {
	srcPath, db := newMigratedDB(t)
	_ = seedProject(t, db)

	out := filepath.Join(t.TempDir(), "out.spectra")
	err := Export(context.Background(), srcPath, uuid.NewString(), out, ExportOptions{})
	if err == nil {
		t.Fatal("expected error for unknown project id")
	}
}

func TestExport_MissingSourceDB_ReturnsError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.db")
	out := filepath.Join(t.TempDir(), "out.spectra")
	err := Export(context.Background(), missing, uuid.NewString(), out, ExportOptions{})
	if err == nil {
		t.Fatal("expected error for missing source db")
	}
}

func TestExport_ConcurrentExportsSameSource_AllSucceed(t *testing.T) {
	srcPath, db := newMigratedDB(t)
	fx := seedProject(t, db)

	const n = 3
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			out := filepath.Join(t.TempDir(), "out.spectra")
			errs[i] = Export(context.Background(), srcPath, fx.ProjectID, out, ExportOptions{IncludeHistory: true, IncludeSecrets: true})
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent export %d: %v", i, err)
		}
	}
}

func TestIsSensitiveHeader_KnownNames(t *testing.T) {
	cases := map[string]bool{
		"Authorization":       true,
		"authorization":       true,
		"COOKIE":              true,
		"Set-Cookie":          true,
		"X-Api-Key":           true,
		"X-Auth-Token":        true,
		"X-OTP":               true,
		"Proxy-Authorization": true,
		"Content-Type":        false,
		"X-Trace":             false,
		"":                    false,
	}
	for name, want := range cases {
		if got := isSensitiveHeader(name); got != want {
			t.Fatalf("isSensitiveHeader(%q)=%v want %v", name, got, want)
		}
	}
}

func TestRedactSensitiveHeadersJSON_HandlesShapes(t *testing.T) {
	multi := `{"Authorization":["Bearer a"],"X-Trace":["1"]}`
	got := redactSensitiveHeadersJSON(multi)
	if !strings.Contains(got, "[redacted]") || strings.Contains(got, "Bearer a") {
		t.Fatalf("multi redact failed: %s", got)
	}

	single := `{"Authorization":"Bearer a","X-Trace":"1"}`
	got = redactSensitiveHeadersJSON(single)
	if !strings.Contains(got, "[redacted]") || strings.Contains(got, "Bearer a") {
		t.Fatalf("single redact failed: %s", got)
	}

	if redactSensitiveHeadersJSON("") != "" {
		t.Fatal("empty input should pass through")
	}
	if redactSensitiveHeadersJSON("not json") != "not json" {
		t.Fatal("non-json input should pass through")
	}
}

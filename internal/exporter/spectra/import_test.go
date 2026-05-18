package spectra

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func exportFixture(t *testing.T, opts ExportOptions) (archivePath string, fx seedFixture) {
	t.Helper()
	srcPath, db := newMigratedDB(t)
	fx = seedProject(t, db)
	archivePath = filepath.Join(t.TempDir(), "out.spectra")
	if err := Export(context.Background(), srcPath, fx.ProjectID, archivePath, opts); err != nil {
		t.Fatalf("Export: %v", err)
	}
	return archivePath, fx
}

func TestImport_RoundTrip_PreservesEndpointsWithNewProjectID(t *testing.T) {
	archive, fx := exportFixture(t, ExportOptions{IncludeHistory: true, IncludeSecrets: true})

	targetPath, _ := newMigratedDB(t)
	// Drop the seeded project so we don't collide on path UNIQUE.
	target := openRawSQLite(t, targetPath)
	if _, err := target.Exec("DELETE FROM projects WHERE id = ?", fx.ProjectID); err != nil {
		t.Fatalf("clear target: %v", err)
	}

	res, err := Import(context.Background(), archive, targetPath)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if res.NewProjectID == "" || res.NewProjectID == fx.ProjectID {
		t.Fatalf("expected new project id, got %q (old=%q)", res.NewProjectID, fx.ProjectID)
	}
	if res.ProjectName != "demo" {
		t.Fatalf("project name=%q", res.ProjectName)
	}

	count := queryInt(t, target, "SELECT count(*) FROM endpoints WHERE project_id = ?", res.NewProjectID)
	if count != 1 {
		t.Fatalf("expected 1 endpoint under new project, got %d", count)
	}
	envCount := queryInt(t, target, "SELECT count(*) FROM environments WHERE project_id = ?", res.NewProjectID)
	if envCount != 1 {
		t.Fatalf("expected 1 environment, got %d", envCount)
	}
	old := queryInt(t, target, "SELECT count(*) FROM endpoints WHERE project_id = ?", fx.ProjectID)
	if old != 0 {
		t.Fatalf("expected zero endpoints under old project id, got %d", old)
	}
}

func TestImport_RejectsNonZipArchive(t *testing.T) {
	bogus := filepath.Join(t.TempDir(), "bad.spectra")
	writeFile(t, bogus, []byte("not a zip"))

	targetPath, _ := newMigratedDB(t)
	if _, err := Import(context.Background(), bogus, targetPath); err == nil {
		t.Fatal("expected error for non-zip archive")
	}
}

func TestImport_RejectsTruncatedArchive(t *testing.T) {
	archive, _ := exportFixture(t, ExportOptions{})
	body, err := os.ReadFile(archive)
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	truncated := filepath.Join(t.TempDir(), "trunc.spectra")
	writeFile(t, truncated, body[:len(body)/2])

	targetPath, _ := newMigratedDB(t)
	if _, err := Import(context.Background(), truncated, targetPath); err == nil {
		t.Fatal("expected error for truncated archive")
	}
}

func TestImport_RejectsArchiveWithMalformedManifest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad-manifest.spectra")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	zw := zip.NewWriter(f)
	entry, _ := zw.Create(manifestFile)
	_, _ = entry.Write([]byte("{not json"))
	dataEntry, _ := zw.Create(dataFile)
	_, _ = dataEntry.Write([]byte("ignored"))
	_ = zw.Close()
	_ = f.Close()

	targetPath, _ := newMigratedDB(t)
	if _, err := Import(context.Background(), path, targetPath); err == nil {
		t.Fatal("expected manifest parse error")
	}
}

func TestImport_RejectsArchiveMissingManifest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-manifest.spectra")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	zw := zip.NewWriter(f)
	entry, _ := zw.Create(dataFile)
	_, _ = entry.Write([]byte("ignored"))
	_ = zw.Close()
	_ = f.Close()

	targetPath, _ := newMigratedDB(t)
	if _, err := Import(context.Background(), path, targetPath); err == nil {
		t.Fatal("expected missing manifest error")
	}
}

func TestImport_RejectsArchiveMissingDataDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-data.spectra")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	zw := zip.NewWriter(f)
	entry, _ := zw.Create(manifestFile)
	manifest := Manifest{FormatVersion: formatVersion, ExportedAt: time.Now().UTC(), ProjectID: "x", ProjectName: "x"}
	body, _ := json.Marshal(manifest)
	_, _ = entry.Write(body)
	_ = zw.Close()
	_ = f.Close()

	targetPath, _ := newMigratedDB(t)
	if _, err := Import(context.Background(), path, targetPath); err == nil {
		t.Fatal("expected missing data error")
	}
}

func TestImport_RejectsNewerFormatVersion(t *testing.T) {
	archive, _ := exportFixture(t, ExportOptions{})

	bumped := filepath.Join(t.TempDir(), "bumped.spectra")
	rewriteArchiveManifest(t, archive, bumped, func(m *Manifest) { m.FormatVersion = formatVersion + 1 })

	targetPath, _ := newMigratedDB(t)
	_, err := Import(context.Background(), bumped, targetPath)
	if err == nil {
		t.Fatal("expected newer-version rejection")
	}
}

func TestImport_AcceptsOlderFormatVersion(t *testing.T) {
	archive, fx := exportFixture(t, ExportOptions{})

	dropped := filepath.Join(t.TempDir(), "old.spectra")
	rewriteArchiveManifest(t, archive, dropped, func(m *Manifest) {
		if m.FormatVersion > 0 {
			m.FormatVersion--
		}
	})

	targetPath, _ := newMigratedDB(t)
	target := openRawSQLite(t, targetPath)
	if _, err := target.Exec("DELETE FROM projects WHERE id = ?", fx.ProjectID); err != nil {
		t.Fatalf("clear: %v", err)
	}

	if _, err := Import(context.Background(), dropped, targetPath); err != nil {
		t.Fatalf("Import older version: %v", err)
	}
}

func rewriteArchiveManifest(t *testing.T, src, dst string, mutate func(*Manifest)) {
	t.Helper()
	zr, err := zip.OpenReader(src)
	if err != nil {
		t.Fatalf("open src archive: %v", err)
	}
	defer zr.Close()

	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create dst: %v", err)
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open entry: %v", err)
		}
		w, err := zw.Create(f.Name)
		if err != nil {
			_ = rc.Close()
			t.Fatalf("create entry: %v", err)
		}
		if f.Name == manifestFile {
			var m Manifest
			if err := json.NewDecoder(rc).Decode(&m); err != nil {
				_ = rc.Close()
				t.Fatalf("decode manifest: %v", err)
			}
			mutate(&m)
			body, err := json.Marshal(&m)
			if err != nil {
				_ = rc.Close()
				t.Fatalf("encode manifest: %v", err)
			}
			if _, err := w.Write(body); err != nil {
				_ = rc.Close()
				t.Fatalf("write manifest: %v", err)
			}
		} else {
			if _, err := io.Copy(w, rc); err != nil {
				_ = rc.Close()
				t.Fatalf("copy entry: %v", err)
			}
		}
		_ = rc.Close()
	}
}

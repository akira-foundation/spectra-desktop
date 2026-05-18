package laravel

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"spectra-desktop/internal/core"
)

func TestCountModels_CountsPhpFiles(t *testing.T) {
	dir := t.TempDir()
	modelsDir := filepath.Join(dir, "app", "Models")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"User.php", "Post.php", "ignore.txt"} {
		if err := os.WriteFile(filepath.Join(modelsDir, name), []byte("<?php"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sub := filepath.Join(modelsDir, "Nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "Thing.php"), []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := countModels(dir); got != 3 {
		t.Fatalf("want 3, got %d", got)
	}
}

func TestCountModels_MissingDirectory(t *testing.T) {
	if got := countModels(t.TempDir()); got != 0 {
		t.Fatalf("want 0, got %d", got)
	}
}

func TestDriver_Stats_AggregatesUniqueValues(t *testing.T) {
	dir := t.TempDir()
	modelsDir := filepath.Join(dir, "app", "Models")
	if err := os.MkdirAll(modelsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modelsDir, "User.php"), []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}

	endpoints := []core.Endpoint{
		{Handler: "App\\Http\\Controllers\\UserController@index", Middleware: []string{"web", "auth"}},
		{Handler: "App\\Http\\Controllers\\UserController@store", Middleware: []string{"web"}, RequestSchema: `{"source":"form_request"}`},
		{Handler: "Closure"},
		{Handler: "App\\Http\\Controllers\\InvokableController"},
	}

	report, err := New().Stats(context.Background(), dir, endpoints)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	cards := map[string]int{}
	for _, c := range report.Cards {
		cards[c.Key] = c.Value
	}
	if cards["routes"] != 4 {
		t.Fatalf("routes: %d", cards["routes"])
	}
	if cards["controllers"] != 2 {
		t.Fatalf("controllers: %d", cards["controllers"])
	}
	if cards["middleware"] != 2 {
		t.Fatalf("middleware: %d", cards["middleware"])
	}
	if cards["form_requests"] != 1 {
		t.Fatalf("form_requests: %d", cards["form_requests"])
	}
	if cards["models"] != 1 {
		t.Fatalf("models: %d", cards["models"])
	}
}

func TestDriver_Stats_EmptyEndpoints(t *testing.T) {
	report, err := New().Stats(context.Background(), t.TempDir(), nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(report.Cards) != 5 {
		t.Fatalf("want 5 cards, got %d", len(report.Cards))
	}
}

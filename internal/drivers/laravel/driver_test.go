package laravel

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"spectra-desktop/internal/core"
)

func TestDriver_NameAndDefaults(t *testing.T) {
	d := New()
	if d.Name() != DriverName {
		t.Fatalf("name: %s", d.Name())
	}
	def := d.Defaults()
	if def.BaseURL != "http://localhost:8000" {
		t.Fatalf("baseURL: %s", def.BaseURL)
	}
	if len(def.Ports) != 1 || def.Ports[0] != 8000 {
		t.Fatalf("ports: %v", def.Ports)
	}
}

func TestDriver_Capabilities(t *testing.T) {
	caps := New().Capabilities()
	if !caps.ScanRoutes || !caps.ResolveAuth || !caps.HasFormRequests {
		t.Fatalf("missing caps: %+v", caps)
	}
	if len(caps.Stats) == 0 {
		t.Fatal("want stat kinds")
	}
}

func TestDriver_Detect_Delegates(t *testing.T) {
	dir := t.TempDir()
	for _, m := range detectionMarkers {
		full := filepath.Join(dir, m)
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		_ = os.WriteFile(full, []byte("x"), 0o644)
	}
	got := New().Detect(dir)
	if !got.Detected {
		t.Fatal("want detected")
	}
}

func TestDriver_GenerateValue(t *testing.T) {
	v := New().GenerateValue("user_id", "integer", nil)
	if _, ok := v.(int); !ok {
		t.Fatalf("want int, got %#v", v)
	}
}

func TestDriver_FormatException(t *testing.T) {
	got, ok := New().FormatException(`{"message":"oops","exception":"E"}`, 500)
	if !ok {
		t.Fatal("want ok")
	}
	if got.Message != "oops" {
		t.Fatalf("got %+v", got)
	}
}

func TestDriver_Scan_NoArtisanReturnsArtisanMissing(t *testing.T) {
	prev := currentPHPOverride()
	SetPHPBinaryOverride("/bin/sh")
	t.Cleanup(func() { SetPHPBinaryOverride(prev) })
	dir := t.TempDir()
	_, err := New().Scan(context.Background(), dir)
	if !errors.Is(err, ErrArtisanMissing) {
		t.Fatalf("want ErrArtisanMissing, got %v", err)
	}
}

func TestEnrichAuthRoles_PopulatesRoleAndHint(t *testing.T) {
	eps := []core.Endpoint{
		{Method: core.MethodPost, Path: "/api/login", Middleware: []string{"guest"}},
		{Method: core.MethodGet, Path: "/api/users"},
	}
	enrichAuthRoles(eps)
	if eps[0].AuthRole != core.AuthRoleLogin {
		t.Fatalf("first role: %s", eps[0].AuthRole)
	}
	if eps[0].AuthHint == "" {
		t.Fatal("want hint")
	}
	if eps[1].AuthRole != core.AuthRoleNone {
		t.Fatalf("second role: %s", eps[1].AuthRole)
	}
}

func TestTryInferSchema_FallsBackToInline(t *testing.T) {
	dir := t.TempDir()
	controllerDir := filepath.Join(dir, "app", "Http", "Controllers")
	if err := os.MkdirAll(controllerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `<?php
	class UserController {
		public function store(Request $request) {
			$request->validate(['email' => 'required|email']);
		}
	}`
	if err := os.WriteFile(filepath.Join(controllerDir, "UserController.php"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	schema := tryInferSchema(dir, `App\Http\Controllers\UserController@store`, "store")
	if schema == nil {
		t.Fatal("want schema")
	}
	if schema.Source != SchemaSourceInline {
		t.Fatalf("want inline, got %s", schema.Source)
	}
}

func TestTryInferSchema_NoMatch(t *testing.T) {
	if schema := tryInferSchema(t.TempDir(), `App\Foo@bar`, "bar"); schema != nil {
		t.Fatalf("want nil, got %+v", schema)
	}
}

func TestEnrichSchemas_SkipsClosureAndUnknown(t *testing.T) {
	eps := []core.Endpoint{
		{Handler: "Closure"},
		{Handler: ""},
	}
	enrichSchemas(t.TempDir(), eps)
	for _, ep := range eps {
		if ep.RequestSchema != "" {
			t.Fatalf("want empty schema, got %s", ep.RequestSchema)
		}
	}
}

func TestEnrichSchemas_PopulatesFromInline(t *testing.T) {
	dir := t.TempDir()
	controllerDir := filepath.Join(dir, "app", "Http", "Controllers")
	if err := os.MkdirAll(controllerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `<?php
	class UserController {
		public function store(Request $request) {
			$request->validate(['email' => 'required|email']);
		}
	}`
	if err := os.WriteFile(filepath.Join(controllerDir, "UserController.php"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	eps := []core.Endpoint{{Handler: `App\Http\Controllers\UserController@store`}}
	enrichSchemas(dir, eps)
	if eps[0].RequestSchema == "" {
		t.Fatal("want schema populated")
	}
	if !strings.Contains(eps[0].RequestSchema, "inline_validation") {
		t.Fatalf("expected inline source, got %s", eps[0].RequestSchema)
	}
}

func TestEnrichSchemas_InvokableHandler(t *testing.T) {
	dir := t.TempDir()
	controllerDir := filepath.Join(dir, "app", "Http", "Controllers")
	if err := os.MkdirAll(controllerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `<?php
	class InvokeController {
		public function __invoke(Request $r) {
			$r->validate(['x' => 'required|string']);
		}
	}`
	if err := os.WriteFile(filepath.Join(controllerDir, "InvokeController.php"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	eps := []core.Endpoint{{Handler: `App\Http\Controllers\InvokeController`}}
	enrichSchemas(dir, eps)
	if eps[0].RequestSchema == "" {
		t.Fatal("want schema populated for invokable")
	}
}

package laravel

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractMethodBody_Found(t *testing.T) {
	src := `class C {
		public function store(Request $r): JsonResponse
		{
			$r->validate(['email' => 'required|email']);
			return ok();
		}
	}`
	body, ok := extractMethodBody(src, "store")
	if !ok {
		t.Fatal("want ok")
	}
	if !contains(body, "validate") {
		t.Fatalf("body=%q", body)
	}
}

func TestExtractMethodBody_NotFound(t *testing.T) {
	if _, ok := extractMethodBody(`class C {}`, "store"); ok {
		t.Fatal("want false")
	}
}

func TestFindInlineRulesArray_ViaRequestValidate(t *testing.T) {
	body := `$request->validate([
		'email' => 'required|email',
		'name' => 'required|string',
	]);`
	got, ok := findInlineRulesArray(body)
	if !ok {
		t.Fatal("want ok")
	}
	if !contains(got, "email") || !contains(got, "name") {
		t.Fatalf("got %q", got)
	}
}

func TestFindInlineRulesArray_ViaValidatorMake(t *testing.T) {
	body := `Validator::make($request->all(), [
		'name' => 'required',
	])->validate();`
	got, ok := findInlineRulesArray(body)
	if !ok {
		t.Fatal("want ok")
	}
	if !contains(got, "name") {
		t.Fatalf("got %q", got)
	}
}

func TestFindInlineRulesArray_None(t *testing.T) {
	if _, ok := findInlineRulesArray(`return [];`); ok {
		t.Fatal("want false")
	}
}

func TestInferFromInlineValidation_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	controllerDir := filepath.Join(dir, "app", "Http", "Controllers")
	if err := os.MkdirAll(controllerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `<?php
	namespace App\Http\Controllers;
	class UserController {
		public function store(Request $request) {
			$request->validate([
				'email' => 'required|email',
				'age' => 'required|integer',
			]);
			return response()->json(['ok' => true]);
		}
	}`
	if err := os.WriteFile(filepath.Join(controllerDir, "UserController.php"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	schema, err := inferFromInlineValidation(dir, `App\Http\Controllers\UserController@store`, "store")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if schema.Source != SchemaSourceInline || schema.Confidence != ConfidenceMedium {
		t.Fatalf("source/confidence wrong: %+v", schema)
	}
	if len(schema.Fields) != 2 {
		t.Fatalf("want 2 fields, got %d", len(schema.Fields))
	}
}

func TestInferFromInlineValidation_NoValidation(t *testing.T) {
	dir := t.TempDir()
	controllerDir := filepath.Join(dir, "app", "Http", "Controllers")
	if err := os.MkdirAll(controllerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `<?php class UserController { public function store() { return; } }`
	if err := os.WriteFile(filepath.Join(controllerDir, "UserController.php"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := inferFromInlineValidation(dir, `App\Http\Controllers\UserController@store`, "store"); !errors.Is(err, ErrNoInlineValidation) {
		t.Fatalf("got %v", err)
	}
}

func TestInferFromInlineValidation_NoController(t *testing.T) {
	if _, err := inferFromInlineValidation(t.TempDir(), `App\X@a`, "a"); !errors.Is(err, ErrControllerNotFound) {
		t.Fatalf("got %v", err)
	}
}

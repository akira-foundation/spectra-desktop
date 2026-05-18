package laravel

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindMethodSignature_Found(t *testing.T) {
	src := `class C {
		public function store(StoreUserRequest $req): JsonResponse {
			return $this->ok();
		}
	}`
	sig, ok := findMethodSignature(src, "store")
	if !ok {
		t.Fatal("want signature")
	}
	if sig != "StoreUserRequest $req" {
		t.Fatalf("got %q", sig)
	}
}

func TestFindMethodSignature_NotFound(t *testing.T) {
	if _, ok := findMethodSignature(`class C {}`, "store"); ok {
		t.Fatal("want false")
	}
}

func TestFindFormRequestParam_Matches(t *testing.T) {
	if got, ok := findFormRequestParam(`StoreUserRequest $req`); !ok || got != "StoreUserRequest" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
	if got, ok := findFormRequestParam(`\App\Http\Requests\UpdateUserRequest $req`); !ok || got != "UpdateUserRequest" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}

func TestFindFormRequestParam_NoMatch(t *testing.T) {
	if _, ok := findFormRequestParam(`int $id`); ok {
		t.Fatal("want false")
	}
}

func TestExtractRulesArray_BasicReturn(t *testing.T) {
	src := `public function rules(): array
	{
		return [
			'email' => 'required|email',
			'name' => 'required|string|max:255',
		];
	}`
	body, ok := extractRulesArray(src)
	if !ok {
		t.Fatal("want ok")
	}
	if body == "" || !contains(body, "email") {
		t.Fatalf("body=%q", body)
	}
}

func TestExtractRulesArray_NoRulesMethod(t *testing.T) {
	if _, ok := extractRulesArray(`class X { public function other() { return []; } }`); ok {
		t.Fatal("want false")
	}
}

func TestExtractRulesArray_NoReturnArray(t *testing.T) {
	src := `public function rules(): array { $a = 1; }`
	if _, ok := extractRulesArray(src); ok {
		t.Fatal("want false")
	}
}

func TestClassToProjectPath_AppPrefixStripped(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "app", "Http", "Controllers", "UserController.php")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := classToProjectPath(dir, `App\Http\Controllers\UserController`)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != target {
		t.Fatalf("want %s, got %s", target, got)
	}
}

func TestClassToProjectPath_MissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := classToProjectPath(dir, `App\Foo\Bar`); !errors.Is(err, ErrControllerNotFound) {
		t.Fatalf("want ErrControllerNotFound, got %v", err)
	}
}

func TestClassToProjectPath_TooFewParts(t *testing.T) {
	if _, err := classToProjectPath(t.TempDir(), `OnlyOne`); !errors.Is(err, ErrControllerNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveControllerFile_HandlerSplit(t *testing.T) {
	dir := makeControllerFixture(t)
	got, err := resolveControllerFile(dir, `App\Http\Controllers\UserController@store`)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !contains(got, "UserController.php") {
		t.Fatalf("got %s", got)
	}
}

func TestResolveFormRequestFile_FromImport(t *testing.T) {
	dir := t.TempDir()
	requests := filepath.Join(dir, "app", "Http", "Requests")
	if err := os.MkdirAll(requests, 0o755); err != nil {
		t.Fatal(err)
	}
	requestFile := filepath.Join(requests, "StoreUserRequest.php")
	if err := os.WriteFile(requestFile, []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	controllerSrc := `<?php
	use App\Http\Requests\StoreUserRequest;
	class UserController {}`
	got, err := resolveFormRequestFile(dir, controllerSrc, "StoreUserRequest")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != requestFile {
		t.Fatalf("got %s", got)
	}
}

func TestResolveFormRequestFile_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	requests := filepath.Join(dir, "app", "Http", "Requests")
	if err := os.MkdirAll(requests, 0o755); err != nil {
		t.Fatal(err)
	}
	requestFile := filepath.Join(requests, "MyRequest.php")
	if err := os.WriteFile(requestFile, []byte("<?php"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveFormRequestFile(dir, "<?php class C {}", "MyRequest")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != requestFile {
		t.Fatalf("got %s", got)
	}
}

func TestResolveFormRequestFile_NotFound(t *testing.T) {
	if _, err := resolveFormRequestFile(t.TempDir(), "<?php", "Missing"); !errors.Is(err, ErrControllerNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestInferFromFormRequest_EndToEnd(t *testing.T) {
	dir := makeFullFormRequestFixture(t)
	schema, err := inferFromFormRequest(dir, `App\Http\Controllers\UserController@store`, "store")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if schema.Source != SchemaSourceFormRequest {
		t.Fatalf("got source %s", schema.Source)
	}
	names := map[string]bool{}
	for _, f := range schema.Fields {
		names[f.Name] = true
	}
	for _, want := range []string{"email", "password", "password_confirmation"} {
		if !names[want] {
			t.Errorf("missing field %s in %+v", want, names)
		}
	}
}

func TestInferFromFormRequest_PlainRequestRejected(t *testing.T) {
	dir := t.TempDir()
	controller := filepath.Join(dir, "app", "Http", "Controllers", "UserController.php")
	if err := os.MkdirAll(filepath.Dir(controller), 0o755); err != nil {
		t.Fatal(err)
	}
	src := `<?php
	class UserController {
		public function store(Request $req) { return; }
	}`
	if err := os.WriteFile(controller, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := inferFromFormRequest(dir, `App\Http\Controllers\UserController@store`, "store"); !errors.Is(err, ErrNoFormRequest) {
		t.Fatalf("want ErrNoFormRequest, got %v", err)
	}
}

func TestInferFromFormRequest_NoControllerFile(t *testing.T) {
	if _, err := inferFromFormRequest(t.TempDir(), `App\Foo@bar`, "bar"); !errors.Is(err, ErrControllerNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(p, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readFile(p)
	if err != nil || got != "hi" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := readFile(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("want err")
	}
}

func makeControllerFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "app", "Http", "Controllers", "UserController.php")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("<?php class UserController {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func makeFullFormRequestFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	controllerDir := filepath.Join(dir, "app", "Http", "Controllers")
	requestsDir := filepath.Join(dir, "app", "Http", "Requests")
	if err := os.MkdirAll(controllerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(requestsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	controllerSrc := `<?php
	namespace App\Http\Controllers;
	use App\Http\Requests\StoreUserRequest;
	class UserController {
		public function store(StoreUserRequest $request) {
			return response()->json(['ok' => true]);
		}
	}`
	requestSrc := `<?php
	namespace App\Http\Requests;
	use Illuminate\Foundation\Http\FormRequest;
	class StoreUserRequest extends FormRequest {
		public function rules(): array
		{
			return [
				'email' => 'required|email',
				'password' => 'required|string|min:8|confirmed',
			];
		}
	}`
	if err := os.WriteFile(filepath.Join(controllerDir, "UserController.php"), []byte(controllerSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(requestsDir, "StoreUserRequest.php"), []byte(requestSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

package laravel

import "testing"

func TestParseLaravelException_BelowErrorStatus(t *testing.T) {
	if _, ok := parseLaravelException(`{"message":"hi"}`, 200); ok {
		t.Fatal("non-error status must return false")
	}
}

func TestParseLaravelException_NonJSON(t *testing.T) {
	if _, ok := parseLaravelException("<html>nope</html>", 500); ok {
		t.Fatal("non-json must return false")
	}
}

func TestParseLaravelException_InvalidJSON(t *testing.T) {
	if _, ok := parseLaravelException(`{not-json`, 500); ok {
		t.Fatal("invalid json must return false")
	}
}

func TestParseLaravelException_EmptyShape(t *testing.T) {
	if _, ok := parseLaravelException(`{"trace":[]}`, 500); ok {
		t.Fatal("empty shape must return false")
	}
}

func TestParseLaravelException_FullPayload(t *testing.T) {
	body := `{
		"message":"Server Error",
		"exception":"RuntimeException",
		"file":"/app/Foo.php",
		"line":42,
		"trace":[
			{"file":"/a.php","line":1,"function":"do","class":"X","type":"->"},
			{"file":"/b.php","line":2,"function":"go"}
		]
	}`
	got, ok := parseLaravelException(body, 500)
	if !ok {
		t.Fatal("want ok")
	}
	if got.Message != "Server Error" || got.Class != "RuntimeException" || got.Line != 42 {
		t.Fatalf("unexpected: %+v", got)
	}
	if len(got.Trace) != 2 {
		t.Fatalf("trace len: %d", len(got.Trace))
	}
	if got.Trace[0].Function != "X->do" {
		t.Fatalf("want X->do, got %s", got.Trace[0].Function)
	}
	if got.Trace[1].Function != "go" {
		t.Fatalf("want go, got %s", got.Trace[1].Function)
	}
}

func TestParseLaravelException_ValidationErrors(t *testing.T) {
	body := `{"message":"The given data was invalid.","errors":{"email":["bad"]}}`
	got, ok := parseLaravelException(body, 422)
	if !ok {
		t.Fatal("want ok")
	}
	if got.Extra["errors"] == nil {
		t.Fatalf("want errors extra, got %+v", got.Extra)
	}
}

func TestParseLaravelException_TraceCappedAt25(t *testing.T) {
	frames := ""
	for i := 0; i < 40; i++ {
		if i > 0 {
			frames += ","
		}
		frames += `{"file":"f","line":1,"function":"fn"}`
	}
	body := `{"message":"m","trace":[` + frames + `]}`
	got, ok := parseLaravelException(body, 500)
	if !ok {
		t.Fatal("want ok")
	}
	if len(got.Trace) != 25 {
		t.Fatalf("want 25 frames cap, got %d", len(got.Trace))
	}
}

func TestBuildFnName_NoClass(t *testing.T) {
	if buildFnName(traceFrame{Function: "fn"}) != "fn" {
		t.Fatal("want fn")
	}
}

func TestBuildFnName_DefaultSeparator(t *testing.T) {
	if buildFnName(traceFrame{Class: "C", Function: "fn"}) != "C::fn" {
		t.Fatal("want C::fn")
	}
}

func TestBuildFnName_ExplicitSeparator(t *testing.T) {
	if buildFnName(traceFrame{Class: "C", Function: "fn", Type: "->"}) != "C->fn" {
		t.Fatal("want C->fn")
	}
}

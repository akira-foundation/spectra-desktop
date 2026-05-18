package app

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestExtractCaptureValue_HeaderSource(t *testing.T) {
	h := http.Header{}
	h.Set("X-Token", "abc")
	v, ok := extractCaptureValue("header", "X-Token", nil, h)
	if !ok || v != "abc" {
		t.Fatalf("got %q ok=%v", v, ok)
	}
}

func TestExtractCaptureValue_MissingHeader(t *testing.T) {
	if _, ok := extractCaptureValue("header", "X-Missing", nil, http.Header{}); ok {
		t.Fatal("expected not ok")
	}
}

func TestExtractCaptureValue_BodyJSONPath(t *testing.T) {
	body := map[string]any{"data": map[string]any{"token": "xyz"}}
	v, ok := extractCaptureValue("body", "$.data.token", body, nil)
	if !ok || v != "xyz" {
		t.Fatalf("got %q ok=%v", v, ok)
	}
}

func TestExtractCaptureValue_DefaultSourceIsBody(t *testing.T) {
	body := map[string]any{"k": "v"}
	v, ok := extractCaptureValue("", "k", body, nil)
	if !ok || v != "v" {
		t.Fatalf("got %q ok=%v", v, ok)
	}
}

func TestExtractCaptureValue_UnknownSource(t *testing.T) {
	if _, ok := extractCaptureValue("cookie", "x", nil, nil); ok {
		t.Fatal("expected unknown source to fail")
	}
}

func TestFormatCapturedValue_StringNilOther(t *testing.T) {
	if formatCapturedValue("a") != "a" {
		t.Fatal("string")
	}
	if formatCapturedValue(nil) != "" {
		t.Fatal("nil")
	}
	got := formatCapturedValue(map[string]any{"k": 1})
	var back map[string]int
	if err := json.Unmarshal([]byte(got), &back); err != nil || back["k"] != 1 {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestLookupJSONPath_DotNotation(t *testing.T) {
	root := map[string]any{"a": map[string]any{"b": "c"}}
	v, ok := lookupJSONPath(root, "$.a.b")
	if !ok || v != "c" {
		t.Fatalf("got %v ok=%v", v, ok)
	}
}

func TestLookupJSONPath_ArrayIndex(t *testing.T) {
	root := map[string]any{"xs": []any{"a", "b", "c"}}
	v, ok := lookupJSONPath(root, "xs[1]")
	if !ok || v != "b" {
		t.Fatalf("got %v ok=%v", v, ok)
	}
}

func TestLookupJSONPath_RootReturnsRoot(t *testing.T) {
	root := map[string]any{"a": 1}
	v, ok := lookupJSONPath(root, "$")
	if !ok || v == nil {
		t.Fatalf("got %v ok=%v", v, ok)
	}
}

func TestLookupJSONPath_MissingKey(t *testing.T) {
	if _, ok := lookupJSONPath(map[string]any{"a": 1}, "b"); ok {
		t.Fatal("expected miss")
	}
}

func TestLookupJSONPath_BadArrayIndex(t *testing.T) {
	root := map[string]any{"xs": []any{"a"}}
	if _, ok := lookupJSONPath(root, "xs[5]"); ok {
		t.Fatal("expected oob")
	}
	if _, ok := lookupJSONPath(root, "xs[x]"); ok {
		t.Fatal("expected non-int fail")
	}
}

func TestSplitJSONPath_EmptyReturnsNil(t *testing.T) {
	if got := splitJSONPath(""); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestSplitJSONPath_DotAndBracket(t *testing.T) {
	got := splitJSONPath("a.b[0].c")
	want := []string{"a", "b", "0", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("idx %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestSplitJSONPath_StripsQuotes(t *testing.T) {
	got := splitJSONPath(`["weird key"]`)
	if len(got) != 1 || got[0] != "weird key" {
		t.Fatalf("got %v", got)
	}
}

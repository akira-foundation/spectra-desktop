package laravel

import (
	"strings"
	"testing"
)

func TestExtractArrayLiteral_Basic(t *testing.T) {
	src := `[a, b, c]`
	body, end, ok := extractArrayLiteral(src, 1)
	if !ok {
		t.Fatal("want ok")
	}
	if body != "a, b, c" {
		t.Fatalf("body=%q", body)
	}
	if src[end] != ']' {
		t.Fatalf("end byte not ]: %q", src[end])
	}
}

func TestExtractArrayLiteral_Nested(t *testing.T) {
	src := `[[1,2],[3]]`
	body, _, ok := extractArrayLiteral(src, 1)
	if !ok {
		t.Fatal("want ok")
	}
	if body != "[1,2],[3]" {
		t.Fatalf("body=%q", body)
	}
}

func TestExtractArrayLiteral_WithStringsContainingBrackets(t *testing.T) {
	src := `['a]b', "c[d]"]`
	body, _, ok := extractArrayLiteral(src, 1)
	if !ok {
		t.Fatalf("want ok, got %q", body)
	}
	if !strings.Contains(body, "a]b") {
		t.Fatalf("body lost string content: %q", body)
	}
}

func TestExtractArrayLiteral_BadStartIndex(t *testing.T) {
	src := `xyz`
	if _, _, ok := extractArrayLiteral(src, 1); ok {
		t.Fatal("want false when no [ before startsAt")
	}
	if _, _, ok := extractArrayLiteral(src, 0); ok {
		t.Fatal("want false for startsAt=0")
	}
}

func TestExtractArrayLiteral_Unterminated(t *testing.T) {
	src := `[1,2`
	if _, _, ok := extractArrayLiteral(src, 1); ok {
		t.Fatal("want false")
	}
}

func TestExtractBraceBlock_Basic(t *testing.T) {
	src := `{ return 1; }`
	body, _, ok := extractBraceBlock(src, 1)
	if !ok {
		t.Fatal("want ok")
	}
	if !strings.Contains(body, "return 1;") {
		t.Fatalf("body=%q", body)
	}
}

func TestExtractBraceBlock_Nested(t *testing.T) {
	src := `{ if (x) { y(); } }`
	body, _, ok := extractBraceBlock(src, 1)
	if !ok {
		t.Fatal("want ok")
	}
	if !strings.Contains(body, "{ y(); }") {
		t.Fatalf("body=%q", body)
	}
}

func TestExtractBraceBlock_BadStart(t *testing.T) {
	if _, _, ok := extractBraceBlock("nope", 1); ok {
		t.Fatal("want false")
	}
}

func TestParseRulePairs_StringRule(t *testing.T) {
	body := `'email' => 'required|email', "name" => 'string|max:255'`
	got := parseRulePairs(body)
	if len(got) != 2 {
		t.Fatalf("want 2 fields, got %d (%+v)", len(got), got)
	}
	if got[0].Name != "email" || got[0].Type != "email" || !got[0].Required {
		t.Fatalf("email field wrong: %+v", got[0])
	}
	if got[1].Name != "name" || got[1].Type != "string" {
		t.Fatalf("name field wrong: %+v", got[1])
	}
}

func TestParseRulePairs_ArrayRule(t *testing.T) {
	body := `'age' => ['required', 'integer']`
	got := parseRulePairs(body)
	if len(got) != 1 || got[0].Type != "integer" || !got[0].Required {
		t.Fatalf("got %+v", got)
	}
}

func TestExpandConfirmedFields_AddsConfirmation(t *testing.T) {
	in := []InferredField{
		{Name: "password", Type: "string", Required: true, Rules: []string{"required", "confirmed"}, Example: "password"},
	}
	out := expandConfirmedFields(in)
	if len(out) != 2 {
		t.Fatalf("want 2 fields, got %d", len(out))
	}
	if out[1].Name != "password_confirmation" {
		t.Fatalf("want password_confirmation, got %s", out[1].Name)
	}
}

func TestExpandConfirmedFields_SkipsWhenConfirmationExists(t *testing.T) {
	in := []InferredField{
		{Name: "password", Rules: []string{"confirmed"}},
		{Name: "password_confirmation"},
	}
	out := expandConfirmedFields(in)
	if len(out) != 2 {
		t.Fatalf("want 2, got %d", len(out))
	}
}

func TestExpandConfirmedFields_NoConfirmedRule(t *testing.T) {
	in := []InferredField{{Name: "a"}}
	out := expandConfirmedFields(in)
	if len(out) != 1 {
		t.Fatalf("want 1, got %d", len(out))
	}
}

func TestSplitArrayStrings(t *testing.T) {
	got := splitArrayStrings(`'required', "email", 'min:3'`)
	want := []string{"required", "email", "min:3"}
	if !equalStringSlice(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestDecodeRulesValue_StringForm(t *testing.T) {
	got := decodeRulesValue(`'required|email'`)
	if !equalStringSlice(got, []string{"required", "email"}) {
		t.Fatalf("got %v", got)
	}
}

func TestDecodeRulesValue_ArrayForm(t *testing.T) {
	got := decodeRulesValue(`['required','email']`)
	if !equalStringSlice(got, []string{"required", "email"}) {
		t.Fatalf("got %v", got)
	}
}

func TestDecodeRulesValue_EmptyAndUnknown(t *testing.T) {
	if decodeRulesValue("") != nil {
		t.Fatal("want nil")
	}
	if decodeRulesValue("123") != nil {
		t.Fatal("want nil for unknown form")
	}
}

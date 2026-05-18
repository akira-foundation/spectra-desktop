package app

import "testing"

func TestSubstituteVars_ReplacesKnownVars(t *testing.T) {
	out := substituteVars("hello {{name}}!", map[string]string{"name": "world"})
	if out != "hello world!" {
		t.Fatalf("got %q", out)
	}
}

func TestSubstituteVars_LeavesUnknownVarsIntact(t *testing.T) {
	out := substituteVars("a={{missing}} b={{known}}", map[string]string{"known": "K"})
	if out != "a={{missing}} b=K" {
		t.Fatalf("got %q", out)
	}
}

func TestSubstituteVars_HandlesWhitespaceInsideBraces(t *testing.T) {
	out := substituteVars("{{  name  }}", map[string]string{"name": "v"})
	if out != "v" {
		t.Fatalf("got %q", out)
	}
}

func TestSubstituteVars_EmptyInput(t *testing.T) {
	if got := substituteVars("", map[string]string{"k": "v"}); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteVars_NoVarsMap(t *testing.T) {
	if got := substituteVars("hi {{x}}", nil); got != "hi {{x}}" {
		t.Fatalf("got %q", got)
	}
}

func TestSubstituteVars_DotAndDashKeys(t *testing.T) {
	out := substituteVars("{{a.b}}-{{x-y}}", map[string]string{"a.b": "1", "x-y": "2"})
	if out != "1-2" {
		t.Fatalf("got %q", out)
	}
}

func TestSubstituteHeaderVars_SubstitutesKeyAndValue(t *testing.T) {
	in := map[string]string{"X-{{h}}": "Bearer {{t}}"}
	out := substituteHeaderVars(in, map[string]string{"h": "Auth", "t": "tok"})
	if out["X-Auth"] != "Bearer tok" {
		t.Fatalf("got %v", out)
	}
}

func TestSubstituteHeaderVars_DropsInvalidHeaderName(t *testing.T) {
	in := map[string]string{"Bad Header": "v", "Good": "ok"}
	out := substituteHeaderVars(in, nil)
	if _, ok := out["Bad Header"]; ok {
		t.Fatalf("expected invalid header dropped")
	}
	if out["Good"] != "ok" {
		t.Fatalf("good header missing: %v", out)
	}
}

func TestSubstituteHeaderVars_EmptyInput(t *testing.T) {
	out := substituteHeaderVars(nil, map[string]string{"x": "y"})
	if out != nil {
		t.Fatalf("expected nil, got %v", out)
	}
}

func TestIsValidHeaderName_AcceptsToken(t *testing.T) {
	cases := []string{"X-Foo", "Content-Type", "a", "X_1.2!"}
	for _, c := range cases {
		if !isValidHeaderName(c) {
			t.Fatalf("expected %q valid", c)
		}
	}
}

func TestIsValidHeaderName_RejectsInvalid(t *testing.T) {
	cases := []string{"", "a b", "a:b", "a/b", "(x)"}
	for _, c := range cases {
		if isValidHeaderName(c) {
			t.Fatalf("expected %q invalid", c)
		}
	}
}

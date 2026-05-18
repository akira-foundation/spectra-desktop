package laravel

import (
	"strings"
	"testing"
)

func TestRegenerateValue_EmailType(t *testing.T) {
	v := RegenerateValue("contact", "email", nil)
	s, ok := v.(string)
	if !ok || !strings.Contains(s, "@") {
		t.Fatalf("want email-like string, got %#v", v)
	}
}

func TestRegenerateValue_IntegerWithIdHint(t *testing.T) {
	v := RegenerateValue("user_id", "integer", nil)
	n, ok := v.(int)
	if !ok || n < 1 || n > 9999 {
		t.Fatalf("want 1..9999 int, got %#v", v)
	}
}

func TestRegenerateValue_ArrayType(t *testing.T) {
	v := RegenerateValue("items", "array", nil)
	if arr, ok := v.([]any); !ok || len(arr) != 0 {
		t.Fatalf("want empty []any, got %#v", v)
	}
}

func TestRegenerateValue_ObjectType(t *testing.T) {
	v := RegenerateValue("meta", "object", nil)
	if _, ok := v.(map[string]any); !ok {
		t.Fatalf("want map, got %#v", v)
	}
}

func TestRegenerateValue_FileTypeIsNil(t *testing.T) {
	if RegenerateValue("upload", "file", nil) != nil {
		t.Fatal("want nil for file")
	}
}

func TestRegenerateValue_BooleanType(t *testing.T) {
	v := RegenerateValue("flag", "boolean", nil)
	if _, ok := v.(bool); !ok {
		t.Fatalf("want bool, got %#v", v)
	}
}

func TestRegenerateValue_PasswordIsLiteral(t *testing.T) {
	if RegenerateValue("password", "string", nil) != "password" {
		t.Fatal("want literal password")
	}
}

func TestRegenerateValue_NameHints(t *testing.T) {
	cases := []string{"first_name", "last_name", "company_name", "user_name"}
	for _, n := range cases {
		v := RegenerateValue(n, "string", nil)
		s, ok := v.(string)
		if !ok || s == "" {
			t.Errorf("%s: want non-empty string, got %#v", n, v)
		}
	}
}

func TestRegenerateValue_EmailFromName(t *testing.T) {
	v := RegenerateValue("contact_email", "string", nil)
	if s, ok := v.(string); !ok || !strings.Contains(s, "@") {
		t.Fatalf("got %#v", v)
	}
}

func TestRegenerateValue_URLFallback(t *testing.T) {
	v := RegenerateValue("website", "string", nil)
	if s, ok := v.(string); !ok || !strings.HasPrefix(s, "http") {
		t.Fatalf("got %#v", v)
	}
}

func TestRegenerateValue_ImageReturnsPicsum(t *testing.T) {
	v := RegenerateValue("avatar", "string", nil)
	if v != "https://picsum.photos/640/480" {
		t.Fatalf("got %v", v)
	}
}

func TestBuildExampleBody(t *testing.T) {
	fields := []InferredField{
		{Name: "a", Example: 1},
		{Name: "b", Example: "x"},
	}
	body := buildExampleBody(fields)
	if body["a"] != 1 || body["b"] != "x" {
		t.Fatalf("got %v", body)
	}
}

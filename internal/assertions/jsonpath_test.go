package assertions

import (
	"strings"
	"testing"
)

func bodyJSON() any {
	return map[string]any{
		"user": map[string]any{
			"id":    float64(42),
			"name":  "Ada",
			"admin": true,
			"score": 3.14,
			"tags":  []any{"go", "rust"},
		},
		"items": []any{
			map[string]any{"sku": "A1"},
			map[string]any{"sku": "B2"},
		},
		"empty": nil,
	}
}

func TestLookupPath_Nested(t *testing.T) {
	v, ok := lookupPath(bodyJSON(), "$.user.name")
	if !ok || v != "Ada" {
		t.Fatalf("got %v ok=%v", v, ok)
	}
}

func TestLookupPath_ArrayIndex(t *testing.T) {
	v, ok := lookupPath(bodyJSON(), "items[1].sku")
	if !ok || v != "B2" {
		t.Fatalf("got %v ok=%v", v, ok)
	}
}

func TestLookupPath_BracketedKey(t *testing.T) {
	v, ok := lookupPath(bodyJSON(), `user["name"]`)
	if !ok || v != "Ada" {
		t.Fatalf("got %v ok=%v", v, ok)
	}
}

func TestLookupPath_Missing(t *testing.T) {
	_, ok := lookupPath(bodyJSON(), "user.missing")
	if ok {
		t.Fatalf("expected miss")
	}
}

func TestLookupPath_IndexOutOfRange(t *testing.T) {
	_, ok := lookupPath(bodyJSON(), "items.5")
	if ok {
		t.Fatalf("expected miss")
	}
}

func TestLookupPath_DescendIntoScalar(t *testing.T) {
	_, ok := lookupPath(bodyJSON(), "user.name.extra")
	if ok {
		t.Fatalf("expected miss")
	}
}

func TestLookupPath_RootDollar(t *testing.T) {
	v, ok := lookupPath(bodyJSON(), "$")
	if !ok || v == nil {
		t.Fatalf("expected root, got %v ok=%v", v, ok)
	}
}

func TestCheckJSONPath_Exists(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name"})
	if !pass {
		t.Fatalf("expected exists pass")
	}
	failPass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.missing"})
	if failPass || !strings.Contains(msg, "path not found") {
		t.Fatalf("expected path not found, got %q", msg)
	}
}

func TestCheckJSONPath_NotExists(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.missing", Op: "not_exists"})
	if !pass {
		t.Fatalf("expected not_exists pass")
	}
	failPass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name", Op: "not_exists"})
	if failPass {
		t.Fatalf("expected not_exists fail")
	}
}

func TestCheckJSONPath_NilBodyNotExists(t *testing.T) {
	pass, _ := checkJSONPath(nil, Test{JSONPath: "anything", Op: "not_exists"})
	if !pass {
		t.Fatalf("nil body should pass not_exists")
	}
}

func TestCheckJSONPath_NilBodyOther(t *testing.T) {
	pass, msg := checkJSONPath(nil, Test{JSONPath: "x", Op: "exists"})
	if pass || !strings.Contains(msg, "not JSON") {
		t.Fatalf("expected not JSON, got %q", msg)
	}
}

func TestCheckJSONPath_EqualsString(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name", Op: "equals", Expected: "Ada"})
	if !pass {
		t.Fatalf("expected pass")
	}
	failPass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name", Op: "equals", Expected: "Bob"})
	if failPass {
		t.Fatalf("expected fail")
	}
	if !strings.Contains(msg, "Ada") || !strings.Contains(msg, "Bob") {
		t.Fatalf("expected path/expected/actual in msg: %q", msg)
	}
}

func TestCheckJSONPath_EqualsNumber(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.id", Op: "equals", Expected: "42"})
	if !pass {
		t.Fatalf("number equals failed")
	}
	failPass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.id", Op: "equals", Expected: "99"})
	if failPass {
		t.Fatalf("expected fail")
	}
}

func TestCheckJSONPath_EqualsBool(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.admin", Op: "equals", Expected: "true"})
	if !pass {
		t.Fatalf("bool equals failed")
	}
}

func TestCheckJSONPath_EqualsNull(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "empty", Op: "equals", Expected: "null"})
	if !pass {
		t.Fatalf("null equals failed")
	}
}

func TestCheckJSONPath_EqualsArray(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.tags", Op: "equals", Expected: `["go","rust"]`})
	if !pass {
		t.Fatalf("array equals failed")
	}
}

func TestCheckJSONPath_EqualsMissingPath(t *testing.T) {
	pass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "nope", Op: "equals", Expected: "x"})
	if pass || !strings.Contains(msg, "path not found") {
		t.Fatalf("expected path not found, got %q", msg)
	}
}

func TestCheckJSONPath_Matches(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name", Op: "matches", Expected: "^A"})
	if !pass {
		t.Fatalf("matches failed")
	}
	failPass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name", Op: "matches", Expected: "^Z"})
	if failPass {
		t.Fatalf("expected matches fail")
	}
}

func TestCheckJSONPath_MatchesNonString(t *testing.T) {
	pass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.id", Op: "matches", Expected: ".+"})
	if pass || !strings.Contains(msg, "not a string") {
		t.Fatalf("got pass=%v msg=%q", pass, msg)
	}
}

func TestCheckJSONPath_MatchesBadRegex(t *testing.T) {
	pass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name", Op: "matches", Expected: "["})
	if pass || !strings.Contains(msg, "invalid regex") {
		t.Fatalf("got pass=%v msg=%q", pass, msg)
	}
}

func TestCheckJSONPath_Type(t *testing.T) {
	cases := []struct {
		path string
		want string
		pass bool
	}{
		{"user.name", "string", true},
		{"user.id", "integer", true},
		{"user.score", "number", true},
		{"user.admin", "boolean", true},
		{"user.tags", "array", true},
		{"user", "object", true},
		{"empty", "null", true},
		{"user.name", "integer", false},
	}
	for _, tc := range cases {
		t.Run(tc.path+"_"+tc.want, func(t *testing.T) {
			pass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: tc.path, Op: "type", Expected: tc.want})
			if pass != tc.pass {
				t.Fatalf("pass=%v want=%v msg=%q", pass, tc.pass, msg)
			}
		})
	}
}

func TestCheckJSONPath_MinLength(t *testing.T) {
	pass, _ := checkJSONPath(bodyJSON(), Test{JSONPath: "user.tags", Op: "min_length", Expected: "2"})
	if !pass {
		t.Fatalf("expected min_length pass")
	}
	failPass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.tags", Op: "min_length", Expected: "5"})
	if failPass || !strings.Contains(msg, "length 2") {
		t.Fatalf("expected len fail, got %q", msg)
	}
}

func TestCheckJSONPath_MinLengthInvalidExpected(t *testing.T) {
	pass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.tags", Op: "min_length", Expected: "abc"})
	if pass || !strings.Contains(msg, "invalid min_length") {
		t.Fatalf("got %v %q", pass, msg)
	}
}

func TestCheckJSONPath_MinLengthOnScalar(t *testing.T) {
	pass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.id", Op: "min_length", Expected: "1"})
	if pass || !strings.Contains(msg, "no length") {
		t.Fatalf("got %v %q", pass, msg)
	}
}

func TestCheckJSONPath_UnknownOp(t *testing.T) {
	pass, msg := checkJSONPath(bodyJSON(), Test{JSONPath: "user.name", Op: "weird"})
	if pass || !strings.Contains(msg, "unknown jsonpath op") {
		t.Fatalf("got %v %q", pass, msg)
	}
}

func TestJSONType_Fallback(t *testing.T) {
	if got := jsonType(int32(1)); got == "" {
		t.Fatalf("expected non-empty fallback type")
	}
}

func TestLengthOf(t *testing.T) {
	if lengthOf("abc") != 3 {
		t.Fatal("string len")
	}
	if lengthOf([]any{1, 2}) != 2 {
		t.Fatal("array len")
	}
	if lengthOf(map[string]any{"a": 1}) != 1 {
		t.Fatal("map len")
	}
	if lengthOf(42) != -1 {
		t.Fatal("scalar len")
	}
}

func TestCompareEquals_NumberMismatch(t *testing.T) {
	if compareEquals(float64(1), "abc") {
		t.Fatal("should not equal")
	}
}

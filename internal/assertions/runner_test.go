package assertions

import (
	"net/http"
	"strings"
	"testing"
)

func TestRun_StatusExactPass(t *testing.T) {
	results := Run(
		[]Test{{Kind: "status", Expected: "200"}},
		ResponseSnapshot{Status: 200},
	)
	if len(results) != 1 || !results[0].Pass {
		t.Fatalf("expected pass, got %+v", results)
	}
}

func TestRun_StatusExactFail(t *testing.T) {
	results := Run(
		[]Test{{Kind: "status", Expected: "200"}},
		ResponseSnapshot{Status: 404},
	)
	if results[0].Pass {
		t.Fatalf("expected fail")
	}
	if !strings.Contains(results[0].Message, "got 404") || !strings.Contains(results[0].Message, "expected 200") {
		t.Fatalf("message missing expected/actual: %q", results[0].Message)
	}
}

func TestCheckStatus_BucketPatterns(t *testing.T) {
	cases := []struct {
		name     string
		actual   int
		expected string
		pass     bool
	}{
		{"2xx pass", 201, "2xx", true},
		{"2xx fail", 301, "2xx", false},
		{"4xx pass", 404, "4xx", true},
		{"5xx pass", 599, "5xx", true},
		{"5xx fail", 200, "5xx", false},
		{"empty expected", 200, "", false},
		{"invalid pattern", 200, "axx", false},
		{"invalid number", 200, "abc", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pass, msg := checkStatus(tc.actual, tc.expected)
			if pass != tc.pass {
				t.Fatalf("pass=%v want=%v msg=%q", pass, tc.pass, msg)
			}
			if !pass && msg == "" {
				t.Fatalf("expected failure message")
			}
		})
	}
}

func TestCheckMaxDuration_Pass(t *testing.T) {
	pass, msg := checkMaxDuration(100, "200")
	if !pass || msg != "" {
		t.Fatalf("pass=%v msg=%q", pass, msg)
	}
}

func TestCheckMaxDuration_Fail(t *testing.T) {
	pass, msg := checkMaxDuration(500, "200")
	if pass {
		t.Fatalf("expected fail")
	}
	if !strings.Contains(msg, "500") || !strings.Contains(msg, "200") {
		t.Fatalf("msg missing values: %q", msg)
	}
}

func TestCheckMaxDuration_Invalid(t *testing.T) {
	pass, msg := checkMaxDuration(50, "abc")
	if pass || !strings.Contains(msg, "invalid") {
		t.Fatalf("expected invalid, got pass=%v msg=%q", pass, msg)
	}
}

func TestCheckHeader_Exists(t *testing.T) {
	h := http.Header{}
	h.Set("X-Foo", "bar")
	pass, _ := checkHeader(h, Test{JSONPath: "X-Foo"})
	if !pass {
		t.Fatalf("expected pass for exists")
	}
}

func TestCheckHeader_ExistsMissingName(t *testing.T) {
	pass, msg := checkHeader(http.Header{}, Test{})
	if pass || !strings.Contains(msg, "missing header name") {
		t.Fatalf("got pass=%v msg=%q", pass, msg)
	}
}

func TestCheckHeader_NotExists(t *testing.T) {
	h := http.Header{}
	passMissing, _ := checkHeader(h, Test{JSONPath: "X-Foo", Op: "not_exists"})
	if !passMissing {
		t.Fatalf("expected pass when absent")
	}
	h.Set("X-Foo", "v")
	passPresent, msg := checkHeader(h, Test{JSONPath: "X-Foo", Op: "not_exists"})
	if passPresent {
		t.Fatalf("expected fail when present, msg=%q", msg)
	}
}

func TestCheckHeader_Equals(t *testing.T) {
	h := http.Header{}
	h.Set("X-Foo", "bar")
	pass, _ := checkHeader(h, Test{JSONPath: "X-Foo", Op: "equals", Expected: "bar"})
	if !pass {
		t.Fatalf("expected equals pass")
	}
	failPass, msg := checkHeader(h, Test{JSONPath: "X-Foo", Op: "equals", Expected: "baz"})
	if failPass || !strings.Contains(msg, "bar") || !strings.Contains(msg, "baz") {
		t.Fatalf("equals fail msg missing values: %q", msg)
	}
}

func TestCheckHeader_Contains(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json; charset=utf-8")
	pass, _ := checkHeader(h, Test{JSONPath: "Content-Type", Op: "contains", Expected: "json"})
	if !pass {
		t.Fatalf("expected contains pass")
	}
	failPass, _ := checkHeader(h, Test{JSONPath: "Content-Type", Op: "contains", Expected: "xml"})
	if failPass {
		t.Fatalf("expected contains fail")
	}
}

func TestCheckHeader_Matches(t *testing.T) {
	h := http.Header{}
	h.Set("X-Trace", "abc-123")
	pass, _ := checkHeader(h, Test{JSONPath: "X-Trace", Op: "matches", Expected: `^[a-z]+-\d+$`})
	if !pass {
		t.Fatalf("expected regex pass")
	}
	failPass, _ := checkHeader(h, Test{JSONPath: "X-Trace", Op: "matches", Expected: `^\d+$`})
	if failPass {
		t.Fatalf("expected regex fail")
	}
	badPass, msg := checkHeader(h, Test{JSONPath: "X-Trace", Op: "matches", Expected: `[`})
	if badPass || !strings.Contains(msg, "invalid regex") {
		t.Fatalf("expected invalid regex, got %q", msg)
	}
}

func TestCheckHeader_UnknownOp(t *testing.T) {
	h := http.Header{}
	h.Set("X-Foo", "bar")
	pass, msg := checkHeader(h, Test{JSONPath: "X-Foo", Op: "weird"})
	if pass || !strings.Contains(msg, "unknown header op") {
		t.Fatalf("expected unknown op, got %q", msg)
	}
}

func TestCheckBody_Contains(t *testing.T) {
	pass, _ := checkBody("hello world", Test{Op: "contains", Expected: "world"})
	if !pass {
		t.Fatalf("expected pass")
	}
	failPass, msg := checkBody("hello", Test{Op: "contains", Expected: "world"})
	if failPass || msg == "" {
		t.Fatalf("expected fail with msg")
	}
}

func TestCheckBody_Matches(t *testing.T) {
	pass, _ := checkBody("abc123", Test{Op: "matches", Expected: `\d+`})
	if !pass {
		t.Fatalf("expected pass")
	}
	failPass, _ := checkBody("abc", Test{Op: "matches", Expected: `\d+`})
	if failPass {
		t.Fatalf("expected fail")
	}
}

func TestCheckBody_BadRegex(t *testing.T) {
	pass, msg := checkBody("x", Test{Op: "matches", Expected: "["})
	if pass || !strings.Contains(msg, "invalid regex") {
		t.Fatalf("got %v %q", pass, msg)
	}
}

func TestCheckBody_UnknownOp(t *testing.T) {
	pass, msg := checkBody("x", Test{Op: "weird"})
	if pass || !strings.Contains(msg, "unknown body op") {
		t.Fatalf("got %v %q", pass, msg)
	}
}

func TestDeriveName_ExplicitWins(t *testing.T) {
	got := deriveName(Test{Name: "explicit", Kind: "status", Expected: "200"})
	if got != "explicit" {
		t.Fatalf("got %q", got)
	}
}

func TestDeriveName_Defaults(t *testing.T) {
	cases := map[string]Test{
		"Status 200":        {Kind: "status", Expected: "200"},
		"Max 500ms":         {Kind: "max_duration", Expected: "500"},
		"Header X-Foo  bar": {Kind: "header", JSONPath: "X-Foo", Expected: "bar"},
		"user.id equals 42": {Kind: "jsonpath", JSONPath: "user.id", Op: "equals", Expected: "42"},
		"Body contains hi":  {Kind: "body", Op: "contains", Expected: "hi"},
		"unknown":           {Kind: "unknown"},
	}
	for want, tc := range cases {
		got := deriveName(tc)
		if got != want {
			t.Fatalf("kind=%s got %q want %q", tc.Kind, got, want)
		}
	}
}

func TestRun_Aggregation(t *testing.T) {
	resp := ResponseSnapshot{
		Status:     200,
		DurationMs: 50,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       `{"user":{"id":1,"name":"Ada"},"items":[10,20]}`,
	}
	tests := []Test{
		{Kind: "status", Expected: "2xx"},
		{Kind: "max_duration", Expected: "100"},
		{Kind: "header", JSONPath: "Content-Type", Op: "contains", Expected: "json"},
		{Kind: "jsonpath", JSONPath: "user.name", Op: "equals", Expected: "Ada"},
		{Kind: "jsonpath", JSONPath: "items.0", Op: "equals", Expected: "10"},
		{Kind: "body", Op: "contains", Expected: "Ada"},
		{Kind: "status", Expected: "500"},
		{Kind: "jsonpath", JSONPath: "user.name", Op: "equals", Expected: "Bob"},
	}
	results := Run(tests, resp)
	if len(results) != len(tests) {
		t.Fatalf("got %d results", len(results))
	}
	pass, fail := 0, 0
	for _, r := range results {
		if r.Pass {
			pass++
		} else {
			fail++
		}
	}
	if pass != 6 || fail != 2 {
		t.Fatalf("pass=%d fail=%d", pass, fail)
	}
}

func TestRun_UnknownKind(t *testing.T) {
	results := Run([]Test{{Kind: "telepathy"}}, ResponseSnapshot{})
	if results[0].Pass || !strings.Contains(results[0].Message, "unknown test kind") {
		t.Fatalf("got %+v", results[0])
	}
}

func TestRun_NonJSONBodyJSONPath(t *testing.T) {
	results := Run(
		[]Test{{Kind: "jsonpath", JSONPath: "x", Op: "exists"}},
		ResponseSnapshot{Body: "not json"},
	)
	if results[0].Pass || !strings.Contains(results[0].Message, "not JSON") {
		t.Fatalf("got %+v", results[0])
	}
}

func TestRun_PreservesIDAndName(t *testing.T) {
	results := Run(
		[]Test{{ID: "t1", Name: "my check", Kind: "status", Expected: "200"}},
		ResponseSnapshot{Status: 200},
	)
	if results[0].ID != "t1" || results[0].Name != "my check" || results[0].Kind != "status" {
		t.Fatalf("got %+v", results[0])
	}
}

package laravel

import "testing"

func TestParseRules_SplitsPipes(t *testing.T) {
	got := parseRules("required|string|max:255")
	want := []string{"required", "string", "max:255"}
	if !equalStringSlice(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestParseRules_TrimsAndDropsEmpty(t *testing.T) {
	got := parseRules(" required | | email ")
	want := []string{"required", "email"}
	if !equalStringSlice(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func TestParseRules_Empty(t *testing.T) {
	if len(parseRules("")) != 0 {
		t.Fatalf("want empty")
	}
}

func TestHasRule_MatchesExactAndPrefix(t *testing.T) {
	rules := []string{"Required", "min:5"}
	if !hasRule(rules, "required") {
		t.Fatal("want match for required (case-insensitive)")
	}
	if !hasRule(rules, "min") {
		t.Fatal("want match for min via min:N prefix")
	}
	if hasRule(rules, "max") {
		t.Fatal("did not expect max")
	}
}

func TestInferType_Cases(t *testing.T) {
	cases := []struct {
		rules []string
		want  string
	}{
		{[]string{"required", "email"}, "email"},
		{[]string{"uuid"}, "uuid"},
		{[]string{"url"}, "url"},
		{[]string{"integer"}, "integer"},
		{[]string{"numeric"}, "numeric"},
		{[]string{"boolean"}, "boolean"},
		{[]string{"array"}, "array"},
		{[]string{"date_format:Y-m-d"}, "date"},
		{[]string{"file"}, "file"},
		{[]string{"image"}, "file"},
		{[]string{"json"}, "object"},
		{[]string{"required", "string", "min:3"}, "string"},
		{nil, "string"},
	}
	for _, c := range cases {
		got := inferType(c.rules)
		if got != c.want {
			t.Errorf("rules=%v want=%s got=%s", c.rules, c.want, got)
		}
	}
}

func TestIsRequired(t *testing.T) {
	if !isRequired([]string{"required", "string"}) {
		t.Fatal("want required")
	}
	if !isRequired([]string{"required_if:foo,bar"}) {
		t.Fatal("want required (required_if)")
	}
	if isRequired([]string{"nullable", "string"}) {
		t.Fatal("did not want required")
	}
}

func TestHasConfirmed(t *testing.T) {
	if !hasConfirmed([]string{"required", "confirmed"}) {
		t.Fatal("want true")
	}
	if hasConfirmed([]string{"required"}) {
		t.Fatal("want false")
	}
}

func TestParseRules_CustomRulesPassThrough(t *testing.T) {
	got := parseRules("required|regex:/^[A-Z]+$/|in:a,b,c")
	want := []string{"required", "regex:/^[A-Z]+$/", "in:a,b,c"}
	if !equalStringSlice(got, want) {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

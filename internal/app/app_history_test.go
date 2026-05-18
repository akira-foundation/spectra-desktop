package app

import "testing"

func TestParseTestResults_EmptyReturnsNil(t *testing.T) {
	if got := parseTestResults(""); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestParseTestResults_ValidJSON(t *testing.T) {
	raw := `[{"name":"a","kind":"k","pass":true},{"name":"b","kind":"k","pass":false,"message":"oops"}]`
	got := parseTestResults(raw)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Name != "a" || !got[0].Pass {
		t.Fatalf("got[0]=%+v", got[0])
	}
	if got[1].Pass || got[1].Message != "oops" {
		t.Fatalf("got[1]=%+v", got[1])
	}
}

func TestParseTestResults_InvalidJSONReturnsNil(t *testing.T) {
	if got := parseTestResults("not json"); got != nil {
		t.Fatalf("got %v", got)
	}
}

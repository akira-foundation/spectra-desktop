package app

import "testing"

func TestEndpointTestKey_UppercasesMethod(t *testing.T) {
	if got := endpointTestKey("get", "/users"); got != "GET /users" {
		t.Fatalf("got %q", got)
	}
	if got := endpointTestKey("Post", "/x"); got != "POST /x" {
		t.Fatalf("got %q", got)
	}
}

func TestTopNBy_SortsAndTruncates(t *testing.T) {
	in := []EndpointMetricDTO{
		{Path: "/a", Count: 3},
		{Path: "/b", Count: 10},
		{Path: "/c", Count: 1},
	}
	less := func(a, b EndpointMetricDTO) bool { return a.Count > b.Count }
	got := topNBy(in, less, 2)
	if len(got) != 2 || got[0].Path != "/b" || got[1].Path != "/a" {
		t.Fatalf("got %+v", got)
	}
}

func TestTopNBy_DoesNotMutateInput(t *testing.T) {
	in := []EndpointMetricDTO{{Path: "/a", Count: 1}, {Path: "/b", Count: 2}}
	less := func(a, b EndpointMetricDTO) bool { return a.Count > b.Count }
	_ = topNBy(in, less, 5)
	if in[0].Path != "/a" || in[1].Path != "/b" {
		t.Fatalf("mutated input: %+v", in)
	}
}

func TestTopNBy_ZeroNReturnsAll(t *testing.T) {
	in := []EndpointMetricDTO{{Path: "/a"}, {Path: "/b"}, {Path: "/c"}}
	got := topNBy(in, func(a, b EndpointMetricDTO) bool { return false }, 0)
	if len(got) != 3 {
		t.Fatalf("got %d", len(got))
	}
}

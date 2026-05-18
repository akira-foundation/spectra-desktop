package app

import "testing"

func TestEnsureSlice_NilBecomesEmpty(t *testing.T) {
	got := ensureSlice(nil)
	if got == nil || len(got) != 0 {
		t.Fatalf("got %v", got)
	}
}

func TestEnsureSlice_PassThrough(t *testing.T) {
	in := []SnapshotDiffEntry{{Path: "/x"}}
	got := ensureSlice(in)
	if len(got) != 1 || got[0].Path != "/x" {
		t.Fatalf("got %v", got)
	}
}

func TestStringSliceEqual(t *testing.T) {
	cases := []struct {
		a, b []string
		want bool
	}{
		{nil, nil, true},
		{[]string{}, nil, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a", "b"}, []string{"a"}, false},
	}
	for i, c := range cases {
		if got := stringSliceEqual(c.a, c.b); got != c.want {
			t.Fatalf("case %d: got %v want %v", i, got, c.want)
		}
	}
}

func TestIndexSnapshot_KeysByMethodAndPath(t *testing.T) {
	items := []snapshotEndpoint{
		{Method: "GET", Path: "/a"},
		{Method: "POST", Path: "/a"},
	}
	idx := indexSnapshot(items)
	if len(idx) != 2 {
		t.Fatalf("len=%d", len(idx))
	}
	if _, ok := idx["GET /a"]; !ok {
		t.Fatal("missing GET /a")
	}
	if _, ok := idx["POST /a"]; !ok {
		t.Fatal("missing POST /a")
	}
}

func TestCompareEndpoint_DetectsAllChanges(t *testing.T) {
	a := snapshotEndpoint{Method: "GET", Path: "/x", Handler: "A", AuthRole: "user", SchemaHash: "h1", Middleware: []string{"a"}}
	b := snapshotEndpoint{Method: "GET", Path: "/x", Handler: "B", AuthRole: "admin", SchemaHash: "h2", Middleware: []string{"a", "b"}}
	changes := compareEndpoint(a, b)
	want := map[string]bool{"handler": true, "authRole": true, "schema": true, "middleware": true}
	if len(changes) != 4 {
		t.Fatalf("got %v", changes)
	}
	for _, c := range changes {
		if !want[c] {
			t.Fatalf("unexpected change %q", c)
		}
	}
}

func TestCompareEndpoint_NoChanges(t *testing.T) {
	a := snapshotEndpoint{Method: "GET", Path: "/x", Handler: "A", SchemaHash: "h", Middleware: []string{"m"}}
	if changes := compareEndpoint(a, a); len(changes) != 0 {
		t.Fatalf("expected no changes, got %v", changes)
	}
}

func TestComputeDiff_AddedRemovedChanged(t *testing.T) {
	prev := `[{"method":"GET","path":"/a","handler":"H1"},{"method":"GET","path":"/b","handler":"H"}]`
	cur := `[{"method":"GET","path":"/a","handler":"H2"},{"method":"GET","path":"/c","handler":"H"}]`
	diff, err := computeDiff(prev, cur)
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 1 || diff.Added[0].Path != "/c" {
		t.Fatalf("added: %+v", diff.Added)
	}
	if len(diff.Removed) != 1 || diff.Removed[0].Path != "/b" {
		t.Fatalf("removed: %+v", diff.Removed)
	}
	if len(diff.Changed) != 1 || diff.Changed[0].Path != "/a" {
		t.Fatalf("changed: %+v", diff.Changed)
	}
}

func TestComputeDiff_EmptyPrevious(t *testing.T) {
	cur := `[{"method":"GET","path":"/x"}]`
	diff, err := computeDiff("", cur)
	if err != nil {
		t.Fatal(err)
	}
	if len(diff.Added) != 1 || len(diff.Removed) != 0 || len(diff.Changed) != 0 {
		t.Fatalf("unexpected: %+v", diff)
	}
}

func TestComputeDiff_InvalidJSON(t *testing.T) {
	if _, err := computeDiff("nope", "[]"); err == nil {
		t.Fatal("expected error")
	}
}

func TestDecodeSnapshotPayload_Empty(t *testing.T) {
	got, err := decodeSnapshotPayload("")
	if err != nil || got != nil {
		t.Fatalf("got %v err=%v", got, err)
	}
}

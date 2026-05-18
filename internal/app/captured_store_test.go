package app

import (
	"context"
	"sort"
	"testing"
)

func newTestStore() *capturedStore {
	return newCapturedStore(nil, func() context.Context { return context.Background() })
}

func TestCapturedStore_SetAndValues(t *testing.T) {
	s := newTestStore()
	s.set("p1", "token", "abc", "POST /login")
	s.set("p1", "user_id", "42", "POST /login")
	vals := s.values("p1")
	if vals["token"] != "abc" || vals["user_id"] != "42" {
		t.Fatalf("got %v", vals)
	}
	if s.values("missing") != nil {
		t.Fatal("expected nil for unknown project")
	}
}

func TestCapturedStore_ListSortedByName(t *testing.T) {
	s := newTestStore()
	s.set("p1", "zeta", "1", "k")
	s.set("p1", "alpha", "2", "k")
	s.set("p1", "mid", "3", "k")
	list := s.list("p1")
	if len(list) != 3 {
		t.Fatalf("len=%d", len(list))
	}
	names := []string{list[0].Name, list[1].Name, list[2].Name}
	if !sort.StringsAreSorted(names) {
		t.Fatalf("not sorted: %v", names)
	}
}

func TestCapturedStore_ListEmptyReturnsEmptySlice(t *testing.T) {
	s := newTestStore()
	got := s.list("unknown")
	if got == nil || len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestCapturedStore_ClearRemovesProject(t *testing.T) {
	s := newTestStore()
	s.set("p1", "k", "v", "ep")
	s.clear("p1")
	if s.values("p1") != nil {
		t.Fatal("expected cleared")
	}
}

func TestCapturedStore_PruneByEndpointKeepsKeyed(t *testing.T) {
	s := newTestStore()
	s.set("p1", "keep", "1", "ep1")
	s.set("p1", "drop", "2", "ep1")
	s.set("p1", "other", "3", "ep2")
	s.pruneByEndpoint("p1", "ep1", map[string]bool{"keep": true})
	vals := s.values("p1")
	if _, ok := vals["drop"]; ok {
		t.Fatalf("expected drop pruned: %v", vals)
	}
	if vals["keep"] != "1" {
		t.Fatalf("expected keep retained: %v", vals)
	}
	if vals["other"] != "3" {
		t.Fatalf("expected other retained: %v", vals)
	}
}

func TestCapturedStore_EnsureLoadedNoRepoIsNoop(t *testing.T) {
	s := newTestStore()
	s.ensureLoaded("p1")
	if vals := s.values("p1"); vals != nil {
		t.Fatalf("expected nil, got %v", vals)
	}
}

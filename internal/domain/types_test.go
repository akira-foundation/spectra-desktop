package domain

import (
	"testing"
	"time"
)

func TestProjectAuth_OptionalExpires(t *testing.T) {
	var a ProjectAuth
	if a.ExpiresAt != nil {
		t.Fatalf("expected nil ExpiresAt")
	}
	now := time.Now().UTC()
	a.ExpiresAt = &now
	if !a.ExpiresAt.Equal(now) {
		t.Fatalf("ExpiresAt round trip failed")
	}
}

func TestEnvironment_NilVars(t *testing.T) {
	var e Environment
	if e.Vars != nil {
		t.Fatalf("expected nil Vars")
	}
	e.Vars = map[string]string{"k": "v"}
	if e.Vars["k"] != "v" {
		t.Fatalf("vars set failed")
	}
}

func TestEnvironmentInput_ZeroValue(t *testing.T) {
	var in EnvironmentInput
	if in.Vars != nil || in.SortOrder != 0 {
		t.Fatalf("zero input not empty: %+v", in)
	}
}

func TestCollection_EmptyItems(t *testing.T) {
	var c Collection
	if c.Items != nil {
		t.Fatalf("expected nil Items")
	}
}

func TestCollectionItem_ZeroValue(t *testing.T) {
	var it CollectionItem
	if it.SkipOnFailure || it.IterateDataset {
		t.Fatalf("zero item not empty: %+v", it)
	}
}

func TestEndpointCapture_ZeroValue(t *testing.T) {
	var c EndpointCapture
	if c.SortOrder != 0 || c.Name != "" {
		t.Fatalf("zero capture not empty: %+v", c)
	}
}

func TestEndpointTest_ZeroValue(t *testing.T) {
	var et EndpointTest
	if et.SortOrder != 0 || et.Op != "" {
		t.Fatalf("zero test not empty: %+v", et)
	}
}

func TestHistoryEntry_ZeroValue(t *testing.T) {
	var h HistoryEntry
	if h.ResponseStatus != 0 || h.DurationMs != 0 || h.SizeBytes != 0 {
		t.Fatalf("zero entry not empty: %+v", h)
	}
}

func TestEndpointSnapshot_ZeroValue(t *testing.T) {
	var s EndpointSnapshot
	if s.EndpointCount != 0 {
		t.Fatalf("zero snapshot not empty: %+v", s)
	}
}

func TestLicense_ZeroValue(t *testing.T) {
	var l License
	if l.CancelAtPeriodEnd || l.GracePeriod {
		t.Fatalf("zero license not empty: %+v", l)
	}
}

func TestUsageBufferEntry_ZeroValue(t *testing.T) {
	var u UsageBufferEntry
	if u.Flushed || u.Amount != 0 {
		t.Fatalf("zero usage not empty: %+v", u)
	}
}

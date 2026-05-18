package app

import "testing"

func TestHashString_EmptyReturnsEmpty(t *testing.T) {
	if got := hashString(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestHashString_StableSHA256Hex(t *testing.T) {
	a := hashString("hello")
	b := hashString("hello")
	c := hashString("world")
	if a != b {
		t.Fatal("unstable")
	}
	if a == c {
		t.Fatal("collisions for different inputs")
	}
	if len(a) != 64 {
		t.Fatalf("len: %d", len(a))
	}
}

func TestExtractSchemaFields_EmptyReturnsNil(t *testing.T) {
	if got := extractSchemaFields(""); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestExtractSchemaFields_ParsesFields(t *testing.T) {
	raw := `{"fields":[{"name":"id","type":"integer","required":true},{"name":"x","type":"string"}]}`
	got := extractSchemaFields(raw)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].Name != "id" || got[0].Type != "integer" || !got[0].Required {
		t.Fatalf("first: %+v", got[0])
	}
	if got[1].Name != "x" || got[1].Required {
		t.Fatalf("second: %+v", got[1])
	}
}

func TestExtractSchemaFields_InvalidJSONReturnsNil(t *testing.T) {
	if got := extractSchemaFields("not json"); got != nil {
		t.Fatalf("got %v", got)
	}
}

func TestStableSchemaHash_EmptyReturnsEmpty(t *testing.T) {
	if got := stableSchemaHash(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestStableSchemaHash_IgnoresUnknownTopLevelFields(t *testing.T) {
	a := `{"source":"x","confidence":"high","fields":[{"name":"id","type":"int"}]}`
	b := `{"source":"x","confidence":"high","fields":[{"name":"id","type":"int"}],"examples":{"id":42}}`
	if stableSchemaHash(a) != stableSchemaHash(b) {
		t.Fatal("expected hash to be insensitive to unknown fields")
	}
}

func TestStableSchemaHash_FallsBackOnInvalidJSON(t *testing.T) {
	got := stableSchemaHash("not json")
	if got == "" {
		t.Fatal("expected fallback hash")
	}
	if got != hashString("not json") {
		t.Fatalf("expected fallback to hashString, got %q", got)
	}
}

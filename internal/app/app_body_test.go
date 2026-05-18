package app

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFallbackBodyValueGen_TypesMapToZeroValues(t *testing.T) {
	g := fallbackBodyValueGen{}
	cases := map[string]any{
		"integer": 0,
		"int":     0,
		"number":  0,
		"numeric": 0,
		"boolean": false,
		"bool":    false,
	}
	for typ, want := range cases {
		got := g.GenerateValue("f", typ, nil)
		if got != want {
			t.Fatalf("type %s: got %v want %v", typ, got, want)
		}
	}
	if arr, ok := g.GenerateValue("f", "array", nil).([]any); !ok || len(arr) != 0 {
		t.Fatalf("array: got %v", arr)
	}
	if m, ok := g.GenerateValue("f", "object", nil).(map[string]any); !ok || len(m) != 0 {
		t.Fatalf("object: got %v", m)
	}
	if got := g.GenerateValue("f", "string", nil); got != "" {
		t.Fatalf("default string: got %v", got)
	}
}

func TestFallbackBodyValueGen_TypeIsCaseInsensitive(t *testing.T) {
	g := fallbackBodyValueGen{}
	if g.GenerateValue("f", "BOOLEAN", nil) != false {
		t.Fatal("BOOLEAN should map to false")
	}
}

func TestOrderedMap_SetPreservesInsertionOrder(t *testing.T) {
	m := orderedMap{}
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("a", 99)
	if len(m.Keys) != 2 || m.Keys[0] != "a" || m.Keys[1] != "b" {
		t.Fatalf("keys: %v", m.Keys)
	}
	if m.Values["a"] != 99 {
		t.Fatalf("value not overwritten: %v", m.Values["a"])
	}
}

func TestOrderedMap_UnmarshalKeepsOrder(t *testing.T) {
	in := []byte(`{"z":1,"a":2,"m":3}`)
	var m orderedMap
	if err := json.Unmarshal(in, &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Keys) != 3 || m.Keys[0] != "z" || m.Keys[1] != "a" || m.Keys[2] != "m" {
		t.Fatalf("order: %v", m.Keys)
	}
}

func TestOrderedMap_UnmarshalRejectsNonObject(t *testing.T) {
	var m orderedMap
	if err := json.Unmarshal([]byte(`[]`), &m); err == nil {
		t.Fatal("expected error on array")
	}
}

func TestMarshalOrdered_StableOutput(t *testing.T) {
	m := orderedMap{}
	m.Set("b", 1)
	m.Set("a", "x")
	out, err := marshalOrdered(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"b"`) || strings.Index(out, `"b"`) > strings.Index(out, `"a"`) {
		t.Fatalf("order not preserved: %s", out)
	}
	var back map[string]any
	if err := json.Unmarshal([]byte(out), &back); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
}

func TestInferTypeFromValue_Scalars(t *testing.T) {
	if inferTypeFromValue(true) != "boolean" {
		t.Fatal("bool")
	}
	if inferTypeFromValue(json.Number("3")) != "integer" {
		t.Fatal("int number")
	}
	if inferTypeFromValue(json.Number("3.14")) != "numeric" {
		t.Fatal("float number")
	}
	if inferTypeFromValue(float64(1.2)) != "numeric" {
		t.Fatal("float64")
	}
	if inferTypeFromValue([]any{}) != "array" {
		t.Fatal("array")
	}
	if inferTypeFromValue(map[string]any{}) != "object" {
		t.Fatal("object")
	}
	if inferTypeFromValue(nil) != "string" {
		t.Fatal("nil")
	}
	if inferTypeFromValue("x") != "string" {
		t.Fatal("string")
	}
}

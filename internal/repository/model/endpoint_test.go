package model

import (
	"reflect"
	"testing"
	"time"

	"spectra-desktop/internal/core"
)

func TestEncodeStringSlice_Empty(t *testing.T) {
	if got := encodeStringSlice(nil); got != "[]" {
		t.Fatalf("nil: %q", got)
	}
	if got := encodeStringSlice([]string{}); got != "[]" {
		t.Fatalf("empty: %q", got)
	}
}

func TestEncodeStringSlice_Values(t *testing.T) {
	got := encodeStringSlice([]string{"a", "b"})
	if got != `["a","b"]` {
		t.Fatalf("got %q", got)
	}
}

func TestDecodeStringSlice_EmptyForms(t *testing.T) {
	if decodeStringSlice("") != nil {
		t.Fatalf("empty string should decode nil")
	}
	if decodeStringSlice("[]") != nil {
		t.Fatalf("[] should decode nil")
	}
}

func TestDecodeStringSlice_Malformed(t *testing.T) {
	if decodeStringSlice("not-json") != nil {
		t.Fatalf("malformed should decode nil")
	}
}

func TestDecodeStringSlice_Values(t *testing.T) {
	got := decodeStringSlice(`["a","b"]`)
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("got %+v", got)
	}
}

func TestStringSlice_RoundTrip(t *testing.T) {
	in := []string{"x", "y", "z"}
	out := decodeStringSlice(encodeStringSlice(in))
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("roundtrip mismatch: %+v vs %+v", in, out)
	}
}

func TestEncodeParameters_Empty(t *testing.T) {
	if got := encodeParameters(nil); got != "[]" {
		t.Fatalf("nil: %q", got)
	}
}

func TestDecodeParameters_Malformed(t *testing.T) {
	if decodeParameters("bad") != nil {
		t.Fatalf("malformed should decode nil")
	}
	if decodeParameters("") != nil || decodeParameters("[]") != nil {
		t.Fatalf("empty should decode nil")
	}
}

func TestParameters_RoundTrip(t *testing.T) {
	in := []core.Parameter{{Name: "id", In: "path", Type: "int", Required: true}}
	out := decodeParameters(encodeParameters(in))
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("roundtrip mismatch: %+v vs %+v", in, out)
	}
}

func TestEndpointFromCore_PreservesFields(t *testing.T) {
	scanned := time.Now().UTC()
	now := scanned.Add(time.Second)
	ep := core.Endpoint{
		ID:                "e1",
		Method:            core.MethodPost,
		Path:              "/x",
		Name:              "create",
		Handler:           "H",
		Middleware:        []string{"auth", "api"},
		Parameters:        []core.Parameter{{Name: "n", In: "body"}},
		Tags:              []string{"t"},
		Source:            core.EndpointSource{File: "f.go", Line: 10},
		Framework:         "laravel",
		Confidence:        0.5,
		RequestSchema:     `{"a":1}`,
		AuthRole:          core.AuthRoleLogin,
		AuthHint:          "hint",
		AuthRoleOverride:  core.AuthRoleRefresh,
		TokenPathOverride: "token.access",
	}
	m := EndpointFromCore("p1", ep, scanned, now)
	if m.ProjectID != "p1" || m.Method != "POST" || m.SourceFile != "f.go" || m.SourceLine != 10 {
		t.Fatalf("mapping failed: %+v", m)
	}
	if m.AuthRole != "login" || m.AuthRoleOverride != "refresh" || m.TokenPathOverride != "token.access" {
		t.Fatalf("auth mapping failed: %+v", m)
	}
	if !m.ScannedAt.Equal(scanned) || !m.CreatedAt.Equal(now) || !m.UpdatedAt.Equal(now) {
		t.Fatalf("time mapping failed: %+v", m)
	}
}

func TestEndpoint_ToCoreFromCore_RoundTrip(t *testing.T) {
	now := time.Now().UTC()
	ep := core.Endpoint{
		ID:         "e2",
		Method:     core.MethodGet,
		Path:       "/y",
		Middleware: []string{"a"},
		Parameters: []core.Parameter{{Name: "id", In: "path", Required: true}},
		Tags:       []string{"t1", "t2"},
		Source:     core.EndpointSource{File: "g.go", Line: 3},
		Framework:  "laravel",
		AuthRole:   core.AuthRoleLogin,
	}
	out := EndpointFromCore("p", ep, now, now).ToCore()
	if out.ID != ep.ID || out.Method != ep.Method || out.Path != ep.Path {
		t.Fatalf("identity mismatch: %+v", out)
	}
	if !reflect.DeepEqual(out.Middleware, ep.Middleware) {
		t.Fatalf("middleware mismatch: %+v", out.Middleware)
	}
	if !reflect.DeepEqual(out.Parameters, ep.Parameters) {
		t.Fatalf("parameters mismatch: %+v", out.Parameters)
	}
	if !reflect.DeepEqual(out.Tags, ep.Tags) {
		t.Fatalf("tags mismatch: %+v", out.Tags)
	}
	if out.Source != ep.Source || out.AuthRole != ep.AuthRole {
		t.Fatalf("source/auth mismatch: %+v", out)
	}
}

func TestEndpoint_ToCore_EmptyJSONFields(t *testing.T) {
	m := Endpoint{
		ID:         "e",
		Method:     "GET",
		Path:       "/x",
		Middleware: "[]",
		Parameters: "",
		Tags:       "[]",
	}
	out := m.ToCore()
	if out.Middleware != nil || out.Parameters != nil || out.Tags != nil {
		t.Fatalf("expected nil slices, got %+v", out)
	}
}

package openapi

import (
	"encoding/json"
	"strings"
	"testing"

	"spectra-desktop/internal/core"
)

func TestBuildRequestSchema_InvalidJSONReturnsNil(t *testing.T) {
	if s := buildRequestSchema(core.Endpoint{RequestSchema: "not-json"}); s != nil {
		t.Fatalf("expected nil, got %+v", s)
	}
}

func TestBuildRequestSchema_EmptyFieldsReturnsNil(t *testing.T) {
	if s := buildRequestSchema(core.Endpoint{RequestSchema: `{"fields":[]}`}); s != nil {
		t.Fatalf("expected nil, got %+v", s)
	}
}

func TestMapType_AllVariants(t *testing.T) {
	cases := map[string]struct {
		typ, format string
	}{
		"integer": {"integer", ""},
		"numeric": {"integer", ""},
		"boolean": {"boolean", ""},
		"object":  {"object", ""},
		"email":   {"string", "email"},
		"date":    {"string", "date"},
		"uuid":    {"string", "uuid"},
		"url":     {"string", "uri"},
		"unknown": {"string", ""},
	}
	for in, want := range cases {
		got := mapType(in)
		if got.Type != want.typ || got.Format != want.format {
			t.Fatalf("mapType(%q)=%+v want %+v", in, got, want)
		}
	}
	arr := mapType("array")
	if arr.Type != "array" || arr.Items == nil || arr.Items.Type != "string" {
		t.Fatalf("array mapping: %+v", arr)
	}
}

func TestToJSON_RoundtripPreservesStructure(t *testing.T) {
	p := newProject()
	p.BaseURL = "https://api.example.com"
	eps := []core.Endpoint{{
		Method:        core.MethodPost,
		Path:          "/users/{id}",
		Name:          "updateUser",
		Tags:          []string{"users"},
		Middleware:    []string{"auth"},
		RequestSchema: `{"fields":[{"name":"name","type":"string","required":true}]}`,
	}}
	spec := Build(p, eps)
	out, err := ToJSON(spec)
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if !strings.Contains(out, `"openapi": "3.0.3"`) {
		t.Fatalf("missing openapi version in output:\n%s", out)
	}
	var back Spec
	if err := json.Unmarshal([]byte(out), &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Info.Title != "Demo API" {
		t.Fatalf("title roundtrip: %q", back.Info.Title)
	}
	if len(back.Servers) != 1 || back.Servers[0].URL != "https://api.example.com" {
		t.Fatalf("servers roundtrip: %+v", back.Servers)
	}
	op := back.Paths["/users/{id}"].Post
	if op == nil {
		t.Fatal("post op missing after roundtrip")
	}
	if op.OperationID != "updateUser" {
		t.Fatalf("operationId roundtrip: %q", op.OperationID)
	}
	if len(op.Parameters) != 1 || op.Parameters[0].Name != "id" || op.Parameters[0].In != "path" {
		t.Fatalf("params roundtrip: %+v", op.Parameters)
	}
	if op.RequestBody == nil || op.RequestBody.Content["application/json"].Schema.Properties["name"].Type != "string" {
		t.Fatalf("requestBody roundtrip: %+v", op.RequestBody)
	}
	if len(op.Security) != 1 {
		t.Fatalf("security roundtrip: %+v", op.Security)
	}
	if back.Components == nil || back.Components.SecuritySchemes["bearerAuth"].Scheme != "bearer" {
		t.Fatalf("components roundtrip: %+v", back.Components)
	}
}

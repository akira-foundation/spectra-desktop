package openapi

import (
	"testing"

	"spectra-desktop/internal/core"
)

func TestBuild_EmptyProject(t *testing.T) {
	spec := Build(newProject(), nil)
	if spec.OpenAPI != "3.0.3" {
		t.Fatalf("openapi version: got %q", spec.OpenAPI)
	}
	if spec.Info.Title != "Demo API" || spec.Info.Version != "1.0.0" {
		t.Fatalf("info mismatch: %+v", spec.Info)
	}
	if spec.Paths == nil {
		t.Fatal("paths must be non-nil map")
	}
	if len(spec.Paths) != 0 {
		t.Fatalf("expected empty paths, got %d", len(spec.Paths))
	}
	if len(spec.Servers) != 0 {
		t.Fatalf("expected no servers, got %d", len(spec.Servers))
	}
}

func TestBuild_BaseURLPopulatesServers(t *testing.T) {
	p := newProject()
	p.BaseURL = "https://api.example.com"
	spec := Build(p, nil)
	if len(spec.Servers) != 1 || spec.Servers[0].URL != "https://api.example.com" {
		t.Fatalf("servers mismatch: %+v", spec.Servers)
	}
}

func TestBuild_SingleGETEndpoint(t *testing.T) {
	eps := []core.Endpoint{{
		ID:     "e1",
		Method: core.MethodGet,
		Path:   "/health",
		Name:   "healthCheck",
	}}
	spec := Build(newProject(), eps)
	item, ok := spec.Paths["/health"]
	if !ok {
		t.Fatalf("missing /health path; have %+v", spec.Paths)
	}
	if item.Get == nil {
		t.Fatal("GET operation missing")
	}
	if item.Get.OperationID != "healthCheck" {
		t.Fatalf("operationId: %q", item.Get.OperationID)
	}
	if _, ok := item.Get.Responses["200"]; !ok {
		t.Fatalf("expected 200 response, got %+v", item.Get.Responses)
	}
	if item.Get.RequestBody != nil {
		t.Fatal("GET must not have requestBody")
	}
}

func TestBuild_POSTWithJSONBody(t *testing.T) {
	schema := `{"fields":[{"name":"email","type":"email","required":true},{"name":"age","type":"integer"}]}`
	eps := []core.Endpoint{{
		Method:        core.MethodPost,
		Path:          "/users",
		Name:          "createUser",
		RequestSchema: schema,
	}}
	spec := Build(newProject(), eps)
	op := spec.Paths["/users"].Post
	if op == nil {
		t.Fatal("POST op missing")
	}
	if op.RequestBody == nil {
		t.Fatal("requestBody missing")
	}
	mt, ok := op.RequestBody.Content["application/json"]
	if !ok {
		t.Fatalf("application/json missing: %+v", op.RequestBody.Content)
	}
	if mt.Schema == nil || mt.Schema.Type != "object" {
		t.Fatalf("schema type: %+v", mt.Schema)
	}
	emailProp, ok := mt.Schema.Properties["email"]
	if !ok || emailProp.Type != "string" || emailProp.Format != "email" {
		t.Fatalf("email property: %+v", emailProp)
	}
	ageProp, ok := mt.Schema.Properties["age"]
	if !ok || ageProp.Type != "integer" {
		t.Fatalf("age property: %+v", ageProp)
	}
	if len(mt.Schema.Required) != 1 || mt.Schema.Required[0] != "email" {
		t.Fatalf("required: %+v", mt.Schema.Required)
	}
}

func TestBuild_AllHTTPMethods(t *testing.T) {
	methods := []core.HTTPMethod{
		core.MethodGet, core.MethodPost, core.MethodPut,
		core.MethodPatch, core.MethodDelete, core.MethodHead, core.MethodOptions,
	}
	eps := make([]core.Endpoint, 0, len(methods))
	for _, m := range methods {
		eps = append(eps, core.Endpoint{Method: m, Path: "/r", Name: string(m)})
	}
	spec := Build(newProject(), eps)
	item := spec.Paths["/r"]
	ops := []*Operation{item.Get, item.Post, item.Put, item.Patch, item.Delete, item.Head, item.Options}
	for i, op := range ops {
		if op == nil {
			t.Fatalf("op %d (%s) missing", i, methods[i])
		}
	}
}

func TestBuild_UnsupportedMethodSkippedGracefully(t *testing.T) {
	eps := []core.Endpoint{
		{Method: "TRACE", Path: "/x", Name: "trace"},
		{Method: core.MethodGet, Path: "/y", Name: "y"},
	}
	spec := Build(newProject(), eps)
	if _, ok := spec.Paths["/x"]; ok {
		t.Fatal("TRACE should be skipped")
	}
	if _, ok := spec.Paths["/y"]; !ok {
		t.Fatal("GET should still be present")
	}
}

func TestBuild_BodyAllowedOnlyForBodyMethods(t *testing.T) {
	schema := `{"fields":[{"name":"x","type":"string","required":true}]}`
	cases := map[core.HTTPMethod]bool{
		core.MethodGet:    false,
		core.MethodHead:   false,
		core.MethodDelete: false,
		core.MethodPost:   true,
		core.MethodPut:    true,
		core.MethodPatch:  true,
	}
	for method, wantBody := range cases {
		spec := Build(newProject(), []core.Endpoint{{Method: method, Path: "/r", RequestSchema: schema, Name: "n"}})
		item := spec.Paths["/r"]
		var op *Operation
		switch method {
		case core.MethodGet:
			op = item.Get
		case core.MethodHead:
			op = item.Head
		case core.MethodDelete:
			op = item.Delete
		case core.MethodPost:
			op = item.Post
		case core.MethodPut:
			op = item.Put
		case core.MethodPatch:
			op = item.Patch
		}
		hasBody := op != nil && op.RequestBody != nil
		if hasBody != wantBody {
			t.Fatalf("%s: hasBody=%v want=%v", method, hasBody, wantBody)
		}
	}
}

func TestBuild_AuthMiddlewareAddsSecurity(t *testing.T) {
	eps := []core.Endpoint{
		{Method: core.MethodGet, Path: "/me", Name: "me", Middleware: []string{"auth:sanctum"}},
		{Method: core.MethodGet, Path: "/auth/check", Name: "check", Middleware: []string{"Authenticate"}},
		{Method: core.MethodGet, Path: "/open", Name: "open", Middleware: []string{"throttle:60"}},
	}
	spec := Build(newProject(), eps)
	if sec := spec.Paths["/me"].Get.Security; len(sec) != 1 || len(sec[0]["bearerAuth"]) != 0 {
		t.Fatalf("auth: security: %+v", sec)
	}
	if spec.Paths["/auth/check"].Get.Security == nil {
		t.Fatal("Authenticate middleware should mark secured")
	}
	if spec.Paths["/open"].Get.Security != nil {
		t.Fatal("throttle should not be secured")
	}
	scheme, ok := spec.Components.SecuritySchemes["bearerAuth"]
	if !ok || scheme.Type != "http" || scheme.Scheme != "bearer" || scheme.BearerFormat != "JWT" {
		t.Fatalf("bearerAuth scheme: %+v", scheme)
	}
}

func TestBuild_TagsPropagated(t *testing.T) {
	eps := []core.Endpoint{{
		Method: core.MethodGet, Path: "/users", Name: "list", Tags: []string{"users", "v1"},
	}}
	spec := Build(newProject(), eps)
	got := spec.Paths["/users"].Get.Tags
	if len(got) != 2 || got[0] != "users" || got[1] != "v1" {
		t.Fatalf("tags: %+v", got)
	}
}

package openapi

import (
	"testing"

	"spectra-desktop/internal/core"
)

func TestBuild_PathParamsBraceSyntax(t *testing.T) {
	eps := []core.Endpoint{{
		Method: core.MethodGet,
		Path:   "/users/{id}/posts/{postId}",
		Name:   "getUserPost",
	}}
	spec := Build(newProject(), eps)
	op := spec.Paths["/users/{id}/posts/{postId}"].Get
	if op == nil {
		t.Fatal("op missing")
	}
	if len(op.Parameters) != 2 {
		t.Fatalf("params: %+v", op.Parameters)
	}
	for _, p := range op.Parameters {
		if p.In != "path" || !p.Required || p.Schema == nil || p.Schema.Type != "string" {
			t.Fatalf("bad param: %+v", p)
		}
	}
	if op.Parameters[0].Name != "id" || op.Parameters[1].Name != "postId" {
		t.Fatalf("param names: %+v", op.Parameters)
	}
}

func TestBuild_PathParamsColonSyntaxConverted(t *testing.T) {
	eps := []core.Endpoint{{
		Method: core.MethodGet,
		Path:   "/users/:id",
		Name:   "showUser",
	}}
	spec := Build(newProject(), eps)
	if _, ok := spec.Paths["/users/{id}"]; !ok {
		t.Fatalf("colon path not converted: %+v", spec.Paths)
	}
	op := spec.Paths["/users/{id}"].Get
	if len(op.Parameters) != 1 || op.Parameters[0].Name != "id" {
		t.Fatalf("params: %+v", op.Parameters)
	}
}

func TestBuild_RelativePathPrefixed(t *testing.T) {
	eps := []core.Endpoint{{
		Method: core.MethodGet,
		Path:   "ping",
		Name:   "ping",
	}}
	spec := Build(newProject(), eps)
	if _, ok := spec.Paths["/ping"]; !ok {
		t.Fatalf("relative path not normalized: %+v", spec.Paths)
	}
}

func TestExtractPathParams_MixedSyntax(t *testing.T) {
	got := extractPathParams("/a/{x}/b/:y/c")
	if len(got) != 2 || got[0] != "x" || got[1] != "y" {
		t.Fatalf("extractPathParams: %+v", got)
	}
}

func TestAllowsBody_CaseInsensitive(t *testing.T) {
	if allowsBody("get") || allowsBody("GET") {
		t.Fatal("GET should not allow body")
	}
	if !allowsBody("post") || !allowsBody("PATCH") {
		t.Fatal("POST/PATCH should allow body")
	}
}

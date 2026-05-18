package app

import (
	"net/http"
	"strings"
	"testing"
)

func TestToHTTPHeader_MergesMultiValues(t *testing.T) {
	in := map[string][]string{
		"X-A": {"1", "2"},
		"X-B": {"only"},
	}
	got := toHTTPHeader(in)
	if vs := got.Values("X-A"); len(vs) != 2 || vs[0] != "1" || vs[1] != "2" {
		t.Fatalf("X-A values: %v", vs)
	}
	if got.Get("X-B") != "only" {
		t.Fatalf("X-B: %q", got.Get("X-B"))
	}
}

func TestToHTTPHeader_EmptyInput(t *testing.T) {
	got := toHTTPHeader(nil)
	if got == nil {
		t.Fatal("expected non-nil header")
	}
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}

func TestToHTTPHeader_Canonicalizes(t *testing.T) {
	got := toHTTPHeader(map[string][]string{"content-type": {"text/plain"}})
	if got.Get("Content-Type") != "text/plain" {
		t.Fatalf("got %q", got.Get("Content-Type"))
	}
	_ = http.Header(got)
}

func TestJoinURL_AppendsPathToBase(t *testing.T) {
	u, err := joinURL("https://api.test", "/users")
	if err != nil || u != "https://api.test/users" {
		t.Fatalf("got %q err=%v", u, err)
	}
}

func TestJoinURL_NormalisesSlashes(t *testing.T) {
	u, err := joinURL("https://api.test/", "users")
	if err != nil || u != "https://api.test/users" {
		t.Fatalf("got %q err=%v", u, err)
	}
}

func TestJoinURL_PreservesBasePath(t *testing.T) {
	u, err := joinURL("https://api.test/v1/", "/posts")
	if err != nil || u != "https://api.test/v1/posts" {
		t.Fatalf("got %q err=%v", u, err)
	}
}

func TestJoinURL_EmptyPathReturnsBase(t *testing.T) {
	u, err := joinURL("https://api.test/v1", "")
	if err != nil || u != "https://api.test/v1" {
		t.Fatalf("got %q err=%v", u, err)
	}
}

func TestJoinURL_EmptyBaseErrors(t *testing.T) {
	_, err := joinURL("   ", "/x")
	if err == nil || !strings.Contains(err.Error(), "empty base url") {
		t.Fatalf("expected empty base error, got %v", err)
	}
}

func TestJoinURL_InvalidBaseErrors(t *testing.T) {
	_, err := joinURL("http://[::1", "/x")
	if err == nil {
		t.Fatal("expected error on invalid base")
	}
}

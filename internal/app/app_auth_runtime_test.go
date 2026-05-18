package app

import (
	"net/url"
	"strings"
	"testing"

	"spectra-desktop/internal/auth"
)

func TestAppendQueryParams_AddsMissingKeys(t *testing.T) {
	got := appendQueryParams("https://x.test/a", map[string]string{"k": "v"})
	u, _ := url.Parse(got)
	if u.Query().Get("k") != "v" {
		t.Fatalf("got %q", got)
	}
}

func TestAppendQueryParams_DoesNotOverwriteExisting(t *testing.T) {
	got := appendQueryParams("https://x.test/?k=orig", map[string]string{"k": "new"})
	u, _ := url.Parse(got)
	if u.Query().Get("k") != "orig" {
		t.Fatalf("expected orig kept, got %q", got)
	}
}

func TestAppendQueryParams_EmptyParamsReturnsInput(t *testing.T) {
	in := "https://x.test/path?z=1"
	if got := appendQueryParams(in, nil); got != in {
		t.Fatalf("got %q", got)
	}
}

func TestAppendQueryParams_InvalidURLReturnsInput(t *testing.T) {
	in := "://bad"
	if got := appendQueryParams(in, map[string]string{"k": "v"}); got != in {
		t.Fatalf("got %q", got)
	}
}

func TestApplyInjection_HeaderOnlyWhenAbsent(t *testing.T) {
	headers := map[string]string{}
	query := map[string]string{}
	applyInjection(headers, query, auth.HeaderInjection{Header: "X-A", Value: "1"})
	if headers["X-A"] != "1" {
		t.Fatalf("header not set: %v", headers)
	}
	applyInjection(headers, query, auth.HeaderInjection{Header: "X-A", Value: "2"})
	if headers["X-A"] != "1" {
		t.Fatalf("expected existing header kept, got %q", headers["X-A"])
	}
}

func TestApplyInjection_QueryAlwaysOverwrites(t *testing.T) {
	headers := map[string]string{}
	query := map[string]string{"k": "old"}
	applyInjection(headers, query, auth.HeaderInjection{QueryKey: "k", QueryValue: "new"})
	if query["k"] != "new" {
		t.Fatalf("got %v", query)
	}
}

func TestApplyInjection_SkipsEmptyFields(t *testing.T) {
	headers := map[string]string{}
	query := map[string]string{}
	applyInjection(headers, query, auth.HeaderInjection{Header: "X-A"})
	applyInjection(headers, query, auth.HeaderInjection{QueryKey: "k"})
	if len(headers) != 0 || len(query) != 0 {
		t.Fatalf("expected nothing applied: headers=%v query=%v", headers, query)
	}
}

func TestAppendQueryParams_PreservesPath(t *testing.T) {
	got := appendQueryParams("https://x.test/api/v1/users", map[string]string{"page": "2"})
	if !strings.HasPrefix(got, "https://x.test/api/v1/users?") {
		t.Fatalf("got %q", got)
	}
}

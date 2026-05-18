package laravel

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractJSONArray_PicksLargestValidArray(t *testing.T) {
	data := []byte(`prefix garbage [] more [{"method":"GET","uri":"/x"},{"method":"POST","uri":"/y"}] tail`)
	got := extractJSONArray(data)
	if got == nil {
		t.Fatal("want array")
	}
	var routes []rawRoute
	if err := json.Unmarshal(got, &routes); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(routes) != 2 {
		t.Fatalf("want 2, got %d", len(routes))
	}
}

func TestExtractJSONArray_NoArray(t *testing.T) {
	if got := extractJSONArray([]byte(`hello`)); got != nil {
		t.Fatalf("want nil, got %s", string(got))
	}
}

func TestExtractJSONArray_Empty(t *testing.T) {
	if got := extractJSONArray(nil); got != nil {
		t.Fatal("want nil")
	}
}

func TestExtractJSONArray_HandlesBracketsInsideStrings(t *testing.T) {
	data := []byte(`[{"method":"GET","uri":"/has[bracket]","action":"X"}]`)
	got := extractJSONArray(data)
	if got == nil {
		t.Fatal("want array")
	}
	var routes []rawRoute
	if err := json.Unmarshal(got, &routes); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(routes) != 1 || routes[0].URI != "/has[bracket]" {
		t.Fatalf("got %+v", routes)
	}
}

func TestExtractJSONArray_HandlesEscapedQuotes(t *testing.T) {
	data := []byte(`[{"method":"GET","uri":"/x","action":"a\"b"}]`)
	got := extractJSONArray(data)
	if got == nil {
		t.Fatal("want array")
	}
}

func TestMatchBalancedArray_Unbalanced(t *testing.T) {
	if got := matchBalancedArray([]byte(`[1,2`), 0); got != -1 {
		t.Fatalf("want -1, got %d", got)
	}
}

func TestMatchBalancedArray_Balanced(t *testing.T) {
	data := []byte(`[1,[2,3]]`)
	got := matchBalancedArray(data, 0)
	if got != len(data)-1 {
		t.Fatalf("want %d, got %d", len(data)-1, got)
	}
}

func TestRunArtisanRouteList_NoArtisan(t *testing.T) {
	dir := t.TempDir()
	prevOverride := currentPHPOverride()
	SetPHPBinaryOverride("/bin/sh")
	t.Cleanup(func() { SetPHPBinaryOverride(prevOverride) })

	_, err := runArtisanRouteList(context.Background(), dir)
	if !errors.Is(err, ErrArtisanMissing) {
		t.Fatalf("want ErrArtisanMissing, got %v", err)
	}
}

func TestRunArtisanRouteList_NoPHP(t *testing.T) {
	dir := t.TempDir()
	mustTouch(t, filepath.Join(dir, "artisan"))
	prevOverride := currentPHPOverride()
	SetPHPBinaryOverride("/definitely/does/not/exist/php-xyz")
	t.Cleanup(func() { SetPHPBinaryOverride(prevOverride) })

	t.Setenv("PATH", "/definitely-empty")
	t.Setenv("SHELL", "/definitely/not/a/shell")
	_, err := runArtisanRouteList(context.Background(), dir)
	if err == nil {
		t.Fatal("want error")
	}
	var afe *ArtisanFailedError
	if !errors.Is(err, ErrPHPNotFound) && !errors.Is(err, ErrNoRoutes) && !errors.Is(err, ErrInvalidJSON) && !errors.As(err, &afe) {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestRawRoute_JSONRoundtrip_MiddlewareForms(t *testing.T) {
	cases := []string{
		`{"method":"GET","uri":"/u","action":"X@a","middleware":["web","auth"]}`,
		`{"method":"POST","uri":"/u","action":"X@b","middleware":"web,auth"}`,
		`{"method":"GET","uri":"/u","action":"X@c"}`,
	}
	for _, c := range cases {
		var r rawRoute
		if err := json.Unmarshal([]byte(c), &r); err != nil {
			t.Fatalf("unmarshal %s: %v", c, err)
		}
		mw := decodeMiddleware(r.Middleware)
		if strings.Contains(c, "web") && (len(mw) == 0 || mw[0] != "web") {
			t.Fatalf("want web first, got %v for %s", mw, c)
		}
	}
}

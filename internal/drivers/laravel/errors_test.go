package laravel

import (
	"errors"
	"strings"
	"testing"
)

func TestArtisanFailedError_Error(t *testing.T) {
	e := &ArtisanFailedError{ExitCode: 2, Stderr: "boom"}
	msg := e.Error()
	if !strings.Contains(msg, "exited with code 2") || !strings.Contains(msg, "boom") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestArtisanFailedError_StderrTruncated(t *testing.T) {
	long := strings.Repeat("x", 500)
	e := &ArtisanFailedError{ExitCode: 1, Stderr: long}
	msg := e.Error()
	if !strings.HasSuffix(msg, "...") {
		t.Fatalf("want truncated suffix, got: %s", msg)
	}
	if len(msg) > 320 {
		t.Fatalf("message too long: %d", len(msg))
	}
}

func TestTruncate_NoChangeWhenShort(t *testing.T) {
	got := truncate("hi", 10)
	if got != "hi" {
		t.Fatalf("want hi, got %s", got)
	}
}

func TestTruncate_AddsEllipsis(t *testing.T) {
	got := truncate("abcdef", 3)
	if got != "abc..." {
		t.Fatalf("want abc..., got %s", got)
	}
}

func TestSentinelErrors_Identity(t *testing.T) {
	cases := []error{
		ErrNotLaravel,
		ErrPHPNotFound,
		ErrArtisanMissing,
		ErrInvalidJSON,
		ErrNoRoutes,
		ErrNoFormRequest,
		ErrControllerNotFound,
		ErrRulesNotFound,
		ErrNoInlineValidation,
	}
	for _, e := range cases {
		if !errors.Is(e, e) {
			t.Fatalf("expected errors.Is identity for %v", e)
		}
	}
}

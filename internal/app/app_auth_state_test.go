package app

import "testing"

func TestPreviewToken_ShortPassesThrough(t *testing.T) {
	if got := previewToken("abc"); got != "abc" {
		t.Fatalf("got %q", got)
	}
	if got := previewToken("abcdefghijkl"); got != "abcdefghijkl" {
		t.Fatalf("got %q", got)
	}
}

func TestPreviewToken_LongIsTruncatedWithEllipsis(t *testing.T) {
	in := "abcdef1234567890ZZZZ"
	got := previewToken(in)
	if got != "abcdef…ZZZZ" {
		t.Fatalf("got %q", got)
	}
}

func TestPreviewToken_EmptyReturnsEmpty(t *testing.T) {
	if got := previewToken(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

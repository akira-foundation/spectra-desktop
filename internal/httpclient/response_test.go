package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponse_JSONBodyDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"n":42,"s":"hi"}`)
	}))
	t.Cleanup(srv.Close)

	resp, err := New().Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	var got struct {
		N int    `json:"n"`
		S string `json:"s"`
	}
	if err := json.Unmarshal([]byte(resp.Body), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.N != 42 || got.S != "hi" {
		t.Errorf("decoded = %+v", got)
	}
}

func TestResponse_PlainTextBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "hello world")
	}))
	t.Cleanup(srv.Close)

	resp, err := New().Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.Body != "hello world" {
		t.Errorf("Body = %q", resp.Body)
	}
	if got := resp.Headers["Content-Type"]; len(got) == 0 || got[0] != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %v", got)
	}
}

func TestResponse_BinaryBodyPreserved(t *testing.T) {
	payload := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0x7F}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	resp, err := New().Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.SizeBytes != len(payload) {
		t.Errorf("SizeBytes = %d, want %d", resp.SizeBytes, len(payload))
	}
	if []byte(resp.Body)[3] != 0xFF {
		t.Errorf("byte[3] = %#x, want 0xFF", []byte(resp.Body)[3])
	}
}

func TestResponse_HeadersCloned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-One", "1")
		w.Header().Add("X-Multi", "a")
		w.Header().Add("X-Multi", "b")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	resp, err := New().Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if one := resp.Headers["X-One"]; len(one) == 0 || one[0] != "1" {
		t.Errorf("X-One = %v", one)
	}
	multi := resp.Headers["X-Multi"]
	if len(multi) != 2 || multi[0] != "a" || multi[1] != "b" {
		t.Errorf("X-Multi = %v", multi)
	}
}

func TestResponse_StatusTextPropagated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	resp, err := New().Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.Status != http.StatusNotFound {
		t.Errorf("Status = %d", resp.Status)
	}
	if resp.StatusText == "" {
		t.Error("StatusText empty")
	}
}

func TestResponse_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	resp, err := New().Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.Status != http.StatusNoContent {
		t.Errorf("Status = %d", resp.Status)
	}
	if resp.Body != "" {
		t.Errorf("Body = %q, want empty", resp.Body)
	}
	if resp.SizeBytes != 0 {
		t.Errorf("SizeBytes = %d, want 0", resp.SizeBytes)
	}
}

func TestResponse_MarshalJSONRoundtrip(t *testing.T) {
	in := &Response{
		Status:     200,
		StatusText: "200 OK",
		Headers:    map[string][]string{"X-A": {"1"}},
		Body:       `{"k":"v"}`,
		DurationMs: 12,
		SizeBytes:  9,
		Timeline:   &Timeline{DNSMs: 1, ConnectMs: 2, TLSMs: 3, TTFBMs: 4, DownloadMs: 5},
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out Response
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Status != in.Status || out.Body != in.Body || out.DurationMs != in.DurationMs {
		t.Errorf("roundtrip mismatch: %+v", out)
	}
	if out.Timeline == nil || out.Timeline.TTFBMs != 4 {
		t.Errorf("Timeline lost: %+v", out.Timeline)
	}
}

func TestResponse_TimelineOmittedWhenAllZero(t *testing.T) {
	r := &Response{Status: 200}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if got := string(raw); contains(got, `"timeline"`) {
		t.Errorf("expected omitempty Timeline, got %s", got)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

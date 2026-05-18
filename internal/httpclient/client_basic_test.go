package httpclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_Send_GetOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	t.Cleanup(srv.Close)

	c := New()
	resp, err := c.Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Errorf("Status = %d, want 200", resp.Status)
	}
	if resp.Body != `{"ok":true}` {
		t.Errorf("Body = %q", resp.Body)
	}
	if resp.SizeBytes != len(resp.Body) {
		t.Errorf("SizeBytes = %d, want %d", resp.SizeBytes, len(resp.Body))
	}
	if got := resp.Headers["Content-Type"]; len(got) == 0 || got[0] != "application/json" {
		t.Errorf("Content-Type header = %v", got)
	}
}

func TestClient_Send_PostJSONBody(t *testing.T) {
	var got struct {
		method      string
		contentType string
		body        string
		custom      string
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.contentType = r.Header.Get("Content-Type")
		got.custom = r.Header.Get("X-Custom")
		b, _ := io.ReadAll(r.Body)
		got.body = string(b)
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, `{"created":1}`)
	}))
	t.Cleanup(srv.Close)

	c := New()
	resp, err := c.Send(context.Background(), Request{
		Method: "post",
		URL:    srv.URL,
		Body:   `{"name":"x"}`,
		Headers: map[string]string{
			"X-Custom":        "yes",
			"Accept-Encoding": "identity",
			"":                "should-be-ignored",
		},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.Status != http.StatusCreated {
		t.Errorf("Status = %d, want 201", resp.Status)
	}
	if got.method != http.MethodPost {
		t.Errorf("method = %q, want POST (upper)", got.method)
	}
	if got.contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json default", got.contentType)
	}
	if got.body != `{"name":"x"}` {
		t.Errorf("body = %q", got.body)
	}
	if got.custom != "yes" {
		t.Errorf("X-Custom = %q", got.custom)
	}
}

func TestClient_Send_ContentTypeRespectedWhenSet(t *testing.T) {
	var ct string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := New()
	_, err := c.Send(context.Background(), Request{
		Method:  http.MethodPost,
		URL:     srv.URL,
		Body:    "name=x",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if ct != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want caller-provided", ct)
	}
}

func TestClient_Send_QueryStringPassthrough(t *testing.T) {
	var rawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := New()
	_, err := c.Send(context.Background(), Request{URL: srv.URL + "/x?a=1&b=two"})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if rawQuery != "a=1&b=two" {
		t.Errorf("RawQuery = %q", rawQuery)
	}
}

func TestClient_Send_GetBodyIgnored(t *testing.T) {
	var bodyLen int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		bodyLen = len(b)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := New()
	_, err := c.Send(context.Background(), Request{
		Method: http.MethodGet,
		URL:    srv.URL,
		Body:   "should be dropped",
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if bodyLen != 0 {
		t.Errorf("server saw body of %d bytes on GET", bodyLen)
	}
}

func TestClient_Send_CookieAttached(t *testing.T) {
	var cookieValue string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ck, err := r.Cookie("sid"); err == nil {
			cookieValue = ck.Value
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c := New()
	_, err := c.Send(context.Background(), Request{
		URL:     srv.URL,
		Cookies: []http.Cookie{{Name: "sid", Value: "abc123"}},
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if cookieValue != "abc123" {
		t.Errorf("cookie = %q, want abc123", cookieValue)
	}
}

func TestClient_Send_NonSuccessReturnedAsResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = io.WriteString(w, `nope`)
	}))
	t.Cleanup(srv.Close)

	c := New()
	resp, err := c.Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send returned error for 4xx: %v", err)
	}
	if resp.Status != http.StatusTeapot {
		t.Errorf("Status = %d, want 418", resp.Status)
	}
	if resp.Body != "nope" {
		t.Errorf("Body = %q", resp.Body)
	}
}

func TestClient_Send_FollowsRedirect(t *testing.T) {
	final := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `landed`)
	}))
	t.Cleanup(final.Close)

	redir := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusFound)
	}))
	t.Cleanup(redir.Close)

	c := New()
	resp, err := c.Send(context.Background(), Request{URL: redir.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Errorf("Status = %d, want 200 after redirect", resp.Status)
	}
	if resp.Body != "landed" {
		t.Errorf("Body = %q, want body of final server", resp.Body)
	}
}

func TestClient_Send_CapturesDurationAndSize(t *testing.T) {
	payload := strings.Repeat("a", 256)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, payload)
	}))
	t.Cleanup(srv.Close)

	c := New()
	resp, err := c.Send(context.Background(), Request{URL: srv.URL})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if resp.SizeBytes != len(payload) {
		t.Errorf("SizeBytes = %d, want %d", resp.SizeBytes, len(payload))
	}
	if resp.DurationMs < 0 {
		t.Errorf("DurationMs negative: %d", resp.DurationMs)
	}
}

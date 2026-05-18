package httpclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Send_TimeoutClassified(t *testing.T) {
	block := make(chan struct{})
	t.Cleanup(func() { close(block) })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-block:
		}
	}))
	t.Cleanup(srv.Close)

	c := New()
	_, err := c.Send(context.Background(), Request{
		URL:     srv.URL,
		Timeout: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("err = %v, want ErrTimeout", err)
	}
	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Errorf("expected *RequestError, got %T", err)
	}
}

func TestClient_Send_InvalidURL(t *testing.T) {
	c := New()
	tests := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"not-a-url", "::::not a url"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.Send(context.Background(), Request{URL: tt.url})
			if err == nil {
				t.Fatal("expected error")
			}
			if !errors.Is(err, ErrInvalidURL) {
				t.Errorf("err = %v, want ErrInvalidURL", err)
			}
		})
	}
}

func TestClient_Send_ConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := srv.URL
	srv.Close()

	c := New()
	_, err := c.Send(context.Background(), Request{
		URL:     addr,
		Timeout: 2 * time.Second,
	})
	if err == nil {
		t.Fatal("expected error connecting to closed server")
	}
	if !errors.Is(err, ErrConnectionRefused) && !errors.Is(err, ErrTimeout) {
		t.Errorf("err = %v, want ErrConnectionRefused or ErrTimeout", err)
	}
}

func TestClient_Send_ContextCanceled(t *testing.T) {
	block := make(chan struct{})
	t.Cleanup(func() { close(block) })

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-block:
		}
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	c := New()
	_, err := c.Send(ctx, Request{URL: srv.URL, Timeout: 5 * time.Second})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("err = %v, want ErrTimeout (covers cancel)", err)
	}
}

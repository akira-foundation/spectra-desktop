package httpclient

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) Send(ctx context.Context, req Request) (*Response, error) {
	if strings.TrimSpace(req.URL) == "" {
		return nil, &RequestError{Kind: ErrInvalidURL, Message: "url is empty"}
	}
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return nil, &RequestError{Kind: ErrInvalidURL, Message: err.Error()}
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}

	var body io.Reader
	if req.Body != "" && allowsBody(method) {
		body = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(reqCtx, method, req.URL, body)
	if err != nil {
		return nil, &RequestError{Kind: ErrInvalidURL, Message: err.Error()}
	}

	for k, v := range req.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		// strip Accept-Encoding so Go's http.Transport handles
		// gzip auto-decompression (it disables when caller sets it explicitly).
		if strings.EqualFold(k, "Accept-Encoding") {
			continue
		}
		httpReq.Header.Set(k, v)
	}
	for i := range req.Cookies {
		httpReq.AddCookie(&req.Cookies[i])
	}
	if body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if httpReq.Header.Get("Accept") == "" {
		httpReq.Header.Set("Accept", "application/json")
	}

	var (
		dnsStart, dnsDone         time.Time
		connectStart, connectDone time.Time
		tlsStart, tlsDone         time.Time
		gotFirstByte              time.Time
	)
	trace := &httptrace.ClientTrace{
		DNSStart:             func(httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:              func(httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart:         func(string, string) { connectStart = time.Now() },
		ConnectDone:          func(string, string, error) { connectDone = time.Now() },
		TLSHandshakeStart:    func() { tlsStart = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { tlsDone = time.Now() },
		GotFirstResponseByte: func() { gotFirstByte = time.Now() },
	}
	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if classified := classifyError(err); !errors.Is(classified, err) {
			return nil, classified
		}
		return nil, classifyError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	end := time.Now()
	duration := end.Sub(start)
	if err != nil {
		return nil, &RequestError{Kind: err, Message: "read body"}
	}

	timeline := buildTimeline(start, dnsStart, dnsDone, connectStart, connectDone, tlsStart, tlsDone, gotFirstByte, end)

	out := &Response{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    resp.Header.Clone(),
		Body:       string(bodyBytes),
		DurationMs: duration.Milliseconds(),
		SizeBytes:  len(bodyBytes),
		Timeline:   timeline,
	}
	return out, nil
}

func buildTimeline(start, dnsStart, dnsDone, connStart, connDone, tlsStart, tlsDone, ttfb, end time.Time) *Timeline {
	t := &Timeline{}
	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		t.DNSMs = dnsDone.Sub(dnsStart).Milliseconds()
	}
	if !connStart.IsZero() && !connDone.IsZero() {
		t.ConnectMs = connDone.Sub(connStart).Milliseconds()
	}
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		t.TLSMs = tlsDone.Sub(tlsStart).Milliseconds()
	}
	if !ttfb.IsZero() {
		var ref time.Time
		switch {
		case !tlsDone.IsZero():
			ref = tlsDone
		case !connDone.IsZero():
			ref = connDone
		default:
			ref = start
		}
		t.TTFBMs = ttfb.Sub(ref).Milliseconds()
		t.DownloadMs = end.Sub(ttfb).Milliseconds()
	}
	if t.DNSMs == 0 && t.ConnectMs == 0 && t.TLSMs == 0 && t.TTFBMs == 0 && t.DownloadMs == 0 {
		return nil
	}
	return t
}

func allowsBody(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions:
		return false
	default:
		return true
	}
}

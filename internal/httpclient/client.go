package httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
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

	start := time.Now()
	resp, err := c.httpClient.Do(httpReq)
	duration := time.Since(start)
	if err != nil {
		if classified := classifyError(err); !errors.Is(classified, err) {
			return nil, classified
		}
		return nil, classifyError(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &RequestError{Kind: err, Message: "read body"}
	}

	out := &Response{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    resp.Header.Clone(),
		Body:       string(bodyBytes),
		DurationMs: duration.Milliseconds(),
		SizeBytes:  len(bodyBytes),
	}
	return out, nil
}

func allowsBody(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions:
		return false
	default:
		return true
	}
}

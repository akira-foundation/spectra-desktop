package httpclient

import (
	"net/http"
	"time"
)

type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Cookies []http.Cookie     `json:"-"`
	Timeout time.Duration     `json:"-"`
}

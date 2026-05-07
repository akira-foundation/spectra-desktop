package httpclient

import "errors"

var ErrNotImplemented = errors.New("httpclient not implemented")

type Client struct{}

type Request struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

func New() *Client {
	return &Client{}
}

func (c *Client) Send(_ Request) (*Response, error) {
	return nil, ErrNotImplemented
}

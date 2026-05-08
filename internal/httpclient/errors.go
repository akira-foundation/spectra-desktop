package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"syscall"
)

var (
	ErrInvalidURL        = errors.New("invalid url")
	ErrConnectionRefused = errors.New("connection refused")
	ErrTimeout           = errors.New("request timed out")
	ErrDNS               = errors.New("dns lookup failed")
	ErrTLS               = errors.New("tls handshake failed")
	ErrEmptyBody         = errors.New("empty response body")
)

type RequestError struct {
	Kind    error
	Message string
}

func (e *RequestError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Kind.Error(), e.Message)
	}
	return e.Kind.Error()
}

func (e *RequestError) Unwrap() error {
	return e.Kind
}

func classifyError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return &RequestError{Kind: ErrTimeout, Message: err.Error()}
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return &RequestError{Kind: ErrTimeout, Message: urlErr.Error()}
		}
		inner := urlErr.Err
		if classified := classifyNetError(inner); classified != nil {
			return classified
		}
		if strings.Contains(strings.ToLower(urlErr.Error()), "tls") {
			return &RequestError{Kind: ErrTLS, Message: urlErr.Error()}
		}
	}
	return err
}

func classifyNetError(err error) error {
	if err == nil {
		return nil
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return &RequestError{Kind: ErrDNS, Message: dnsErr.Error()}
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if errors.Is(opErr.Err, syscall.ECONNREFUSED) {
			return &RequestError{Kind: ErrConnectionRefused, Message: opErr.Error()}
		}
	}
	return nil
}

package app

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func toHTTPHeader(in map[string][]string) http.Header {
	h := http.Header{}
	for k, vs := range in {
		for _, v := range vs {
			h.Add(k, v)
		}
	}
	return h
}

func joinURL(base, path string) (string, error) {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	path = strings.TrimSpace(path)
	if base == "" {
		return "", fmt.Errorf("empty base url")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	if path == "" {
		return u.String(), nil
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	return u.String(), nil
}

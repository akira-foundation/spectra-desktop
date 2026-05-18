package httpclient

import (
	"encoding/json"
	"fmt"
	"strings"
)

type HAREntry struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	BaseURL   string            `json:"baseURL,omitempty"`
	Path      string            `json:"path,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body,omitempty"`
	Query     map[string]string `json:"query,omitempty"`
	Status    int               `json:"status,omitempty"`
	Size      int               `json:"size,omitempty"`
	StartedAt string            `json:"startedAt,omitempty"`
}

func ParseHAR(raw string) ([]HAREntry, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, fmt.Errorf("empty HAR")
	}
	var doc struct {
		Log struct {
			Entries []struct {
				StartedAt string `json:"startedDateTime"`
				Request   struct {
					Method  string `json:"method"`
					URL     string `json:"url"`
					Headers []struct {
						Name  string `json:"name"`
						Value string `json:"value"`
					} `json:"headers"`
					QueryString []struct {
						Name  string `json:"name"`
						Value string `json:"value"`
					} `json:"queryString"`
					PostData *struct {
						MimeType string `json:"mimeType"`
						Text     string `json:"text"`
					} `json:"postData"`
				} `json:"request"`
				Response struct {
					Status  int `json:"status"`
					Content struct {
						Size int `json:"size"`
					} `json:"content"`
				} `json:"response"`
			} `json:"entries"`
		} `json:"log"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return nil, fmt.Errorf("invalid HAR JSON: %w", err)
	}
	out := make([]HAREntry, 0, len(doc.Log.Entries))
	for _, e := range doc.Log.Entries {
		entry := HAREntry{
			Method:    e.Request.Method,
			URL:       e.Request.URL,
			Status:    e.Response.Status,
			Size:      e.Response.Content.Size,
			StartedAt: e.StartedAt,
			Headers:   map[string]string{},
			Query:     map[string]string{},
		}
		if e.Request.PostData != nil {
			entry.Body = e.Request.PostData.Text
		}
		for _, h := range e.Request.Headers {
			if strings.HasPrefix(h.Name, ":") {
				continue
			}
			entry.Headers[h.Name] = h.Value
		}
		for _, q := range e.Request.QueryString {
			entry.Query[q.Name] = q.Value
		}
		if idx := strings.Index(e.Request.URL, "://"); idx > 0 {
			rest := e.Request.URL[idx+3:]
			slash := strings.Index(rest, "/")
			if slash > 0 {
				entry.BaseURL = e.Request.URL[:idx+3] + rest[:slash]
				path := rest[slash:]
				if q := strings.Index(path, "?"); q >= 0 {
					path = path[:q]
				}
				entry.Path = path
			}
		}
		out = append(out, entry)
	}
	return out, nil
}

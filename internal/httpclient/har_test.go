package httpclient

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseHAR_EmptyInput(t *testing.T) {
	tests := []string{"", "   ", "\n\t "}
	for _, in := range tests {
		_, err := ParseHAR(in)
		if err == nil {
			t.Errorf("ParseHAR(%q) = nil, want error", in)
		}
	}
}

func TestParseHAR_InvalidJSON(t *testing.T) {
	_, err := ParseHAR(`{not-json`)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid HAR JSON") {
		t.Errorf("err = %v, want containing 'invalid HAR JSON'", err)
	}
}

func TestParseHAR_EmptyEntries(t *testing.T) {
	entries, err := ParseHAR(`{"log":{"entries":[]}}`)
	if err != nil {
		t.Fatalf("ParseHAR: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %d, want 0", len(entries))
	}
}

func TestParseHAR_FullEntryFields(t *testing.T) {
	raw := `{
      "log": {
        "entries": [
          {
            "startedDateTime": "2026-01-01T00:00:00Z",
            "request": {
              "method": "POST",
              "url": "https://api.example.com/v1/users?page=2&size=10",
              "headers": [
                {"name": "Content-Type", "value": "application/json"},
                {"name": "Authorization", "value": "Bearer abc"},
                {"name": ":authority", "value": "api.example.com"}
              ],
              "queryString": [
                {"name": "page", "value": "2"},
                {"name": "size", "value": "10"}
              ],
              "postData": {
                "mimeType": "application/json",
                "text": "{\"name\":\"x\"}"
              }
            },
            "response": {
              "status": 201,
              "content": {"size": 128}
            }
          }
        ]
      }
    }`
	entries, err := ParseHAR(raw)
	if err != nil {
		t.Fatalf("ParseHAR: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	e := entries[0]
	if e.Method != "POST" {
		t.Errorf("Method = %q", e.Method)
	}
	if e.URL != "https://api.example.com/v1/users?page=2&size=10" {
		t.Errorf("URL = %q", e.URL)
	}
	if e.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL = %q", e.BaseURL)
	}
	if e.Path != "/v1/users" {
		t.Errorf("Path = %q", e.Path)
	}
	if e.Body != `{"name":"x"}` {
		t.Errorf("Body = %q", e.Body)
	}
	if e.Status != 201 {
		t.Errorf("Status = %d", e.Status)
	}
	if e.Size != 128 {
		t.Errorf("Size = %d", e.Size)
	}
	if e.StartedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("StartedAt = %q", e.StartedAt)
	}
	if e.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type header = %q", e.Headers["Content-Type"])
	}
	if e.Headers["Authorization"] != "Bearer abc" {
		t.Errorf("Authorization header = %q", e.Headers["Authorization"])
	}
	if _, ok := e.Headers[":authority"]; ok {
		t.Error("pseudo-header :authority should be stripped")
	}
	if e.Query["page"] != "2" || e.Query["size"] != "10" {
		t.Errorf("Query = %+v", e.Query)
	}
}

func TestParseHAR_NoPathWhenBareHost(t *testing.T) {
	raw := `{"log":{"entries":[{
      "startedDateTime":"",
      "request":{"method":"GET","url":"https://example.com","headers":[],"queryString":[]},
      "response":{"status":200,"content":{"size":0}}
    }]}}`
	entries, err := ParseHAR(raw)
	if err != nil {
		t.Fatalf("ParseHAR: %v", err)
	}
	if entries[0].Path != "" {
		t.Errorf("Path = %q, want empty", entries[0].Path)
	}
	if entries[0].BaseURL != "" {
		t.Errorf("BaseURL = %q, want empty when no slash found", entries[0].BaseURL)
	}
}

func TestParseHAR_PathStripsQuery(t *testing.T) {
	raw := `{"log":{"entries":[{
      "startedDateTime":"",
      "request":{"method":"GET","url":"http://h/x/y?q=1&r=2","headers":[],"queryString":[]},
      "response":{"status":200,"content":{"size":0}}
    }]}}`
	entries, err := ParseHAR(raw)
	if err != nil {
		t.Fatalf("ParseHAR: %v", err)
	}
	if entries[0].Path != "/x/y" {
		t.Errorf("Path = %q, want /x/y", entries[0].Path)
	}
	if entries[0].BaseURL != "http://h" {
		t.Errorf("BaseURL = %q, want http://h", entries[0].BaseURL)
	}
}

func TestParseHAR_NilPostDataLeavesBodyEmpty(t *testing.T) {
	raw := `{"log":{"entries":[{
      "startedDateTime":"",
      "request":{"method":"GET","url":"http://h/p","headers":[],"queryString":[]},
      "response":{"status":204,"content":{"size":0}}
    }]}}`
	entries, err := ParseHAR(raw)
	if err != nil {
		t.Fatalf("ParseHAR: %v", err)
	}
	if entries[0].Body != "" {
		t.Errorf("Body = %q, want empty", entries[0].Body)
	}
}

func TestHAREntry_MarshalRoundtrip(t *testing.T) {
	in := HAREntry{
		Method:    "POST",
		URL:       "https://api.example.com/v1/items",
		BaseURL:   "https://api.example.com",
		Path:      "/v1/items",
		Headers:   map[string]string{"X-A": "1", "X-B": "2"},
		Body:      `{"n":1}`,
		Query:     map[string]string{"k": "v"},
		Status:    201,
		Size:      42,
		StartedAt: "2026-05-17T00:00:00Z",
	}
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var out HAREntry
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Method != in.Method || out.URL != in.URL || out.BaseURL != in.BaseURL {
		t.Errorf("scalar mismatch: %+v", out)
	}
	if out.Path != in.Path || out.Body != in.Body || out.Status != in.Status || out.Size != in.Size {
		t.Errorf("scalar mismatch: %+v", out)
	}
	if out.StartedAt != in.StartedAt {
		t.Errorf("StartedAt = %q", out.StartedAt)
	}
	if len(out.Headers) != len(in.Headers) {
		t.Errorf("Headers len = %d, want %d", len(out.Headers), len(in.Headers))
	}
	for k, v := range in.Headers {
		if out.Headers[k] != v {
			t.Errorf("Headers[%q] = %q, want %q", k, out.Headers[k], v)
		}
	}
	for k, v := range in.Query {
		if out.Query[k] != v {
			t.Errorf("Query[%q] = %q, want %q", k, out.Query[k], v)
		}
	}
}

func TestParseHAR_MultipleEntries(t *testing.T) {
	raw := `{"log":{"entries":[
      {"startedDateTime":"t1","request":{"method":"GET","url":"http://a/1","headers":[],"queryString":[]},"response":{"status":200,"content":{"size":1}}},
      {"startedDateTime":"t2","request":{"method":"PUT","url":"http://a/2","headers":[],"queryString":[]},"response":{"status":204,"content":{"size":0}}}
    ]}}`
	entries, err := ParseHAR(raw)
	if err != nil {
		t.Fatalf("ParseHAR: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0].Method != "GET" || entries[1].Method != "PUT" {
		t.Errorf("methods = %q, %q", entries[0].Method, entries[1].Method)
	}
}

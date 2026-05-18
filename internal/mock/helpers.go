package mock

import (
	"encoding/json"
	"net/http"
	"strings"
)

func withJSONContentType(h map[string]string) map[string]string {
	if h == nil {
		h = map[string]string{}
	}
	hasContentType := false
	for k := range h {
		if strings.EqualFold(k, "Content-Type") {
			hasContentType = true
			break
		}
	}
	if !hasContentType {
		h["Content-Type"] = "application/json"
	}
	return h
}

func parseStoredResponseHeaders(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	var multi map[string][]string
	if err := json.Unmarshal([]byte(raw), &multi); err == nil {
		out := make(map[string]string, len(multi))
		for k, v := range multi {
			if len(v) > 0 {
				out[k] = v[0]
			}
		}
		return out
	}
	var single map[string]string
	if err := json.Unmarshal([]byte(raw), &single); err == nil {
		return single
	}
	return nil
}

func statusOrDefault(value, fallback int) int {
	if value > 0 {
		return value
	}
	if fallback > 0 {
		return fallback
	}
	return http.StatusOK
}

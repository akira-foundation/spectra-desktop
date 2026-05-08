package laravel

import (
	"encoding/json"
	"strconv"
	"strings"

	"spectra-desktop/internal/core"
)

const driverConfidence = 0.95

func normalize(raws []rawRoute) []core.Endpoint {
	endpoints := make([]core.Endpoint, 0, len(raws))
	seen := make(map[string]int)
	for _, raw := range raws {
		methods := splitMethods(raw.Method)
		path := normalizePath(raw.URI)
		middleware := decodeMiddleware(raw.Middleware)
		handler := strings.TrimSpace(raw.Action)
		name := strings.TrimSpace(raw.Name)

		for _, method := range methods {
			key := method + ":" + path
			id := key
			if name != "" {
				id = key + "#" + name
			}
			if count, ok := seen[id]; ok {
				seen[id] = count + 1
				id = id + "@" + strconv.Itoa(count+1)
			} else {
				seen[id] = 1
			}
			endpoints = append(endpoints, core.Endpoint{
				ID:         id,
				Method:     core.HTTPMethod(method),
				Path:       path,
				Name:       name,
				Handler:    handler,
				Middleware: middleware,
				Framework:  DriverName,
				Confidence: driverConfidence,
			})
		}
	}
	return endpoints
}

func splitMethods(raw string) []string {
	parts := strings.Split(raw, "|")
	cleaned := make([]string, 0, len(parts))
	hasGet := false
	for _, p := range parts {
		m := strings.ToUpper(strings.TrimSpace(p))
		if m == "" {
			continue
		}
		if m == "GET" {
			hasGet = true
		}
		cleaned = append(cleaned, m)
	}
	if !hasGet {
		return cleaned
	}
	out := make([]string, 0, len(cleaned))
	for _, m := range cleaned {
		if m == "HEAD" {
			continue
		}
		out = append(out, m)
	}
	return out
}

func normalizePath(uri string) string {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return "/"
	}
	if !strings.HasPrefix(uri, "/") {
		uri = "/" + uri
	}
	return uri
}

func decodeMiddleware(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var asSlice []string
	if err := json.Unmarshal(raw, &asSlice); err == nil {
		return cleanStrings(asSlice)
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		if asString == "" {
			return nil
		}
		return cleanStrings(strings.Split(asString, ","))
	}
	return nil
}

func cleanStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

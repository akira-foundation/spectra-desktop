package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func extractCaptureValue(source, path string, body any, headers http.Header) (string, bool) {
	switch source {
	case "header":
		if v := headers.Get(path); v != "" {
			return v, true
		}
		return "", false
	case "body", "":
		v, ok := lookupJSONPath(body, path)
		if !ok {
			return "", false
		}
		return formatCapturedValue(v), true
	}
	return "", false
}

func formatCapturedValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	default:
		buf, err := json.Marshal(x)
		if err != nil {
			return fmt.Sprintf("%v", x)
		}
		return string(buf)
	}
}

func lookupJSONPath(root any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" {
		return root, true
	}
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")
	cur := root
	tokens := splitJSONPath(path)
	for _, p := range tokens {
		switch v := cur.(type) {
		case map[string]any:
			next, ok := v[p]
			if !ok {
				return nil, false
			}
			cur = next
		case []any:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, false
			}
			cur = v[idx]
		default:
			return nil, false
		}
	}
	return cur, true
}

func splitJSONPath(path string) []string {
	if path == "" {
		return nil
	}
	out := []string{}
	cur := []byte{}
	flush := func() {
		if len(cur) > 0 {
			out = append(out, strings.Trim(string(cur), `"'`))
			cur = cur[:0]
		}
	}
	for i := 0; i < len(path); i++ {
		c := path[i]
		switch c {
		case '.', '[':
			flush()
		case ']':
			flush()
		default:
			cur = append(cur, c)
		}
	}
	flush()
	return out
}

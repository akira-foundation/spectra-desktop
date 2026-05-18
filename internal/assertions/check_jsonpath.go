package assertions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func checkJSONPath(body any, t Test) (bool, string) {
	if body == nil && t.Op != "not_exists" {
		return false, "response body is not JSON"
	}
	value, ok := lookupPath(body, t.JSONPath)
	op := t.Op
	if op == "" {
		op = "exists"
	}
	switch op {
	case "exists":
		if ok {
			return true, ""
		}
		return false, "path not found"
	case "not_exists":
		if !ok {
			return true, ""
		}
		return false, fmt.Sprintf("path resolved to %v", value)
	case "equals":
		if !ok {
			return false, "path not found"
		}
		if compareEquals(value, t.Expected) {
			return true, ""
		}
		return false, fmt.Sprintf("got %v, expected %s", value, t.Expected)
	case "matches":
		s, isStr := value.(string)
		if !isStr {
			return false, "value is not a string"
		}
		re, err := regexp.Compile(t.Expected)
		if err != nil {
			return false, "invalid regex: " + err.Error()
		}
		if re.MatchString(s) {
			return true, ""
		}
		return false, fmt.Sprintf("%q does not match /%s/", s, t.Expected)
	case "type":
		actualType := jsonType(value)
		if actualType == strings.ToLower(t.Expected) {
			return true, ""
		}
		return false, fmt.Sprintf("type %s, expected %s", actualType, t.Expected)
	case "min_length":
		want, err := strconv.Atoi(strings.TrimSpace(t.Expected))
		if err != nil {
			return false, "invalid min_length"
		}
		length := lengthOf(value)
		if length < 0 {
			return false, "value has no length"
		}
		if length >= want {
			return true, ""
		}
		return false, fmt.Sprintf("length %d, expected ≥ %d", length, want)
	}
	return false, "unknown jsonpath op: " + t.Op
}

func lookupPath(root any, path string) (any, bool) {
	path = strings.TrimSpace(path)
	if path == "" || path == "$" {
		return root, true
	}
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")
	parts := splitPath(path)
	cur := root
	for _, p := range parts {
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

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	out := []string{}
	cur := strings.Builder{}
	for i := 0; i < len(path); i++ {
		c := path[i]
		switch c {
		case '.':
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		case '[':
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		case ']':
			if cur.Len() > 0 {
				key := strings.Trim(cur.String(), `"'`)
				out = append(out, key)
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

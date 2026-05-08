package assertions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Test struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	JSONPath string `json:"jsonPath,omitempty"`
	Op       string `json:"op,omitempty"`
	Expected string `json:"expected,omitempty"`
}

type Result struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Pass    bool   `json:"pass"`
	Message string `json:"message,omitempty"`
}

type ResponseSnapshot struct {
	Status     int
	Headers    http.Header
	Body       string
	DurationMs int
}

func Run(tests []Test, resp ResponseSnapshot) []Result {
	results := make([]Result, 0, len(tests))
	var bodyValue any
	if resp.Body != "" {
		_ = json.Unmarshal([]byte(resp.Body), &bodyValue)
	}
	for _, t := range tests {
		results = append(results, evaluate(t, resp, bodyValue))
	}
	return results
}

func evaluate(t Test, resp ResponseSnapshot, body any) Result {
	r := Result{ID: t.ID, Name: deriveName(t), Kind: t.Kind}
	switch t.Kind {
	case "status":
		r.Pass, r.Message = checkStatus(resp.Status, t.Expected)
	case "max_duration":
		r.Pass, r.Message = checkMaxDuration(resp.DurationMs, t.Expected)
	case "header":
		r.Pass, r.Message = checkHeader(resp.Headers, t)
	case "jsonpath":
		r.Pass, r.Message = checkJSONPath(body, t)
	case "body":
		r.Pass, r.Message = checkBody(resp.Body, t)
	default:
		r.Pass = false
		r.Message = "unknown test kind: " + t.Kind
	}
	return r
}

func deriveName(t Test) string {
	if t.Name != "" {
		return t.Name
	}
	switch t.Kind {
	case "status":
		return "Status " + t.Expected
	case "max_duration":
		return "Max " + t.Expected + "ms"
	case "header":
		return "Header " + t.JSONPath + " " + t.Op + " " + t.Expected
	case "jsonpath":
		return t.JSONPath + " " + t.Op + " " + t.Expected
	case "body":
		return "Body " + t.Op + " " + t.Expected
	}
	return t.Kind
}

func checkStatus(actual int, expected string) (bool, string) {
	expected = strings.TrimSpace(strings.ToLower(expected))
	if expected == "" {
		return false, "missing expected status"
	}
	// supports "2xx", "3xx", "4xx", "5xx"
	if len(expected) == 3 && strings.HasSuffix(expected, "xx") {
		if !unicodeDigit(expected[0]) {
			return false, "invalid status pattern"
		}
		bucket := int(expected[0]-'0') * 100
		if actual >= bucket && actual < bucket+100 {
			return true, ""
		}
		return false, fmt.Sprintf("got %d, expected %s", actual, expected)
	}
	want, err := strconv.Atoi(expected)
	if err != nil {
		return false, "invalid status: " + expected
	}
	if actual == want {
		return true, ""
	}
	return false, fmt.Sprintf("got %d, expected %d", actual, want)
}

func unicodeDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func checkMaxDuration(actual int, expected string) (bool, string) {
	want, err := strconv.Atoi(strings.TrimSpace(expected))
	if err != nil {
		return false, "invalid max_duration: " + expected
	}
	if actual <= want {
		return true, ""
	}
	return false, fmt.Sprintf("took %dms, expected ≤ %dms", actual, want)
}

func checkHeader(headers http.Header, t Test) (bool, string) {
	if t.JSONPath == "" {
		return false, "missing header name"
	}
	value := headers.Get(t.JSONPath)
	switch t.Op {
	case "exists":
		if value != "" {
			return true, ""
		}
		return false, "header not present"
	case "not_exists":
		if value == "" {
			return true, ""
		}
		return false, "header present: " + value
	case "equals":
		if value == t.Expected {
			return true, ""
		}
		return false, fmt.Sprintf("got %q, expected %q", value, t.Expected)
	case "contains":
		if strings.Contains(value, t.Expected) {
			return true, ""
		}
		return false, fmt.Sprintf("%q does not contain %q", value, t.Expected)
	case "matches":
		re, err := regexp.Compile(t.Expected)
		if err != nil {
			return false, "invalid regex: " + err.Error()
		}
		if re.MatchString(value) {
			return true, ""
		}
		return false, fmt.Sprintf("%q does not match /%s/", value, t.Expected)
	}
	return false, "unknown header op: " + t.Op
}

func checkJSONPath(body any, t Test) (bool, string) {
	if body == nil && t.Op != "not_exists" {
		return false, "response body is not JSON"
	}
	value, ok := lookupPath(body, t.JSONPath)
	switch t.Op {
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

func checkBody(body string, t Test) (bool, string) {
	switch t.Op {
	case "contains":
		if strings.Contains(body, t.Expected) {
			return true, ""
		}
		return false, "body does not contain expected"
	case "matches":
		re, err := regexp.Compile(t.Expected)
		if err != nil {
			return false, "invalid regex: " + err.Error()
		}
		if re.MatchString(body) {
			return true, ""
		}
		return false, "body does not match regex"
	}
	return false, "unknown body op: " + t.Op
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

func compareEquals(actual any, expected string) bool {
	expected = strings.TrimSpace(expected)
	switch v := actual.(type) {
	case string:
		return v == expected
	case bool:
		return strconv.FormatBool(v) == strings.ToLower(expected)
	case float64:
		want, err := strconv.ParseFloat(expected, 64)
		if err != nil {
			return false
		}
		return v == want
	case nil:
		return strings.ToLower(expected) == "null"
	default:
		buf, _ := json.Marshal(actual)
		var wantVal any
		if err := json.Unmarshal([]byte(expected), &wantVal); err == nil {
			wantBuf, _ := json.Marshal(wantVal)
			return string(buf) == string(wantBuf)
		}
		return string(buf) == expected
	}
}

func jsonType(v any) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case bool:
		return "boolean"
	case string:
		return "string"
	case float64:
		f := v.(float64)
		if f == float64(int64(f)) {
			return "integer"
		}
		return "number"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return reflect.TypeOf(v).Kind().String()
	}
}

func lengthOf(v any) int {
	switch x := v.(type) {
	case string:
		return len(x)
	case []any:
		return len(x)
	case map[string]any:
		return len(x)
	}
	return -1
}

package assertions

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func checkStatus(actual int, expected string) (bool, string) {
	expected = strings.TrimSpace(strings.ToLower(expected))
	if expected == "" {
		return false, "missing expected status"
	}
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
	op := t.Op
	if op == "" {
		op = "exists"
	}
	switch op {
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

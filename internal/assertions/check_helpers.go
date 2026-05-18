package assertions

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

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

package assertions

import (
	"encoding/json"
	"net/http"
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

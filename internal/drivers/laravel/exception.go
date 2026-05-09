package laravel

import (
	"encoding/json"
	"strings"

	"spectra-desktop/internal/core"
)

// parseLaravelException tries to parse Laravel exception payloads as returned by
// `App\Exceptions\Handler` in JSON mode (`Accept: application/json`).
//
// Typical shape:
//
//	{
//	  "message": "...",
//	  "exception": "Illuminate\\Validation\\ValidationException",
//	  "file": "/var/www/.../Foo.php",
//	  "line": 123,
//	  "trace": [{ "file": "...", "line": 1, "function": "..." }, ...]
//	}
func parseLaravelException(body string, status int) (core.FormattedException, bool) {
	if status < 400 {
		return core.FormattedException{}, false
	}
	trimmed := strings.TrimSpace(body)
	if !strings.HasPrefix(trimmed, "{") {
		return core.FormattedException{}, false
	}
	var raw struct {
		Message   string         `json:"message"`
		Exception string         `json:"exception"`
		File      string         `json:"file"`
		Line      int            `json:"line"`
		Trace     []traceFrame   `json:"trace"`
		Errors    map[string]any `json:"errors,omitempty"`
	}
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		return core.FormattedException{}, false
	}
	if raw.Message == "" && raw.Exception == "" && len(raw.Errors) == 0 {
		return core.FormattedException{}, false
	}
	out := core.FormattedException{
		Message: raw.Message,
		Class:   raw.Exception,
		File:    raw.File,
		Line:    raw.Line,
	}
	for _, t := range raw.Trace {
		out.Trace = append(out.Trace, core.FormattedTraceFrame{
			File:     t.File,
			Line:     t.Line,
			Function: buildFnName(t),
		})
		if len(out.Trace) >= 25 {
			break
		}
	}
	if len(raw.Errors) > 0 {
		out.Extra = map[string]any{"errors": raw.Errors}
	}
	return out, true
}

type traceFrame struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Class    string `json:"class"`
	Type     string `json:"type"`
}

func buildFnName(t traceFrame) string {
	if t.Class == "" {
		return t.Function
	}
	sep := t.Type
	if sep == "" {
		sep = "::"
	}
	return t.Class + sep + t.Function
}

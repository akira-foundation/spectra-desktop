package mock

import (
	"encoding/json"
	"regexp"
	"strings"

	"spectra-desktop/internal/core"
)

type matchResult struct {
	endpoint core.Endpoint
	params   map[string]string
}

var pathParamPattern = regexp.MustCompile(`\{([^}]+)\}`)

func findEndpointMatchingRequest(method, path string, eps []core.Endpoint) *matchResult {
	method = strings.ToUpper(method)
	for _, ep := range eps {
		if !strings.EqualFold(string(ep.Method), method) {
			continue
		}
		params, ok := matchPathTemplate(ep.Path, path)
		if ok {
			return &matchResult{endpoint: ep, params: params}
		}
	}
	return nil
}

func matchPathTemplate(template, actual string) (map[string]string, bool) {
	tParts := splitPathSegments(template)
	aParts := splitPathSegments(actual)
	if len(tParts) != len(aParts) {
		return nil, false
	}
	params := map[string]string{}
	for i, tp := range tParts {
		ap := aParts[i]
		if matches := pathParamPattern.FindStringSubmatch(tp); len(matches) == 2 {
			name := strings.TrimSuffix(matches[1], "?")
			params[name] = ap
			continue
		}
		if !strings.EqualFold(tp, ap) {
			return nil, false
		}
	}
	return params, true
}

func splitPathSegments(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func endpointSchemaFieldNames(ep core.Endpoint) []string {
	if ep.RequestSchema == "" {
		return nil
	}
	var generic struct {
		Fields []struct {
			Name string `json:"name"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(ep.RequestSchema), &generic); err != nil {
		return nil
	}
	out := make([]string, 0, len(generic.Fields))
	for _, f := range generic.Fields {
		if f.Name != "" {
			out = append(out, f.Name)
		}
	}
	return out
}

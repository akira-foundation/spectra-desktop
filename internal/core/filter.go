package core

import "strings"

const (
	FilterModeAuto       = "auto"
	FilterModeMiddleware = "middleware"
	FilterModePrefix     = "prefix"
	FilterModeAll        = "all"
)

type FilterResult struct {
	Mode      string     `json:"mode"`
	Value     string     `json:"value"`
	Endpoints []Endpoint `json:"endpoints"`
}

func ApplyFilter(eps []Endpoint, mode, value string) FilterResult {
	mode = strings.ToLower(strings.TrimSpace(mode))
	value = strings.TrimSpace(value)

	switch mode {
	case FilterModeAll:
		return FilterResult{Mode: FilterModeAll, Value: "", Endpoints: append([]Endpoint(nil), eps...)}
	case FilterModeMiddleware:
		return FilterResult{Mode: mode, Value: value, Endpoints: filterByMiddleware(eps, value)}
	case FilterModePrefix:
		return FilterResult{Mode: mode, Value: value, Endpoints: filterByPrefix(eps, value)}
	default:
		// auto: try middleware "api" first, fallback prefix "/api"
		mw := filterByMiddleware(eps, "api")
		if len(mw) > 0 {
			return FilterResult{Mode: FilterModeMiddleware, Value: "api", Endpoints: mw}
		}
		px := filterByPrefix(eps, "api")
		if len(px) > 0 {
			return FilterResult{Mode: FilterModePrefix, Value: "api", Endpoints: px}
		}
		return FilterResult{Mode: FilterModeAuto, Value: "", Endpoints: nil}
	}
}

func filterByMiddleware(eps []Endpoint, target string) []Endpoint {
	if target == "" {
		return nil
	}
	target = strings.ToLower(target)
	out := make([]Endpoint, 0, len(eps))
	for _, e := range eps {
		for _, m := range e.Middleware {
			if matchesMiddleware(m, target) {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

func matchesMiddleware(actual, target string) bool {
	actual = strings.ToLower(actual)
	if actual == target {
		return true
	}
	if i := strings.LastIndex(actual, "\\"); i >= 0 {
		actual = actual[i+1:]
	}
	if i := strings.Index(actual, ":"); i >= 0 {
		actual = actual[:i]
	}
	return strings.EqualFold(actual, target)
}

func filterByPrefix(eps []Endpoint, prefix string) []Endpoint {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return nil
	}
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix == "" {
		return nil
	}
	prefix = "/" + strings.ToLower(prefix)

	out := make([]Endpoint, 0, len(eps))
	for _, e := range eps {
		path := strings.ToLower(e.Path)
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			out = append(out, e)
		}
	}
	return out
}

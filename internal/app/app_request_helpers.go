package app

import (
	"regexp"
	"strings"
)

func (a *App) resolveEnvVars(projectID string) map[string]string {
	out := map[string]string{}
	if projectID == "" {
		return out
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err == nil && project != nil && project.ActiveEnvironmentID != "" {
		if env, err := a.envs.GetByID(a.ctx, project.ActiveEnvironmentID); err == nil && env != nil {
			for k, v := range env.Vars {
				out[k] = v
			}
		}
	}
	if a.captured != nil {
		a.captured.ensureLoaded(projectID)
		for k, v := range a.captured.values(projectID) {
			out[k] = v
		}
	}
	return out
}

var varPattern = regexp.MustCompile(`\{\{\s*([A-Za-z0-9_.\-]+)\s*\}\}`)

func substituteVars(input string, vars map[string]string) string {
	if input == "" || len(vars) == 0 {
		return input
	}
	return varPattern.ReplaceAllStringFunc(input, func(match string) string {
		groups := varPattern.FindStringSubmatch(match)
		if len(groups) < 2 {
			return match
		}
		key := groups[1]
		if v, ok := vars[key]; ok {
			return v
		}
		return match
	})
}

func substituteHeaderVars(headers map[string]string, vars map[string]string) map[string]string {
	if len(headers) == 0 {
		return headers
	}
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		resolvedKey := substituteVars(k, vars)
		if !isValidHeaderName(resolvedKey) {
			continue
		}
		out[resolvedKey] = substituteVars(v, vars)
	}
	return out
}

func isValidHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", rune(c)):
		default:
			return false
		}
	}
	return true
}

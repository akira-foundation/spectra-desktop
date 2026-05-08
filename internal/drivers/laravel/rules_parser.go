package laravel

import "strings"

func parseRules(raw string) []string {
	parts := strings.Split(raw, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func hasRule(rules []string, name string) bool {
	for _, r := range rules {
		if equalRule(r, name) {
			return true
		}
	}
	return false
}

func equalRule(rule, name string) bool {
	rule = strings.TrimSpace(strings.ToLower(rule))
	name = strings.ToLower(name)
	if rule == name {
		return true
	}
	if strings.HasPrefix(rule, name+":") {
		return true
	}
	return false
}

func inferType(rules []string) string {
	if hasRule(rules, "email") {
		return "email"
	}
	if hasRule(rules, "uuid") {
		return "uuid"
	}
	if hasRule(rules, "url") {
		return "url"
	}
	if hasRule(rules, "integer") {
		return "integer"
	}
	if hasRule(rules, "numeric") {
		return "numeric"
	}
	if hasRule(rules, "boolean") {
		return "boolean"
	}
	if hasRule(rules, "array") {
		return "array"
	}
	if hasRule(rules, "date") || hasRule(rules, "date_format") {
		return "date"
	}
	if hasRule(rules, "file") || hasRule(rules, "image") {
		return "file"
	}
	if hasRule(rules, "json") {
		return "object"
	}
	return "string"
}

func isRequired(rules []string) bool {
	if hasRule(rules, "required") {
		return true
	}
	for _, r := range rules {
		lower := strings.ToLower(strings.TrimSpace(r))
		if strings.HasPrefix(lower, "required_") {
			return true
		}
	}
	return false
}

func hasConfirmed(rules []string) bool {
	return hasRule(rules, "confirmed")
}

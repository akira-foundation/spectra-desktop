package laravel

import (
	"regexp"
	"strings"
)

var pairRegex = regexp.MustCompile(`(?s)['"]([A-Za-z0-9_.\-*]+)['"]\s*=>\s*(\[[^\]]*\]|['"][^'"]*['"])`)

// extractArrayLiteral returns the source between the matching brackets right
// after the given start position. The byte at startsAt-1 must be '['.
func extractArrayLiteral(src string, startsAt int) (string, int, bool) {
	if startsAt <= 0 || startsAt > len(src) {
		return "", -1, false
	}
	if src[startsAt-1] != '[' {
		return "", -1, false
	}
	depth := 1
	i := startsAt
	inString := byte(0)
	escape := false
	for i < len(src) {
		c := src[i]
		if escape {
			escape = false
			i++
			continue
		}
		if inString != 0 {
			if c == '\\' {
				escape = true
				i++
				continue
			}
			if c == inString {
				inString = 0
			}
			i++
			continue
		}
		switch c {
		case '\'', '"':
			inString = c
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return src[startsAt:i], i, true
			}
		}
		i++
	}
	return "", -1, false
}

// extractBraceBlock returns the source between matching `{ ... }` after
// startsAt-1. The byte at startsAt-1 must be '{'.
func extractBraceBlock(src string, startsAt int) (string, int, bool) {
	if startsAt <= 0 || startsAt > len(src) {
		return "", -1, false
	}
	if src[startsAt-1] != '{' {
		return "", -1, false
	}
	depth := 1
	i := startsAt
	inString := byte(0)
	escape := false
	for i < len(src) {
		c := src[i]
		if escape {
			escape = false
			i++
			continue
		}
		if inString != 0 {
			if c == '\\' {
				escape = true
				i++
				continue
			}
			if c == inString {
				inString = 0
			}
			i++
			continue
		}
		switch c {
		case '\'', '"':
			inString = c
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[startsAt:i], i, true
			}
		}
		i++
	}
	return "", -1, false
}

// parseRulePairs walks an array literal body and pulls 'key' => rules pairs.
func parseRulePairs(body string) []InferredField {
	matches := pairRegex.FindAllStringSubmatch(body, -1)
	out := make([]InferredField, 0, len(matches))
	for _, m := range matches {
		if len(m) < 3 {
			continue
		}
		name := strings.TrimSpace(m[1])
		raw := strings.TrimSpace(m[2])
		rules := decodeRulesValue(raw)
		field := buildField(name, rules)
		out = append(out, field)
	}
	return out
}

func decodeRulesValue(raw string) []string {
	if raw == "" {
		return nil
	}
	if raw[0] == '[' {
		inside := strings.TrimSuffix(strings.TrimPrefix(raw, "["), "]")
		return splitArrayStrings(inside)
	}
	if raw[0] == '\'' || raw[0] == '"' {
		quoted := strings.Trim(raw, "'\"")
		return parseRules(quoted)
	}
	return nil
}

func splitArrayStrings(body string) []string {
	parts := strings.Split(body, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "'\"")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func buildField(name string, rules []string) InferredField {
	t := inferType(rules)
	return InferredField{
		Name:     name,
		Type:     t,
		Required: isRequired(rules),
		Rules:    rules,
		Example:  generateExample(name, t, rules),
	}
}

// expandConfirmedFields adds <field>_confirmation entries for fields with the
// confirmed rule when not already present.
func expandConfirmedFields(fields []InferredField) []InferredField {
	existing := make(map[string]bool, len(fields))
	for _, f := range fields {
		existing[f.Name] = true
	}
	expanded := make([]InferredField, 0, len(fields))
	for _, f := range fields {
		expanded = append(expanded, f)
		if !hasConfirmed(f.Rules) {
			continue
		}
		conf := f.Name + "_confirmation"
		if existing[conf] {
			continue
		}
		expanded = append(expanded, InferredField{
			Name:     conf,
			Type:     f.Type,
			Required: f.Required,
			Rules:    []string{"required"},
			Example:  f.Example,
		})
		existing[conf] = true
	}
	return expanded
}

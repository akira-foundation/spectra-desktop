package laravel

import "strings"

func generateExample(name, fieldType string, rules []string) any {
	lowerName := strings.ToLower(name)

	switch fieldType {
	case "email":
		return "user@example.com"
	case "uuid":
		return "00000000-0000-4000-8000-000000000000"
	case "url":
		return "https://example.com"
	case "integer":
		if strings.Contains(lowerName, "id") {
			return 1
		}
		return 0
	case "numeric":
		return 0
	case "boolean":
		return false
	case "array":
		return []any{}
	case "object":
		return map[string]any{}
	case "date":
		return "2025-01-01"
	case "file":
		return nil
	}

	switch {
	case strings.Contains(lowerName, "email"):
		return "user@example.com"
	case lowerName == "name", strings.HasSuffix(lowerName, "_name"):
		return "John Doe"
	case strings.Contains(lowerName, "password"):
		return "password"
	case strings.Contains(lowerName, "phone"):
		return "+1234567890"
	case strings.Contains(lowerName, "url"):
		return "https://example.com"
	case strings.Contains(lowerName, "slug"):
		return "example"
	case strings.Contains(lowerName, "title"):
		return "Example Title"
	case strings.Contains(lowerName, "description"), strings.Contains(lowerName, "body"), strings.Contains(lowerName, "content"):
		return "Example content"
	case strings.HasSuffix(lowerName, "_id"), lowerName == "id":
		return 1
	case strings.Contains(lowerName, "address"):
		return "123 Example St"
	case strings.Contains(lowerName, "city"):
		return "Lisbon"
	case strings.Contains(lowerName, "country"):
		return "Portugal"
	}

	_ = rules
	return ""
}

func buildExampleBody(fields []InferredField) map[string]any {
	body := make(map[string]any, len(fields))
	for _, f := range fields {
		body[f.Name] = f.Example
	}
	return body
}

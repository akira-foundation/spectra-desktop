package mock

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func GenerateBody(method, path string, fields []string, pathParams map[string]string) string {
	wantsList := pathLooksLikeListEndpoint(method, path)
	obj := buildFakeRecord(fields, pathParams)

	if wantsList {
		items := []map[string]any{obj, buildFakeRecord(fields, pathParams), buildFakeRecord(fields, pathParams)}
		payload := map[string]any{
			"data": items,
			"meta": map[string]any{
				"total":    len(items),
				"page":     1,
				"per_page": len(items),
			},
		}
		raw, _ := json.MarshalIndent(payload, "", "  ")
		return string(raw)
	}

	payload := map[string]any{"data": obj, "message": "OK"}
	raw, _ := json.MarshalIndent(payload, "", "  ")
	return string(raw)
}

func pathLooksLikeListEndpoint(method, path string) bool {
	if !strings.EqualFold(method, "GET") {
		return false
	}
	clean := strings.TrimSuffix(path, "/")
	last := lastPathSegment(clean)
	if strings.Contains(last, "{") {
		return false
	}
	return strings.HasSuffix(last, "s") || last == "list" || last == "index"
}

func lastPathSegment(p string) string {
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		return p
	}
	return p[idx+1:]
}

func buildFakeRecord(fields []string, pathParams map[string]string) map[string]any {
	obj := map[string]any{}
	if len(fields) == 0 {
		obj["id"] = randomFakeID()
		obj["created_at"] = currentISOTimestamp()
	}
	for _, name := range fields {
		obj[name] = fakeValueForFieldName(name)
	}
	for k, v := range pathParams {
		obj[k] = coercePathParamValue(v)
	}
	return obj
}

func fakeValueForFieldName(name string) any {
	lower := strings.ToLower(name)
	switch {
	case lower == "id" || strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "id"):
		if lower == "uuid" || strings.Contains(lower, "uuid") {
			return randomFakeUUID()
		}
		return randomFakeID()
	case strings.Contains(lower, "email"):
		return randomFakeEmail()
	case strings.Contains(lower, "username") || lower == "user_name":
		return randomFakeUsername()
	case strings.Contains(lower, "first_name"):
		return "Alex"
	case strings.Contains(lower, "last_name"):
		return "Doe"
	case strings.Contains(lower, "name"):
		return randomFakeFullName()
	case strings.Contains(lower, "phone"):
		return "+1 555-0100"
	case strings.Contains(lower, "address"):
		return "742 Evergreen Terrace"
	case strings.Contains(lower, "city"):
		return "Springfield"
	case strings.Contains(lower, "country"):
		return "US"
	case strings.Contains(lower, "url") || strings.Contains(lower, "link"):
		return "https://example.com"
	case strings.Contains(lower, "avatar") || strings.Contains(lower, "image") || strings.Contains(lower, "photo"):
		return "https://example.com/avatar.png"
	case strings.HasPrefix(lower, "is_") || strings.HasPrefix(lower, "has_") || strings.HasPrefix(lower, "can_"):
		return false
	case strings.Contains(lower, "password"):
		return nil
	case strings.HasSuffix(lower, "_at") || lower == "date" || strings.Contains(lower, "_date"):
		return currentISOTimestamp()
	case strings.Contains(lower, "count") || strings.Contains(lower, "amount") || strings.Contains(lower, "total"):
		return rand.Intn(1000)
	case strings.Contains(lower, "price") || strings.Contains(lower, "cost"):
		return float64(rand.Intn(10000)) / 100.0
	case strings.Contains(lower, "description") || strings.Contains(lower, "bio") || strings.Contains(lower, "notes"):
		return "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	case strings.Contains(lower, "role"):
		return "user"
	case strings.Contains(lower, "status"):
		return "active"
	case strings.Contains(lower, "token"):
		return "tok_" + randomFakeToken(24)
	}
	return ""
}

func coercePathParamValue(s string) any {
	if s == "" {
		return s
	}
	allDigits := true
	for _, r := range s {
		if r < '0' || r > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		var n int
		_, err := fmt.Sscanf(s, "%d", &n)
		if err == nil {
			return n
		}
	}
	return s
}

func currentISOTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func randomFakeID() int {
	return rand.Intn(9999) + 1
}

func randomFakeEmail() string {
	users := []string{"alex", "sam", "jordan", "morgan", "taylor", "casey"}
	return users[rand.Intn(len(users))] + "@example.com"
}

func randomFakeUsername() string {
	return []string{"alex", "sam", "jordan", "morgan", "taylor", "casey"}[rand.Intn(6)]
}

func randomFakeFullName() string {
	return []string{"Alex Carter", "Sam Riley", "Jordan Lee", "Morgan Hayes", "Taylor Kim"}[rand.Intn(5)]
}

func randomFakeUUID() string {
	const chars = "0123456789abcdef"
	parts := []int{8, 4, 4, 4, 12}
	out := make([]byte, 0, 36)
	for i, n := range parts {
		if i > 0 {
			out = append(out, '-')
		}
		for j := 0; j < n; j++ {
			out = append(out, chars[rand.Intn(len(chars))])
		}
	}
	return string(out)
}

func randomFakeToken(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	out := make([]byte, n)
	for i := range out {
		out[i] = chars[rand.Intn(len(chars))]
	}
	return string(out)
}

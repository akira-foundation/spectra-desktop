package laravel

import (
	"encoding/json"
	"strings"

	"spectra-desktop/internal/core"
)

func lookupString(m map[string]any, path []string) (string, bool) {
	v, ok := lookup(m, path)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(s), s != ""
}

func lookupObject(m map[string]any, path []string) (map[string]any, bool) {
	v, ok := lookup(m, path)
	if !ok {
		return nil, false
	}
	obj, ok := v.(map[string]any)
	return obj, ok
}

func lookup(m map[string]any, path []string) (any, bool) {
	var cur any = m
	for _, key := range path {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, exists := obj[key]
		if !exists {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

func buildUser(obj map[string]any) *core.AuthUser {
	if obj == nil {
		return nil
	}
	user := &core.AuthUser{}
	user.ID = firstString(obj, "id", "uuid", "sub")
	user.Name = firstString(obj, "name", "full_name", "fullname", "display_name", "displayName")
	user.Username = firstString(obj, "username", "user_name", "userName", "login", "handle")
	user.Email = firstString(obj, "email", "email_address", "mail")
	user.Role = firstString(obj, "role", "role_name", "type")
	if user.ID == "" && user.Name == "" && user.Username == "" && user.Email == "" {
		return nil
	}
	if raw, err := json.Marshal(obj); err == nil {
		user.Raw = string(raw)
	}
	return user
}

func firstString(obj map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := obj[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

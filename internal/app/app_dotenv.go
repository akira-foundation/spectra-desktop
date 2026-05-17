package app

import (
	"os"
	"path/filepath"
	"strings"
)

func readDotenvAppURL(projectPath string) string {
	candidates := []string{".env", ".env.local", ".env.example"}
	for _, name := range candidates {
		path := filepath.Join(projectPath, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if url := extractDotenvKey(string(data), "APP_URL"); url != "" {
			return url
		}
	}
	return ""
}

func extractDotenvKey(content, key string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		idx := strings.Index(trimmed, "=")
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(trimmed[:idx])
		if k != key {
			continue
		}
		v := strings.TrimSpace(trimmed[idx+1:])
		v = strings.Trim(v, `"'`)
		if v == "" {
			continue
		}
		return v
	}
	return ""
}

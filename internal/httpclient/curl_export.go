package httpclient

import (
	"fmt"
	"strings"
)

// FormatCurl returns a shell-escaped curl command for the given request.
func FormatCurl(method, fullURL string, headers map[string]string, body string) string {
	parts := []string{"curl"}
	if method != "" && strings.ToUpper(method) != "GET" {
		parts = append(parts, "-X", strings.ToUpper(method))
	}
	parts = append(parts, shellQuote(fullURL))
	for k, v := range headers {
		parts = append(parts, "-H", shellQuote(fmt.Sprintf("%s: %s", k, v)))
	}
	if body != "" {
		parts = append(parts, "--data-raw", shellQuote(body))
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if !strings.ContainsAny(s, " \t\n'\"\\$`") {
		return s
	}
	// Single-quote and escape embedded single quotes.
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

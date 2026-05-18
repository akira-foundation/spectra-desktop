package spectra

import (
	"context"
	"database/sql"
	"encoding/json"
)

func scrubSensitiveHistoryColumns(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "SELECT id, request_headers, response_headers FROM main.request_history")
	if err != nil {
		return err
	}
	defer rows.Close()
	type entry struct {
		id              string
		requestHeaders  string
		responseHeaders string
	}
	var batch []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.id, &e.requestHeaders, &e.responseHeaders); err != nil {
			return err
		}
		batch = append(batch, e)
	}
	rows.Close()
	for _, e := range batch {
		stmt := `UPDATE main.request_history SET request_headers = ?, response_headers = ? WHERE id = ?`
		if _, err := db.ExecContext(ctx, stmt,
			redactSensitiveHeadersJSON(e.requestHeaders),
			redactSensitiveHeadersJSON(e.responseHeaders),
			e.id); err != nil {
			return err
		}
	}
	return nil
}

func redactSensitiveHeadersJSON(headersJSON string) string {
	if headersJSON == "" {
		return headersJSON
	}
	var asMultiValue map[string][]string
	if err := json.Unmarshal([]byte(headersJSON), &asMultiValue); err == nil {
		for key := range asMultiValue {
			if isSensitiveHeader(key) {
				asMultiValue[key] = []string{"[redacted]"}
			}
		}
		raw, _ := json.Marshal(asMultiValue)
		return string(raw)
	}
	var asSingle map[string]string
	if err := json.Unmarshal([]byte(headersJSON), &asSingle); err == nil {
		for key := range asSingle {
			if isSensitiveHeader(key) {
				asSingle[key] = "[redacted]"
			}
		}
		raw, _ := json.Marshal(asSingle)
		return string(raw)
	}
	return headersJSON
}

func isSensitiveHeader(name string) bool {
	switch lowerASCII(name) {
	case "authorization", "cookie", "set-cookie", "x-api-key", "x-auth-token", "x-otp", "proxy-authorization":
		return true
	}
	return false
}

func lowerASCII(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}

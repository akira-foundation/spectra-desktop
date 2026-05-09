package httpclient

import (
	"net/url"
	"strings"
)

// CurlImport is a parsed curl command, suitable for filling a request form.
type CurlImport struct {
	Method  string
	URL     string
	Path    string
	BaseURL string
	Headers map[string]string
	Body    string
	Query   map[string]string
}

// ParseCurl parses a `curl` command (as copied from Chrome/Firefox DevTools or
// shells) into a CurlImport. It is intentionally lenient: it understands the
// most common flags (-X, -H, --data, --data-raw, --data-binary, -d, -F, -u,
// -b/--cookie) and ignores anything it does not recognise.
func ParseCurl(raw string) (*CurlImport, error) {
	tokens, err := tokenizeShell(raw)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, nil
	}
	if strings.EqualFold(tokens[0], "curl") {
		tokens = tokens[1:]
	}
	out := &CurlImport{
		Method:  "GET",
		Headers: map[string]string{},
		Query:   map[string]string{},
	}
	dataParts := []string{}
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		switch {
		case t == "-X" || t == "--request":
			i++
			if i < len(tokens) {
				out.Method = strings.ToUpper(tokens[i])
			}
		case t == "-H" || t == "--header":
			i++
			if i < len(tokens) {
				k, v := splitHeader(tokens[i])
				if k != "" {
					out.Headers[k] = v
				}
			}
		case t == "-d" || t == "--data" || t == "--data-raw" || t == "--data-binary" || t == "--data-ascii":
			i++
			if i < len(tokens) {
				dataParts = append(dataParts, tokens[i])
			}
		case t == "-F" || t == "--form":
			// multipart — currently surfaced as raw value pairs in body comment
			i++
			if i < len(tokens) {
				dataParts = append(dataParts, tokens[i])
			}
		case t == "-u" || t == "--user":
			i++
			if i < len(tokens) {
				out.Headers["Authorization"] = "Basic " + tokens[i]
			}
		case t == "-b" || t == "--cookie":
			i++
			if i < len(tokens) {
				out.Headers["Cookie"] = tokens[i]
			}
		case t == "--url":
			i++
			if i < len(tokens) {
				out.URL = tokens[i]
			}
		case strings.HasPrefix(t, "--"):
			// skip long flag values we don't handle
			if i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "-") {
				i++
			}
		case strings.HasPrefix(t, "-"):
			// skip short flag values we don't handle
			if i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "-") {
				i++
			}
		default:
			if out.URL == "" {
				out.URL = t
			}
		}
	}
	if len(dataParts) > 0 {
		out.Body = strings.Join(dataParts, "&")
		if out.Method == "GET" {
			out.Method = "POST"
		}
	}
	if out.URL != "" {
		if u, err := url.Parse(out.URL); err == nil {
			out.BaseURL = u.Scheme + "://" + u.Host
			out.Path = u.Path
			if u.RawQuery != "" {
				for k, vs := range u.Query() {
					if len(vs) > 0 {
						out.Query[k] = vs[0]
					}
				}
			}
		}
	}
	return out, nil
}

func splitHeader(raw string) (string, string) {
	idx := strings.Index(raw, ":")
	if idx < 0 {
		return "", ""
	}
	return strings.TrimSpace(raw[:idx]), strings.TrimSpace(raw[idx+1:])
}

// tokenizeShell splits the input into shell-like tokens, respecting single and
// double quotes and backslash escapes. Newline continuations (`\\\n`) are
// treated as whitespace.
func tokenizeShell(raw string) ([]string, error) {
	tokens := []string{}
	cur := strings.Builder{}
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	type quote int
	const (
		none quote = iota
		single
		double
	)
	mode := none
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch {
		case mode == single:
			if c == '\'' {
				mode = none
			} else {
				cur.WriteByte(c)
			}
		case mode == double:
			if c == '"' {
				mode = none
			} else if c == '\\' && i+1 < len(raw) {
				next := raw[i+1]
				if next == '"' || next == '\\' || next == '$' || next == '`' || next == '\n' {
					if next != '\n' {
						cur.WriteByte(next)
					}
					i++
				} else {
					cur.WriteByte(c)
				}
			} else {
				cur.WriteByte(c)
			}
		default:
			switch c {
			case '\'':
				mode = single
			case '"':
				mode = double
			case '\\':
				if i+1 < len(raw) {
					next := raw[i+1]
					if next == '\n' {
						i++
						continue
					}
					cur.WriteByte(next)
					i++
				}
			case ' ', '\t', '\n', '\r':
				flush()
			default:
				cur.WriteByte(c)
			}
		}
	}
	flush()
	return tokens, nil
}

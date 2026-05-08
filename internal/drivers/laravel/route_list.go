package laravel

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type rawRoute struct {
	Method     string          `json:"method"`
	URI        string          `json:"uri"`
	Name       string          `json:"name"`
	Action     string          `json:"action"`
	Middleware json.RawMessage `json:"middleware"`
	Domain     string          `json:"domain,omitempty"`
}

func runArtisanRouteList(ctx context.Context, projectPath string) ([]rawRoute, error) {
	if _, err := exec.LookPath("php"); err != nil {
		return nil, ErrPHPNotFound
	}
	artisan := filepath.Join(projectPath, "artisan")
	if _, err := os.Stat(artisan); err != nil {
		return nil, ErrArtisanMissing
	}

	cmd := exec.CommandContext(ctx, "php", "artisan", "route:list", "--json")
	cmd.Dir = projectPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		exitCode := -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		return nil, &ArtisanFailedError{ExitCode: exitCode, Stderr: strings.TrimSpace(stderr.String())}
	}

	out := bytes.TrimSpace(stdout.Bytes())
	if len(out) == 0 {
		return nil, ErrNoRoutes
	}

	jsonBytes := extractJSONArray(out)
	if jsonBytes == nil {
		return nil, fmt.Errorf("%w: no JSON array found in output", ErrInvalidJSON)
	}

	var routes []rawRoute
	if err := json.Unmarshal(jsonBytes, &routes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return routes, nil
}

func extractJSONArray(data []byte) []byte {
	var best []byte
	bestCount := 0
	for offset := 0; offset < len(data); {
		idx := bytes.IndexByte(data[offset:], '[')
		if idx == -1 {
			break
		}
		start := offset + idx
		end := matchBalancedArray(data, start)
		if end > start {
			candidate := data[start : end+1]
			var probe []rawRoute
			if json.Unmarshal(candidate, &probe) == nil && len(probe) > bestCount {
				best = candidate
				bestCount = len(probe)
			}
			offset = end + 1
		} else {
			offset = start + 1
		}
	}
	return best
}

func matchBalancedArray(data []byte, start int) int {
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(data); i++ {
		c := data[i]
		if escape {
			escape = false
			continue
		}
		if inString {
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		switch c {
		case '"':
			inString = true
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

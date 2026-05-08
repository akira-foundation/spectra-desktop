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

	var routes []rawRoute
	if err := json.Unmarshal(out, &routes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}
	return routes, nil
}

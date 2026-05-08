package laravel

import (
	"errors"
	"fmt"
)

var (
	ErrNotLaravel     = errors.New("not a laravel project")
	ErrPHPNotFound    = errors.New("php not found in PATH")
	ErrArtisanMissing = errors.New("artisan not found in project root")
	ErrInvalidJSON    = errors.New("invalid JSON output from artisan")
	ErrNoRoutes       = errors.New("no routes returned by artisan")
)

type ArtisanFailedError struct {
	ExitCode int
	Stderr   string
}

func (e *ArtisanFailedError) Error() string {
	return fmt.Sprintf("artisan exited with code %d: %s", e.ExitCode, truncate(e.Stderr, 240))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

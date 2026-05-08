package laravel

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"spectra-desktop/internal/core"
)

func (d *Driver) Stats(ctx context.Context, projectPath string) (core.StatsReport, error) {
	endpoints, err := d.Scan(ctx, projectPath)
	if err != nil {
		return core.StatsReport{}, err
	}

	controllers := make(map[string]struct{})
	middleware := make(map[string]struct{})
	formRequests := make(map[string]struct{})
	for _, ep := range endpoints {
		if ep.Handler != "" && ep.Handler != "Closure" {
			if at := strings.LastIndex(ep.Handler, "@"); at > 0 {
				controllers[ep.Handler[:at]] = struct{}{}
			} else {
				controllers[ep.Handler] = struct{}{}
			}
		}
		for _, m := range ep.Middleware {
			middleware[m] = struct{}{}
		}
		if ep.RequestSchema != "" && strings.Contains(ep.RequestSchema, "form_request") {
			formRequests[ep.Handler] = struct{}{}
		}
	}

	models := countModels(projectPath)

	return core.StatsReport{
		Cards: []core.StatCard{
			{Key: "routes", Kind: core.StatRoutes, Label: "Routes", Value: len(endpoints)},
			{Key: "controllers", Kind: core.StatControllers, Label: "Controllers", Value: len(controllers)},
			{Key: "middleware", Kind: core.StatMiddleware, Label: "Middleware", Value: len(middleware)},
			{Key: "form_requests", Kind: core.StatFormRequests, Label: "Form Requests", Value: len(formRequests)},
			{Key: "models", Kind: core.StatModels, Label: "Models", Value: models},
		},
	}, nil
}

func countModels(projectPath string) int {
	dir := filepath.Join(projectPath, "app", "Models")
	count := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".php") {
			count++
		}
		return nil
	})
	return count
}

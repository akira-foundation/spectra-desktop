package app

import (
	"fmt"
	"os"
	"spectra-desktop/internal/exporter/openapi"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) ExportOpenAPI(projectID string) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("project id required")
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return "", err
	}
	endpoints, err := a.endpoints.List(a.ctx, projectID)
	if err != nil {
		return "", err
	}
	spec := openapi.Build(project, endpoints)
	return openapi.ToJSON(spec)
}

func (a *App) SaveOpenAPIToFile(projectID string) (string, error) {
	content, err := a.ExportOpenAPI(projectID)
	if err != nil {
		return "", err
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return "", err
	}
	defaultName := strings.ReplaceAll(strings.ToLower(project.Name), " ", "-") + "-openapi.json"
	target, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save OpenAPI spec",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON files", Pattern: "*.json"},
		},
	})
	if err != nil || target == "" {
		return "", err
	}
	if err := os.WriteFile(target, []byte(content), 0644); err != nil {
		return "", err
	}
	return target, nil
}

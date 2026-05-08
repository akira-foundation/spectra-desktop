package app

import (
	"context"
	"fmt"
	"log"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/drivers/laravel"
	"spectra-desktop/internal/httpclient"
	"spectra-desktop/internal/repository"
	"spectra-desktop/internal/storage"
	"spectra-desktop/internal/watcher"
	"spectra-desktop/internal/workspace"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx       context.Context
	scanner   *core.Scanner
	workspace *workspace.Service
	storage   *storage.Storage
	projects  domain.ProjectRepository
	settings  domain.SettingsRepository
	endpoints domain.EndpointRepository
	http      *httpclient.Client
	watcher   *watcher.Watcher
}

func New() (*App, error) {
	scanner := core.NewScanner()
	scanner.Register(laravel.New())

	store := storage.New()
	if err := store.Open(""); err != nil {
		return nil, fmt.Errorf("open storage: %w", err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("migrate storage: %w", err)
	}

	return &App{
		scanner:   scanner,
		workspace: workspace.NewService(),
		storage:   store,
		projects:  repository.NewProjectRepository(store.DB),
		settings:  repository.NewSettingsRepository(store.DB),
		endpoints: repository.NewEndpointRepository(store.DB),
		http:      httpclient.New(),
		watcher:   watcher.New(),
	}, nil
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown(_ context.Context) {
	if a.storage != nil {
		if err := a.storage.Close(); err != nil {
			log.Printf("close storage: %v", err)
		}
	}
}

func (a *App) OpenProject(path string) (*workspace.Workspace, error) {
	return a.workspace.Open(path)
}

func (a *App) DetectFramework(path string) (string, core.DetectionResult, error) {
	driver, result, err := a.scanner.Resolve(path)
	if err != nil {
		return "", core.DetectionResult{}, err
	}
	return driver.Name(), result, nil
}

func (a *App) SelectProjectFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                "Select Project Folder",
		CanCreateDirectories: false,
	})
}

type ProjectInfo struct {
	Path             string               `json:"path"`
	Name             string               `json:"name"`
	Framework        string               `json:"framework"`
	FrameworkVersion string               `json:"frameworkVersion"`
	Detection        core.DetectionResult `json:"detection"`
}

func (a *App) InspectProject(path string) (ProjectInfo, error) {
	ws, err := a.workspace.Open(path)
	if err != nil {
		return ProjectInfo{}, err
	}
	info := ProjectInfo{
		Path:      ws.Path,
		Name:      ws.Name,
		Framework: "other",
	}
	driver, det, err := a.scanner.Resolve(ws.Path)
	if err == nil {
		info.Framework = driver.Name()
		info.Detection = det
	}
	return info, nil
}

func (a *App) Drivers() []string {
	names := make([]string, 0)
	for _, d := range a.scanner.Drivers() {
		names = append(names, d.Name())
	}
	return names
}

func (a *App) ListProjects() ([]domain.Project, error) {
	return a.projects.List(a.ctx)
}

func (a *App) SaveProject(input domain.ProjectInput) (*domain.Project, error) {
	return a.projects.Save(a.ctx, input)
}

func (a *App) DeleteProject(id string) error {
	return a.projects.Delete(a.ctx, id)
}

func (a *App) MarkProjectSynced(id string) error {
	return a.projects.MarkSynced(a.ctx, id)
}

func (a *App) GetActiveProjectID() (string, error) {
	return a.settings.Get(a.ctx, domain.SettingActiveProjectID)
}

func (a *App) SetActiveProjectID(id string) error {
	if id == "" {
		return a.settings.Delete(a.ctx, domain.SettingActiveProjectID)
	}
	return a.settings.Set(a.ctx, domain.SettingActiveProjectID, id)
}

func (a *App) ScanWorkspace(projectID string) ([]core.Endpoint, error) {
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	driver, _, err := a.scanner.Resolve(project.Path)
	if err != nil {
		return nil, err
	}
	endpoints, err := driver.Scan(a.ctx, project.Path)
	if err != nil {
		return nil, err
	}
	if err := a.endpoints.Replace(a.ctx, projectID, endpoints); err != nil {
		return nil, fmt.Errorf("persist endpoints: %w", err)
	}
	return endpoints, nil
}

func (a *App) ScanActiveProject() ([]core.Endpoint, error) {
	id, err := a.settings.Get(a.ctx, domain.SettingActiveProjectID)
	if err != nil {
		return nil, err
	}
	if id == "" {
		return []core.Endpoint{}, nil
	}
	return a.ScanWorkspace(id)
}

func (a *App) ListEndpoints(projectID string) ([]core.Endpoint, error) {
	if projectID == "" {
		return []core.Endpoint{}, nil
	}
	return a.endpoints.List(a.ctx, projectID)
}

func (a *App) GetProjectStats(projectID string) (domain.ProjectStats, error) {
	if projectID == "" {
		return domain.ProjectStats{}, nil
	}
	return a.endpoints.Stats(a.ctx, projectID)
}

func (a *App) DetectProject(id string) (core.DetectionResult, error) {
	project, err := a.projects.GetByID(a.ctx, id)
	if err != nil {
		return core.DetectionResult{}, err
	}
	for _, d := range a.scanner.Drivers() {
		result := d.Detect(project.Path)
		if result.Detected {
			return result, nil
		}
	}
	return core.DetectionResult{}, nil
}

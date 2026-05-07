package app

import (
	"context"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/drivers/laravel"
	"spectra-desktop/internal/httpclient"
	"spectra-desktop/internal/storage"
	"spectra-desktop/internal/watcher"
	"spectra-desktop/internal/workspace"
)

type App struct {
	ctx       context.Context
	scanner   *core.Scanner
	workspace *workspace.Service
	storage   *storage.Storage
	http      *httpclient.Client
	watcher   *watcher.Watcher
}

func New() *App {
	scanner := core.NewScanner()
	scanner.Register(laravel.New())
	return &App{
		scanner:   scanner,
		workspace: workspace.NewService(),
		storage:   storage.New(),
		http:      httpclient.New(),
		watcher:   watcher.New(),
	}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
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

func (a *App) ListEndpoints() ([]core.Endpoint, error) {
	ws := a.workspace.Current()
	if ws == nil {
		return []core.Endpoint{}, nil
	}
	return a.scanner.Scan(a.ctx, ws.Path)
}

func (a *App) Drivers() []string {
	names := make([]string, 0)
	for _, d := range a.scanner.Drivers() {
		names = append(names, d.Name())
	}
	return names
}

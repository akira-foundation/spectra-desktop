package laravel

import (
	"context"

	"spectra-desktop/internal/core"
)

const DriverName = "laravel"

type Driver struct{}

func New() *Driver {
	return &Driver{}
}

func (d *Driver) Name() string {
	return DriverName
}

func (d *Driver) Detect(projectPath string) core.DetectionResult {
	return detect(projectPath)
}

func (d *Driver) Scan(ctx context.Context, projectPath string) ([]core.Endpoint, error) {
	return scanRoutes(ctx, projectPath)
}

func (d *Driver) Capabilities() core.DriverCapabilities {
	return core.DriverCapabilities{
		ScanRoutes:      true,
		ScanControllers: false,
		ResolveAuth:     false,
		WatchChanges:    false,
		RunRequests:     false,
	}
}

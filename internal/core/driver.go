package core

import "context"

type FrameworkDriver interface {
	Name() string
	Detect(projectPath string) DetectionResult
	Scan(ctx context.Context, projectPath string) ([]Endpoint, error)
	Capabilities() DriverCapabilities
}

type DetectionResult struct {
	Detected   bool     `json:"detected"`
	Confidence float64  `json:"confidence"`
	Version    string   `json:"version,omitempty"`
	Markers    []string `json:"markers,omitempty"`
}

type DriverCapabilities struct {
	ScanRoutes      bool `json:"scanRoutes"`
	ScanControllers bool `json:"scanControllers"`
	ResolveAuth     bool `json:"resolveAuth"`
	WatchChanges    bool `json:"watchChanges"`
	RunRequests     bool `json:"runRequests"`
}

package core

import "context"

type FrameworkDriver interface {
	Name() string
	Detect(projectPath string) DetectionResult
	Scan(ctx context.Context, projectPath string) ([]Endpoint, error)
	Defaults() DriverDefaults
	Capabilities() DriverCapabilities
}

type DriverDefaults struct {
	BaseURL string `json:"baseURL"`
	Ports   []int  `json:"ports,omitempty"`
}

type DetectionResult struct {
	Detected   bool     `json:"detected"`
	Confidence float64  `json:"confidence"`
	Version    string   `json:"version,omitempty"`
	Markers    []string `json:"markers,omitempty"`
}

type DriverCapabilities struct {
	ScanRoutes      bool     `json:"scanRoutes"`
	ScanControllers bool     `json:"scanControllers"`
	ResolveAuth     bool     `json:"resolveAuth"`
	WatchChanges    bool     `json:"watchChanges"`
	RunRequests     bool     `json:"runRequests"`
	Stats           []string `json:"stats,omitempty"`
	HasModels       bool     `json:"hasModels,omitempty"`
	HasControllers  bool     `json:"hasControllers,omitempty"`
	HasMiddleware   bool     `json:"hasMiddleware,omitempty"`
	HasFormRequests bool     `json:"hasFormRequests,omitempty"`
	HasJobs         bool     `json:"hasJobs,omitempty"`
	HasMailers      bool     `json:"hasMailers,omitempty"`
	HasServices     bool     `json:"hasServices,omitempty"`
}

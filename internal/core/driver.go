package core

import "context"

type FrameworkDriver interface {
	Name() string
	Detect(projectPath string) DetectionResult
	Scan(ctx context.Context, projectPath string) ([]Endpoint, error)
	Defaults() DriverDefaults
	Capabilities() DriverCapabilities
}

// BodyValueGen is implemented by drivers that can generate example values
// for fields based on field name, type and validation rules.
type BodyValueGen interface {
	GenerateValue(name, fieldType string, rules []string) any
}

// ExceptionFormatter is implemented by drivers that can parse a framework
// error response (e.g. Laravel Whoops/handler) into a structured form.
type ExceptionFormatter interface {
	FormatException(body string, status int) (FormattedException, bool)
}

type FormattedException struct {
	Message string                `json:"message"`
	Class   string                `json:"class,omitempty"`
	File    string                `json:"file,omitempty"`
	Line    int                   `json:"line,omitempty"`
	Trace   []FormattedTraceFrame `json:"trace,omitempty"`
	Extra   map[string]any        `json:"extra,omitempty"`
}

type FormattedTraceFrame struct {
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Function string `json:"function,omitempty"`
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

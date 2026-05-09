package laravel

import (
	"context"
	"encoding/json"
	"strings"

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
	raws, err := runArtisanRouteList(ctx, projectPath)
	if err != nil {
		return nil, err
	}
	endpoints := normalize(raws)
	if len(endpoints) == 0 {
		return nil, ErrNoRoutes
	}
	enrichSchemas(projectPath, endpoints)
	enrichAuthRoles(endpoints)
	return endpoints, nil
}

func enrichAuthRoles(endpoints []core.Endpoint) {
	cap := AuthCapability{}
	for i := range endpoints {
		hint := cap.DetectAuthRole(endpoints[i])
		if hint.Role == core.AuthRoleNone {
			continue
		}
		endpoints[i].AuthRole = hint.Role
		endpoints[i].AuthHint = string(hint.Confidence) + ": " + hint.Reason
	}
}

func enrichSchemas(projectPath string, endpoints []core.Endpoint) {
	for i := range endpoints {
		ep := &endpoints[i]
		if ep.Handler == "" || ep.Handler == "Closure" {
			continue
		}
		methodName := ""
		if at := strings.LastIndex(ep.Handler, "@"); at > 0 && at+1 < len(ep.Handler) {
			methodName = strings.TrimSpace(ep.Handler[at+1:])
		}
		if methodName == "" {
			methodName = "__invoke"
		}
		schema := tryInferSchema(projectPath, ep.Handler, methodName)
		if schema == nil {
			continue
		}
		raw, err := json.Marshal(schema)
		if err != nil {
			continue
		}
		ep.RequestSchema = string(raw)
	}
}

func tryInferSchema(projectPath, handler, methodName string) *RequestSchema {
	if schema, err := inferFromFormRequest(projectPath, handler, methodName); err == nil && schema != nil {
		return schema
	}
	if schema, err := inferFromInlineValidation(projectPath, handler, methodName); err == nil && schema != nil {
		return schema
	}
	return nil
}

func (d *Driver) Defaults() core.DriverDefaults {
	return core.DriverDefaults{
		BaseURL: "http://localhost:8000",
		Ports:   []int{8000},
	}
}

func (d *Driver) Capabilities() core.DriverCapabilities {
	return core.DriverCapabilities{
		ScanRoutes:      true,
		ScanControllers: true,
		ResolveAuth:     true,
		WatchChanges:    false,
		RunRequests:     true,
		Stats:           []string{"routes", "controllers", "middleware", "form_requests", "models"},
		HasModels:       true,
		HasControllers:  true,
		HasMiddleware:   true,
		HasFormRequests: true,
	}
}

// GenerateValue implements core.BodyValueGen.
func (d *Driver) GenerateValue(name, fieldType string, rules []string) any {
	return RegenerateValue(name, fieldType, rules)
}

// FormatException implements core.ExceptionFormatter for Laravel error responses.
func (d *Driver) FormatException(body string, status int) (core.FormattedException, bool) {
	return parseLaravelException(body, status)
}

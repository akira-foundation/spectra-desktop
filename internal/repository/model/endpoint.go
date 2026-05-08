package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/core"
)

type Endpoint struct {
	bun.BaseModel `bun:"table:endpoints"`

	ID         string    `bun:"id,pk"`
	ProjectID  string    `bun:"project_id,notnull"`
	Method     string    `bun:"method,notnull"`
	Path       string    `bun:"path,notnull"`
	Name       string    `bun:"name,notnull,default:''"`
	Handler    string    `bun:"handler,notnull,default:''"`
	Middleware string    `bun:"middleware,notnull,default:'[]'"`
	Parameters string    `bun:"parameters,notnull,default:'[]'"`
	Tags       string    `bun:"tags,notnull,default:'[]'"`
	SourceFile string    `bun:"source_file,notnull,default:''"`
	SourceLine int       `bun:"source_line,notnull,default:0"`
	Framework  string    `bun:"framework,notnull,default:''"`
	Confidence float64   `bun:"confidence,notnull,default:0"`
	ScannedAt  time.Time `bun:"scanned_at,notnull"`
	CreatedAt  time.Time `bun:"created_at,notnull"`
	UpdatedAt  time.Time `bun:"updated_at,notnull"`
}

func (e Endpoint) ToCore() core.Endpoint {
	return core.Endpoint{
		ID:         e.ID,
		Method:     core.HTTPMethod(e.Method),
		Path:       e.Path,
		Name:       e.Name,
		Handler:    e.Handler,
		Middleware: decodeStringSlice(e.Middleware),
		Parameters: decodeParameters(e.Parameters),
		Tags:       decodeStringSlice(e.Tags),
		Source: core.EndpointSource{
			File: e.SourceFile,
			Line: e.SourceLine,
		},
		Framework:  e.Framework,
		Confidence: e.Confidence,
	}
}

func EndpointFromCore(projectID string, ep core.Endpoint, scannedAt time.Time, now time.Time) Endpoint {
	return Endpoint{
		ID:         ep.ID,
		ProjectID:  projectID,
		Method:     string(ep.Method),
		Path:       ep.Path,
		Name:       ep.Name,
		Handler:    ep.Handler,
		Middleware: encodeStringSlice(ep.Middleware),
		Parameters: encodeParameters(ep.Parameters),
		Tags:       encodeStringSlice(ep.Tags),
		SourceFile: ep.Source.File,
		SourceLine: ep.Source.Line,
		Framework:  ep.Framework,
		Confidence: ep.Confidence,
		ScannedAt:  scannedAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func encodeStringSlice(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(items)
	return string(b)
}

func decodeStringSlice(raw string) []string {
	if raw == "" || raw == "[]" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func encodeParameters(items []core.Parameter) string {
	if len(items) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(items)
	return string(b)
}

func decodeParameters(raw string) []core.Parameter {
	if raw == "" || raw == "[]" {
		return nil
	}
	var out []core.Parameter
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

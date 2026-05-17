package app

import (
	"encoding/json"
	"fmt"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/httpclient"
)

type EndpointCaptureDTO struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name"`
	Source string `json:"source"`
	Path   string `json:"path"`
}

type SaveCapturesInput struct {
	ProjectID   string               `json:"projectID"`
	EndpointKey string               `json:"endpointKey"`
	Captures    []EndpointCaptureDTO `json:"captures"`
}

type CapturedValueDTO struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	EndpointKey string `json:"endpointKey,omitempty"`
	CapturedAt  int64  `json:"capturedAt,omitempty"`
}

func (a *App) ListEndpointCaptures(projectID, endpointKey string) ([]EndpointCaptureDTO, error) {
	if projectID == "" || endpointKey == "" {
		return []EndpointCaptureDTO{}, nil
	}
	rows, err := a.captures.List(a.ctx, projectID, endpointKey)
	if err != nil {
		return nil, err
	}
	out := make([]EndpointCaptureDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, EndpointCaptureDTO{
			ID:     r.ID,
			Name:   r.Name,
			Source: r.Source,
			Path:   r.Path,
		})
	}
	return out, nil
}

func (a *App) SaveEndpointCaptures(input SaveCapturesInput) error {
	if input.ProjectID == "" || input.EndpointKey == "" {
		return fmt.Errorf("project id and endpoint key required")
	}
	captures := make([]domain.EndpointCapture, 0, len(input.Captures))
	keep := make(map[string]bool, len(input.Captures))
	for i, c := range input.Captures {
		captures = append(captures, domain.EndpointCapture{
			ID:        c.ID,
			Name:      c.Name,
			Source:    c.Source,
			Path:      c.Path,
			SortOrder: i,
		})
		if c.Name != "" {
			keep[c.Name] = true
		}
	}
	if err := a.captures.Replace(a.ctx, input.ProjectID, input.EndpointKey, captures); err != nil {
		return err
	}
	if a.captured != nil {
		a.captured.pruneByEndpoint(input.ProjectID, input.EndpointKey, keep)
	}
	return nil
}

func (a *App) ListCapturedValues(projectID string) []CapturedValueDTO {
	if projectID == "" || a.captured == nil {
		return []CapturedValueDTO{}
	}
	a.captured.ensureLoaded(projectID)
	return a.captured.list(projectID)
}

func (a *App) ClearCapturedValues(projectID string) {
	if projectID == "" || a.captured == nil {
		return
	}
	a.captured.clear(projectID)
}

func (a *App) runCapturesForRequest(input ExecuteRequestInput, resp *httpclient.Response) {
	if input.ProjectID == "" || a.captured == nil {
		return
	}
	key := endpointTestKey(input.Method, input.Path)
	rows, err := a.captures.List(a.ctx, input.ProjectID, key)
	if err != nil || len(rows) == 0 {
		return
	}
	var bodyValue any
	if resp != nil && resp.Body != "" {
		_ = json.Unmarshal([]byte(resp.Body), &bodyValue)
	}
	headers := toHTTPHeader(resp.Headers)
	for _, c := range rows {
		if c.Name == "" {
			continue
		}
		val, ok := extractCaptureValue(c.Source, c.Path, bodyValue, headers)
		if !ok {
			continue
		}
		a.captured.set(input.ProjectID, c.Name, val, key)
	}
}

package app

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DatasetRowResultDTO struct {
	Index      int    `json:"index"`
	Status     int    `json:"status"`
	DurationMs int    `json:"durationMs"`
	Pass       bool   `json:"pass"`
	Error      string `json:"error,omitempty"`
}

type DatasetRunDTO struct {
	EndpointKey string                `json:"endpointKey"`
	Total       int                   `json:"total"`
	PassCount   int                   `json:"passCount"`
	FailCount   int                   `json:"failCount"`
	DurationMs  int                   `json:"durationMs"`
	Rows        []DatasetRowResultDTO `json:"rows"`
}

func (a *App) GetEndpointDataset(projectID, endpointKey string) (string, error) {
	if projectID == "" || endpointKey == "" {
		return "[]", nil
	}
	return a.datasets.Get(a.ctx, projectID, endpointKey)
}

func (a *App) SaveEndpointDataset(projectID, endpointKey, rowsJSON string) error {
	if projectID == "" || endpointKey == "" {
		return fmt.Errorf("project id and endpoint key required")
	}
	if rowsJSON == "" {
		rowsJSON = "[]"
	}
	if !json.Valid([]byte(rowsJSON)) {
		return fmt.Errorf("invalid rows json")
	}
	return a.datasets.Save(a.ctx, projectID, endpointKey, rowsJSON)
}

func (a *App) GenerateDatasetRows(endpointID string, count int) (string, error) {
	if endpointID == "" {
		return "[]", nil
	}
	if count <= 0 {
		count = 1
	}
	if count > 500 {
		count = 500
	}
	ep, err := a.endpoints.GetByID(a.ctx, endpointID)
	if err != nil || ep == nil {
		return "[]", err
	}
	projectID, _ := a.endpoints.ProjectIDOf(a.ctx, endpointID)
	var raw struct {
		Fields []RegenerateFieldInput `json:"fields"`
	}
	if ep.RequestSchema != "" {
		_ = json.Unmarshal([]byte(ep.RequestSchema), &raw)
	}
	out := make([]json.RawMessage, 0, count)
	for i := 0; i < count; i++ {
		body, err := a.regenerateFromFields(raw.Fields, projectID)
		if err != nil {
			return "[]", err
		}
		out = append(out, json.RawMessage(body))
	}
	buf, err := json.Marshal(out)
	if err != nil {
		return "[]", err
	}
	return string(buf), nil
}

func (a *App) RunEndpointDataset(projectID, endpointID string) (*DatasetRunDTO, error) {
	if projectID == "" || endpointID == "" {
		return nil, fmt.Errorf("project id and endpoint id required")
	}
	ep, err := a.endpoints.GetByID(a.ctx, endpointID)
	if err != nil || ep == nil {
		return nil, fmt.Errorf("endpoint not found")
	}
	key := endpointTestKey(string(ep.Method), ep.Path)
	rowsJSON, err := a.datasets.Get(a.ctx, projectID, key)
	if err != nil {
		return nil, err
	}
	var rows []json.RawMessage
	if err := json.Unmarshal([]byte(rowsJSON), &rows); err != nil {
		return nil, fmt.Errorf("invalid dataset rows: %w", err)
	}
	run := &DatasetRunDTO{
		EndpointKey: key,
		Total:       len(rows),
		Rows:        make([]DatasetRowResultDTO, 0, len(rows)),
	}
	runtime.EventsEmit(a.ctx, "dataset:run:start", map[string]any{
		"endpointID": endpointID,
		"total":      len(rows),
	})
	start := time.Now()
	for i, row := range rows {
		input := ExecuteRequestInput{
			ProjectID:  projectID,
			EndpointID: endpointID,
			Method:     string(ep.Method),
			Path:       ep.Path,
			Body:       string(row),
		}
		resp, sendErr := a.ExecuteRequest(input)
		result := DatasetRowResultDTO{Index: i}
		if sendErr != nil {
			result.Error = sendErr.Error()
		} else if resp != nil {
			result.Status = resp.Status
			result.DurationMs = int(resp.DurationMs)
			result.Pass = resp.Status < 400
		}
		if result.Pass {
			run.PassCount++
		} else {
			run.FailCount++
		}
		run.Rows = append(run.Rows, result)
		runtime.EventsEmit(a.ctx, "dataset:run:progress", map[string]any{
			"endpointID": endpointID,
			"index":      i,
			"total":      len(rows),
			"row":        result,
		})
	}
	run.DurationMs = int(time.Since(start).Milliseconds())
	runtime.EventsEmit(a.ctx, "dataset:run:done", run)
	return run, nil
}

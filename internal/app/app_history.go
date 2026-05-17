package app

import (
	"encoding/json"
	"log"
	"spectra-desktop/internal/assertions"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/httpclient"
	"time"
)

const historyFreeLimit = 5000

func (a *App) saveHistory(input ExecuteRequestInput, url string, headers map[string]string, resp *httpclient.Response, sendErr error, testResults []assertions.Result) {
	headersJSON, _ := json.Marshal(headers)
	respHeaders := ""
	respBody := ""
	respStatus := 0
	durationMs := 0
	sizeBytes := 0
	errStr := ""
	if resp != nil {
		respStatus = resp.Status
		respBody = resp.Body
		durationMs = int(resp.DurationMs)
		sizeBytes = resp.SizeBytes
		if rh, err := json.Marshal(resp.Headers); err == nil {
			respHeaders = string(rh)
		}
	}
	if sendErr != nil {
		errStr = sendErr.Error()
	}
	testResultsJSON := ""
	if len(testResults) > 0 {
		if buf, err := json.Marshal(testResults); err == nil {
			testResultsJSON = string(buf)
		}
	}
	entry := domain.HistoryEntry{
		ProjectID:       input.ProjectID,
		EndpointID:      input.EndpointID,
		Method:          input.Method,
		URL:             url,
		RequestHeaders:  string(headersJSON),
		RequestBody:     input.Body,
		ResponseStatus:  respStatus,
		ResponseHeaders: respHeaders,
		ResponseBody:    respBody,
		DurationMs:      durationMs,
		SizeBytes:       sizeBytes,
		Error:           errStr,
		TestResultsJSON: testResultsJSON,
	}
	if err := a.history.Save(a.ctx, entry); err != nil {
		log.Printf("save history: %v", err)
		return
	}
	if err := a.history.TrimOldest(a.ctx, input.ProjectID, historyFreeLimit); err != nil {
		log.Printf("trim history: %v", err)
	}
}

type HistoryListItem struct {
	ID             string    `json:"id"`
	EndpointID     string    `json:"endpointID,omitempty"`
	Method         string    `json:"method"`
	URL            string    `json:"url"`
	ResponseStatus int       `json:"responseStatus"`
	DurationMs     int       `json:"durationMs"`
	SizeBytes      int       `json:"sizeBytes"`
	Error          string    `json:"error,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type HistoryEntryDetail struct {
	HistoryListItem
	RequestHeaders  string          `json:"requestHeaders"`
	RequestBody     string          `json:"requestBody"`
	ResponseHeaders string          `json:"responseHeaders"`
	ResponseBody    string          `json:"responseBody"`
	TestResults     []TestResultDTO `json:"testResults,omitempty"`
}

func (a *App) ListHistory(projectID string, limit int) ([]HistoryListItem, error) {
	if projectID == "" {
		return []HistoryListItem{}, nil
	}
	if limit <= 0 {
		limit = 100
	}
	entries, err := a.history.List(a.ctx, projectID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]HistoryListItem, 0, len(entries))
	for _, e := range entries {
		out = append(out, HistoryListItem{
			ID:             e.ID,
			EndpointID:     e.EndpointID,
			Method:         e.Method,
			URL:            e.URL,
			ResponseStatus: e.ResponseStatus,
			DurationMs:     e.DurationMs,
			SizeBytes:      e.SizeBytes,
			Error:          e.Error,
			CreatedAt:      e.CreatedAt,
		})
	}
	return out, nil
}

func (a *App) GetHistoryEntry(id string) (*HistoryEntryDetail, error) {
	entry, err := a.history.GetByID(a.ctx, id)
	if err != nil || entry == nil {
		return nil, err
	}
	return &HistoryEntryDetail{
		HistoryListItem: HistoryListItem{
			ID:             entry.ID,
			EndpointID:     entry.EndpointID,
			Method:         entry.Method,
			URL:            entry.URL,
			ResponseStatus: entry.ResponseStatus,
			DurationMs:     entry.DurationMs,
			SizeBytes:      entry.SizeBytes,
			Error:          entry.Error,
			CreatedAt:      entry.CreatedAt,
		},
		RequestHeaders:  entry.RequestHeaders,
		RequestBody:     entry.RequestBody,
		ResponseHeaders: entry.ResponseHeaders,
		ResponseBody:    entry.ResponseBody,
		TestResults:     parseTestResults(entry.TestResultsJSON),
	}, nil
}

func parseTestResults(raw string) []TestResultDTO {
	if raw == "" {
		return nil
	}
	var out []TestResultDTO
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

func (a *App) ClearHistory(projectID string) error {
	return a.history.Clear(a.ctx, projectID)
}

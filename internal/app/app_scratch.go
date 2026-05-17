package app

import (
	"fmt"
	"spectra-desktop/internal/repository/model"
	"time"
)

// --- Scratch requests ---

type ScratchRequestDTO struct {
	ID           string `json:"id"`
	ProjectID    string `json:"projectID"`
	Name         string `json:"name"`
	Method       string `json:"method"`
	URL          string `json:"url"`
	HeadersJSON  string `json:"headersJson"`
	Body         string `json:"body"`
	ResponseJSON string `json:"responseJson,omitempty"`
	SortOrder    int    `json:"sortOrder"`
}

func (a *App) ListScratchRequests(projectID string) ([]ScratchRequestDTO, error) {
	rows, err := a.scratch.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]ScratchRequestDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, ScratchRequestDTO{
			ID:           r.ID,
			ProjectID:    r.ProjectID,
			Name:         r.Name,
			Method:       r.Method,
			URL:          r.URL,
			HeadersJSON:  r.HeadersJSON,
			Body:         r.Body,
			ResponseJSON: r.ResponseJSON,
			SortOrder:    r.SortOrder,
		})
	}
	return out, nil
}

func (a *App) SaveScratchRequest(input ScratchRequestDTO) (ScratchRequestDTO, error) {
	if input.ID == "" {
		input.ID = fmt.Sprintf("scr_%d_%s", time.Now().UnixNano(), randSuffix(6))
	}
	row := &model.ScratchRequest{
		ID:           input.ID,
		ProjectID:    input.ProjectID,
		Name:         input.Name,
		Method:       input.Method,
		URL:          input.URL,
		HeadersJSON:  input.HeadersJSON,
		Body:         input.Body,
		ResponseJSON: input.ResponseJSON,
		SortOrder:    input.SortOrder,
	}
	if err := a.scratch.Save(a.ctx, row); err != nil {
		return ScratchRequestDTO{}, err
	}
	input.ID = row.ID
	return input, nil
}

func (a *App) DeleteScratchRequest(id string) error {
	return a.scratch.Delete(a.ctx, id)
}

func randSuffix(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = alphabet[time.Now().UnixNano()%int64(len(alphabet))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

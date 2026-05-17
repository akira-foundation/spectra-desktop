package app

import (
	"encoding/json"
	"fmt"
	"spectra-desktop/internal/domain"
)

type CollectionItemDTO struct {
	ID              string `json:"id,omitempty"`
	EndpointID      string `json:"endpointID"`
	BodyOverride    string `json:"bodyOverride,omitempty"`
	HeadersOverride string `json:"headersOverride,omitempty"`
	SkipOnFailure   bool   `json:"skipOnFailure,omitempty"`
	IterateDataset  bool   `json:"iterateDataset,omitempty"`
}

type CollectionDTO struct {
	ID          string              `json:"id"`
	ProjectID   string              `json:"projectID"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	SortOrder   int                 `json:"sortOrder"`
	Items       []CollectionItemDTO `json:"items"`
}

type SaveCollectionInput struct {
	ID          string              `json:"id,omitempty"`
	ProjectID   string              `json:"projectID"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	SortOrder   int                 `json:"sortOrder,omitempty"`
	Items       []CollectionItemDTO `json:"items"`
}

type CollectionRunItemDTO struct {
	EndpointID  string          `json:"endpointID"`
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	Status      int             `json:"status"`
	DurationMs  int             `json:"durationMs"`
	Pass        bool            `json:"pass"`
	Skipped     bool            `json:"skipped,omitempty"`
	Error       string          `json:"error,omitempty"`
	TestResults []TestResultDTO `json:"testResults,omitempty"`
}

type CollectionRunDTO struct {
	CollectionID string                 `json:"collectionID"`
	StartedAt    int64                  `json:"startedAt"`
	DurationMs   int                    `json:"durationMs"`
	PassCount    int                    `json:"passCount"`
	FailCount    int                    `json:"failCount"`
	SkipCount    int                    `json:"skipCount"`
	Items        []CollectionRunItemDTO `json:"items"`
}

func (a *App) ListCollections(projectID string) ([]CollectionDTO, error) {
	if projectID == "" {
		return []CollectionDTO{}, nil
	}
	rows, err := a.collections.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]CollectionDTO, 0, len(rows))
	for _, c := range rows {
		out = append(out, collectionToDTO(c))
	}
	return out, nil
}

func (a *App) SaveCollection(input SaveCollectionInput) (*CollectionDTO, error) {
	if input.ProjectID == "" || input.Name == "" {
		return nil, fmt.Errorf("project id and name required")
	}
	c := domain.Collection{
		ID:          input.ID,
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		SortOrder:   input.SortOrder,
	}
	if c.ID == "" {
		created, err := a.collections.Create(a.ctx, c)
		if err != nil {
			return nil, err
		}
		c = *created
	} else {
		if err := a.collections.Update(a.ctx, c); err != nil {
			return nil, err
		}
	}
	items := make([]domain.CollectionItem, 0, len(input.Items))
	for i, it := range input.Items {
		items = append(items, domain.CollectionItem{
			ID:              it.ID,
			CollectionID:    c.ID,
			EndpointID:      it.EndpointID,
			SortOrder:       i,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
			SkipOnFailure:   it.SkipOnFailure,
			IterateDataset:  it.IterateDataset,
		})
	}
	if err := a.collections.ReplaceItems(a.ctx, c.ID, items); err != nil {
		return nil, err
	}
	full, err := a.collections.Get(a.ctx, c.ID)
	if err != nil || full == nil {
		return nil, err
	}
	dto := collectionToDTO(*full)
	return &dto, nil
}

type ExportedTest struct {
	Name     string `json:"name,omitempty"`
	Kind     string `json:"kind"`
	JSONPath string `json:"jsonPath,omitempty"`
	Op       string `json:"op,omitempty"`
	Expected string `json:"expected,omitempty"`
}

type ExportedCapture struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Path   string `json:"path"`
}

type ExportedItem struct {
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	IterateDataset  bool              `json:"iterateDataset,omitempty"`
	SkipOnFailure   bool              `json:"skipOnFailure,omitempty"`
	BodyOverride    string            `json:"bodyOverride,omitempty"`
	HeadersOverride string            `json:"headersOverride,omitempty"`
	Tests           []ExportedTest    `json:"tests,omitempty"`
	Captures        []ExportedCapture `json:"captures,omitempty"`
	Dataset         []json.RawMessage `json:"dataset,omitempty"`
}

type ExportedCollection struct {
	SpectraVersion string         `json:"spectraVersion"`
	Name           string         `json:"name"`
	Description    string         `json:"description,omitempty"`
	ExportedAt     int64          `json:"exportedAt"`
	Items          []ExportedItem `json:"items"`
}

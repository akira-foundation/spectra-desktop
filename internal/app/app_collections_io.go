package app

import (
	"encoding/json"
	"fmt"
	"os"
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) ExportCollectionToFile(id string) (string, error) {
	json, err := a.ExportCollection(id)
	if err != nil {
		return "", err
	}
	c, _ := a.collections.Get(a.ctx, id)
	defaultName := "collection.spectra.json"
	if c != nil && c.Name != "" {
		defaultName = strings.ToLower(strings.ReplaceAll(c.Name, " ", "_")) + ".spectra.json"
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export collection",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON", Pattern: "*.json"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := os.WriteFile(path, []byte(json), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ExportCollection(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("id required")
	}
	c, err := a.collections.Get(a.ctx, id)
	if err != nil || c == nil {
		return "", fmt.Errorf("collection not found")
	}
	endpoints, err := a.endpoints.List(a.ctx, c.ProjectID)
	if err != nil {
		return "", err
	}
	byID := map[string]core.Endpoint{}
	for _, e := range endpoints {
		byID[e.ID] = e
	}
	out := ExportedCollection{
		SpectraVersion: "1",
		Name:           c.Name,
		Description:    c.Description,
		ExportedAt:     time.Now().UTC().Unix(),
		Items:          make([]ExportedItem, 0, len(c.Items)),
	}
	for _, it := range c.Items {
		ep, ok := byID[it.EndpointID]
		if !ok {
			continue
		}
		key := endpointTestKey(string(ep.Method), ep.Path)
		exported := ExportedItem{
			Method:          string(ep.Method),
			Path:            ep.Path,
			IterateDataset:  it.IterateDataset,
			SkipOnFailure:   it.SkipOnFailure,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
		}
		if tests, err := a.tests.List(a.ctx, c.ProjectID, key); err == nil {
			for _, t := range tests {
				exported.Tests = append(exported.Tests, ExportedTest{
					Name:     t.Name,
					Kind:     t.Kind,
					JSONPath: t.JSONPath,
					Op:       t.Op,
					Expected: t.Expected,
				})
			}
		}
		if caps, err := a.captures.List(a.ctx, c.ProjectID, key); err == nil {
			for _, cap := range caps {
				exported.Captures = append(exported.Captures, ExportedCapture{
					Name:   cap.Name,
					Source: cap.Source,
					Path:   cap.Path,
				})
			}
		}
		if it.IterateDataset {
			if rowsJSON, err := a.datasets.Get(a.ctx, c.ProjectID, key); err == nil {
				var rows []json.RawMessage
				if json.Unmarshal([]byte(rowsJSON), &rows) == nil {
					exported.Dataset = rows
				}
			}
		}
		out.Items = append(out.Items, exported)
	}
	buf, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

type ImportCollectionResult struct {
	Collection       CollectionDTO `json:"collection"`
	MissingEndpoints []string      `json:"missingEndpoints,omitempty"`
}

func (a *App) ImportCollection(projectID, payload string) (*ImportCollectionResult, error) {
	if projectID == "" {
		return nil, fmt.Errorf("project id required")
	}
	if payload == "" {
		return nil, fmt.Errorf("payload required")
	}
	var imported ExportedCollection
	if err := json.Unmarshal([]byte(payload), &imported); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if imported.Name == "" {
		imported.Name = "Imported collection"
	}
	endpoints, err := a.endpoints.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	byKey := map[string]core.Endpoint{}
	for _, e := range endpoints {
		byKey[endpointTestKey(string(e.Method), e.Path)] = e
	}
	c := domain.Collection{
		ProjectID:   projectID,
		Name:        imported.Name,
		Description: imported.Description,
	}
	created, err := a.collections.Create(a.ctx, c)
	if err != nil {
		return nil, err
	}
	items := make([]domain.CollectionItem, 0, len(imported.Items))
	missing := []string{}
	for _, it := range imported.Items {
		key := endpointTestKey(it.Method, it.Path)
		ep, ok := byKey[key]
		if !ok {
			missing = append(missing, key)
			continue
		}
		items = append(items, domain.CollectionItem{
			CollectionID:    created.ID,
			EndpointID:      ep.ID,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
			SkipOnFailure:   it.SkipOnFailure,
			IterateDataset:  it.IterateDataset,
		})
		if len(it.Tests) > 0 {
			tests := make([]domain.EndpointTest, 0, len(it.Tests))
			for _, t := range it.Tests {
				tests = append(tests, domain.EndpointTest{
					Name:     t.Name,
					Kind:     t.Kind,
					JSONPath: t.JSONPath,
					Op:       t.Op,
					Expected: t.Expected,
				})
			}
			_ = a.tests.Replace(a.ctx, projectID, key, tests)
		}
		if len(it.Captures) > 0 {
			caps := make([]domain.EndpointCapture, 0, len(it.Captures))
			for _, cp := range it.Captures {
				caps = append(caps, domain.EndpointCapture{
					Name:   cp.Name,
					Source: cp.Source,
					Path:   cp.Path,
				})
			}
			_ = a.captures.Replace(a.ctx, projectID, key, caps)
		}
		if len(it.Dataset) > 0 {
			if buf, err := json.Marshal(it.Dataset); err == nil {
				_ = a.datasets.Save(a.ctx, projectID, key, string(buf))
			}
		}
	}
	if err := a.collections.ReplaceItems(a.ctx, created.ID, items); err != nil {
		return nil, err
	}
	full, err := a.collections.Get(a.ctx, created.ID)
	if err != nil || full == nil {
		return nil, err
	}
	dto := collectionToDTO(*full)
	return &ImportCollectionResult{Collection: dto, MissingEndpoints: missing}, nil
}

func (a *App) DeleteCollection(id string) error {
	if id == "" {
		return fmt.Errorf("id required")
	}
	return a.collections.Delete(a.ctx, id)
}

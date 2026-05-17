package app

import (
	"fmt"
	"os"
	"spectra-desktop/internal/httpclient"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type HAREntryDTO struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	BaseURL   string            `json:"baseURL,omitempty"`
	Path      string            `json:"path,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body,omitempty"`
	Query     map[string]string `json:"query,omitempty"`
	Status    int               `json:"status,omitempty"`
	Size      int               `json:"size,omitempty"`
	StartedAt string            `json:"startedAt,omitempty"`
}

func (a *App) ImportHAR(raw string) ([]HAREntryDTO, error) {
	parsed, err := httpclient.ParseHAR(raw)
	if err != nil {
		return nil, err
	}
	out := make([]HAREntryDTO, 0, len(parsed))
	for _, e := range parsed {
		out = append(out, HAREntryDTO{
			Method: e.Method, URL: e.URL, BaseURL: e.BaseURL, Path: e.Path,
			Headers: e.Headers, Body: e.Body, Query: e.Query, Status: e.Status,
			Size: e.Size, StartedAt: e.StartedAt,
		})
	}
	return out, nil
}

type CurlImportDTO struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	BaseURL string            `json:"baseURL"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body,omitempty"`
	Query   map[string]string `json:"query,omitempty"`
}

func (a *App) ImportCurl(raw string) (*CurlImportDTO, error) {
	parsed, err := httpclient.ParseCurl(raw)
	if err != nil || parsed == nil {
		return nil, err
	}
	return &CurlImportDTO{
		Method:  parsed.Method,
		URL:     parsed.URL,
		BaseURL: parsed.BaseURL,
		Path:    parsed.Path,
		Headers: parsed.Headers,
		Body:    parsed.Body,
		Query:   parsed.Query,
	}, nil
}

func (a *App) ExportCurl(method, fullURL string, headers map[string]string, body string) string {
	return httpclient.FormatCurl(method, fullURL, headers, body)
}

func (a *App) PickFile() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select file",
	})
}

func (a *App) SaveResponseToFile(method, path, body string) (string, error) {
	if body == "" {
		return "", fmt.Errorf("empty body")
	}
	defaultName := "response"
	if path != "" {
		safe := strings.ReplaceAll(strings.Trim(path, "/"), "/", "_")
		safe = strings.ReplaceAll(safe, "{", "")
		safe = strings.ReplaceAll(safe, "}", "")
		if safe != "" {
			defaultName = safe
		}
	}
	if method != "" {
		defaultName = strings.ToLower(method) + "_" + defaultName
	}
	defaultName += ".json"
	target, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save response",
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON", Pattern: "*.json"},
			{DisplayName: "All", Pattern: "*"},
		},
	})
	if err != nil || target == "" {
		return "", err
	}
	if err := os.WriteFile(target, []byte(body), 0644); err != nil {
		return "", err
	}
	return target, nil
}

package app

import (
	"fmt"
	"spectra-desktop/internal/assertions"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/httpclient"
	"strings"
)

func topNBy(items []EndpointMetricDTO, less func(a, b EndpointMetricDTO) bool, n int) []EndpointMetricDTO {
	out := append([]EndpointMetricDTO(nil), items...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if less(out[j], out[i]) {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if n > 0 && len(out) > n {
		out = out[:n]
	}
	return out
}

func endpointTestKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func (a *App) runTestsForRequest(input ExecuteRequestInput, resp *httpclient.Response) []assertions.Result {
	if input.ProjectID == "" || input.Method == "" || input.Path == "" {
		return nil
	}
	key := endpointTestKey(input.Method, input.Path)
	tests, err := a.tests.List(a.ctx, input.ProjectID, key)
	if err != nil || len(tests) == 0 {
		return nil
	}
	domainTests := make([]assertions.Test, 0, len(tests))
	for _, t := range tests {
		domainTests = append(domainTests, assertions.Test{
			ID:       t.ID,
			Name:     t.Name,
			Kind:     t.Kind,
			JSONPath: t.JSONPath,
			Op:       t.Op,
			Expected: t.Expected,
		})
	}
	return assertions.Run(domainTests, assertions.ResponseSnapshot{
		Status:     resp.Status,
		Headers:    toHTTPHeader(resp.Headers),
		Body:       resp.Body,
		DurationMs: int(resp.DurationMs),
	})
}

type EndpointTestDTO struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Kind     string `json:"kind"`
	JSONPath string `json:"jsonPath,omitempty"`
	Op       string `json:"op,omitempty"`
	Expected string `json:"expected,omitempty"`
}

type TestResultDTO struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Pass    bool   `json:"pass"`
	Message string `json:"message,omitempty"`
}

type SaveTestsInput struct {
	ProjectID   string            `json:"projectID"`
	EndpointKey string            `json:"endpointKey"`
	Tests       []EndpointTestDTO `json:"tests"`
}

func (a *App) ListEndpointTests(projectID, endpointKey string) ([]EndpointTestDTO, error) {
	if projectID == "" || endpointKey == "" {
		return []EndpointTestDTO{}, nil
	}
	rows, err := a.tests.List(a.ctx, projectID, endpointKey)
	if err != nil {
		return nil, err
	}
	out := make([]EndpointTestDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, EndpointTestDTO{
			ID:       r.ID,
			Name:     r.Name,
			Kind:     r.Kind,
			JSONPath: r.JSONPath,
			Op:       r.Op,
			Expected: r.Expected,
		})
	}
	return out, nil
}

func (a *App) SaveEndpointTests(input SaveTestsInput) error {
	if input.ProjectID == "" || input.EndpointKey == "" {
		return fmt.Errorf("project id and endpoint key required")
	}
	tests := make([]domain.EndpointTest, 0, len(input.Tests))
	for i, t := range input.Tests {
		tests = append(tests, domain.EndpointTest{
			ID:        t.ID,
			Name:      t.Name,
			Kind:      t.Kind,
			JSONPath:  t.JSONPath,
			Op:        t.Op,
			Expected:  t.Expected,
			SortOrder: i,
		})
	}
	return a.tests.Replace(a.ctx, input.ProjectID, input.EndpointKey, tests)
}

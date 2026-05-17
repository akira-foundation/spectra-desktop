package app

import (
	"fmt"
	"log"
	"net/http"
	"spectra-desktop/internal/assertions"
	"spectra-desktop/internal/httpclient"
	"strings"
	"time"
)

type ExecuteRequestInput struct {
	ProjectID  string             `json:"projectID"`
	EndpointID string             `json:"endpointID,omitempty"`
	AccountID  string             `json:"accountID,omitempty"`
	Method     string             `json:"method"`
	Path       string             `json:"path"`
	Headers    map[string]string  `json:"headers,omitempty"`
	Body       string             `json:"body,omitempty"`
	Multipart  []MultipartPartDTO `json:"multipart,omitempty"`
	BaseURL    string             `json:"baseUrl,omitempty"`
	TimeoutMs  int                `json:"timeoutMs,omitempty"`
	SkipAuth   bool               `json:"skipAuth,omitempty"`
}

type MultipartPartDTO struct {
	Name     string `json:"name"`
	Value    string `json:"value,omitempty"`
	FilePath string `json:"filePath,omitempty"`
}

func (a *App) ExecuteRequest(input ExecuteRequestInput) (*httpclient.Response, error) {
	if a.billingGate != nil {
		if err := a.billingGate.Require(a.ctx, "requests_per_day"); err != nil {
			a.emitUpsell("requests_per_day", err)
			return nil, err
		}
	}
	baseURL := strings.TrimSpace(input.BaseURL)
	if baseURL == "" && input.ProjectID != "" {
		project, err := a.projects.GetByID(a.ctx, input.ProjectID)
		if err != nil {
			return nil, err
		}
		baseURL = strings.TrimSpace(project.BaseURL)
	}
	if baseURL == "" {
		return nil, fmt.Errorf("missing base url")
	}
	vars := a.resolveEnvVars(input.ProjectID)
	// Inject account-bound vars (account.username, account.password, etc.)
	// so login bodies can template `{{account.username}}` and similar.
	if !input.SkipAuth && input.ProjectID != "" {
		a.injectAccountVars(vars, input.ProjectID, input.AccountID)
	}
	resolvedPath := substituteVars(input.Path, vars)
	target, err := joinURL(baseURL, resolvedPath)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(input.TimeoutMs) * time.Millisecond

	headers := substituteHeaderVars(input.Headers, vars)
	body := substituteVars(input.Body, vars)
	var cookies []http.Cookie
	var queryParams map[string]string
	if !input.SkipAuth && input.ProjectID != "" {
		merged, ck, qp := a.applyProjectAuth(input.ProjectID, input.AccountID, headers)
		headers = merged
		cookies = ck
		queryParams = qp
	}
	if len(queryParams) > 0 {
		target = appendQueryParams(target, queryParams)
	}

	if len(input.Multipart) > 0 {
		parts := make([]httpclient.MultipartPart, 0, len(input.Multipart))
		for _, p := range input.Multipart {
			parts = append(parts, httpclient.MultipartPart{
				Name: p.Name, Value: p.Value, FilePath: p.FilePath,
			})
		}
		mpBody, ct, mpErr := httpclient.BuildMultipart(parts)
		if mpErr != nil {
			return nil, mpErr
		}
		body = mpBody
		if headers == nil {
			headers = map[string]string{}
		}
		headers["Content-Type"] = ct
	}
	resp, sendErr := a.http.Send(a.ctx, httpclient.Request{
		Method:  input.Method,
		URL:     target,
		Headers: headers,
		Body:    body,
		Cookies: cookies,
		Timeout: timeout,
	})

	var testResults []assertions.Result
	if input.ProjectID != "" && resp != nil && sendErr == nil {
		testResults = a.runTestsForRequest(input, resp)
		a.runCapturesForRequest(input, resp)
	}

	if a.usage != nil && sendErr == nil {
		_ = a.usage.Track(a.ctx, "requests_per_day", 1)
	}

	if input.ProjectID != "" {
		a.saveHistory(input, target, headers, resp, sendErr, testResults)
	}

	if sendErr != nil {
		return resp, sendErr
	}

	if input.ProjectID != "" && input.EndpointID != "" {
		if a.isLogoutEndpoint(input.ProjectID, input.EndpointID) && resp.Status < 400 {
			if err := a.auth.Clear(a.ctx, input.ProjectID); err != nil {
				log.Printf("clear auth on logout: %v", err)
			}
		} else {
			a.captureAuthFromResponse(input.ProjectID, input.EndpointID, input.AccountID, resp)
		}
	}
	return resp, nil
}

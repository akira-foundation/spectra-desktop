package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/mock"
)

type MockStatusDTO struct {
	Running      bool      `json:"running"`
	ProjectID    string    `json:"projectId,omitempty"`
	Port         int       `json:"port,omitempty"`
	URL          string    `json:"url,omitempty"`
	StartedAt    time.Time `json:"startedAt,omitempty"`
	RequestCount int64     `json:"requestCount"`
}

type MockOverrideDTO struct {
	ID         string            `json:"id"`
	ProjectID  string            `json:"projectID"`
	EndpointID string            `json:"endpointId"`
	Enabled    bool              `json:"enabled"`
	Status     int               `json:"status"`
	LatencyMs  int               `json:"latencyMs"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
	Source     string            `json:"source"`
	UpdatedAt  time.Time         `json:"updatedAt"`
}

type SaveMockOverrideInput struct {
	ID         string            `json:"id,omitempty"`
	ProjectID  string            `json:"projectID"`
	EndpointID string            `json:"endpointId"`
	Enabled    bool              `json:"enabled"`
	Status     int               `json:"status,omitempty"`
	LatencyMs  int               `json:"latencyMs,omitempty"`
	Body       string            `json:"body,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Source     string            `json:"source,omitempty"`
}

func (a *App) StartMockServer(projectID string, port int) (*MockStatusDTO, error) {
	if a.mock == nil {
		return nil, fmt.Errorf("mock subsystem not initialized")
	}
	if a.billingGate != nil {
		if err := a.billingGate.Require(a.ctx, "mock_server"); err != nil {
			a.emitUpsell("mock_server", err)
			return nil, err
		}
	}
	status, err := a.mock.Start(a.ctx, projectID, port)
	if err != nil {
		return nil, err
	}
	return toMockStatusDTO(status), nil
}

func (a *App) StopMockServer() error {
	if a.mock == nil {
		return nil
	}
	return a.mock.Stop()
}

func (a *App) MockServerStatus() *MockStatusDTO {
	if a.mock == nil {
		return &MockStatusDTO{Running: false}
	}
	status := a.mock.Status()
	return toMockStatusDTO(status)
}

func (a *App) ListMockOverrides(projectID string) ([]MockOverrideDTO, error) {
	if a.mockRepo == nil {
		return []MockOverrideDTO{}, nil
	}
	rows, err := a.mockRepo.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]MockOverrideDTO, len(rows))
	for i, r := range rows {
		out[i] = toMockOverrideDTO(r)
	}
	return out, nil
}

func (a *App) SaveMockOverride(input SaveMockOverrideInput) (*MockOverrideDTO, error) {
	if a.mockRepo == nil {
		return nil, fmt.Errorf("mock subsystem not initialized")
	}
	if input.ProjectID == "" || input.EndpointID == "" {
		return nil, fmt.Errorf("projectID and endpointId required")
	}
	id := strings.TrimSpace(input.ID)
	if id == "" {
		if existing, _ := a.mockRepo.Get(a.ctx, input.ProjectID, input.EndpointID); existing != nil {
			id = existing.ID
		} else {
			id = uuid.NewString()
		}
	}
	source := strings.TrimSpace(input.Source)
	if source == "" {
		source = string(domain.MockSourceAuto)
	}
	override := domain.MockOverride{
		ID:         id,
		ProjectID:  input.ProjectID,
		EndpointID: input.EndpointID,
		Enabled:    input.Enabled,
		Status:     input.Status,
		LatencyMs:  input.LatencyMs,
		Body:       input.Body,
		Source:     domain.MockSource(source),
	}
	if len(input.Headers) > 0 {
		raw, err := json.Marshal(input.Headers)
		if err != nil {
			return nil, err
		}
		override.HeadersJSON = string(raw)
	}
	if err := a.mockRepo.Save(a.ctx, override); err != nil {
		return nil, err
	}
	saved, err := a.mockRepo.Get(a.ctx, input.ProjectID, input.EndpointID)
	if err != nil || saved == nil {
		return nil, err
	}
	dto := toMockOverrideDTO(*saved)
	return &dto, nil
}

func (a *App) DeleteMockOverride(id string) error {
	if a.mockRepo == nil {
		return nil
	}
	return a.mockRepo.Delete(a.ctx, id)
}

func toMockStatusDTO(status mock.Status) *MockStatusDTO {
	return &MockStatusDTO{
		Running:      status.Running,
		ProjectID:    status.ProjectID,
		Port:         status.Port,
		URL:          status.URL,
		StartedAt:    status.StartedAt,
		RequestCount: status.RequestCount,
	}
}

func toMockOverrideDTO(o domain.MockOverride) MockOverrideDTO {
	dto := MockOverrideDTO{
		ID:         o.ID,
		ProjectID:  o.ProjectID,
		EndpointID: o.EndpointID,
		Enabled:    o.Enabled,
		Status:     o.Status,
		LatencyMs:  o.LatencyMs,
		Body:       o.Body,
		Source:     string(o.Source),
		UpdatedAt:  o.UpdatedAt,
	}
	if o.HeadersJSON != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(o.HeadersJSON), &headers); err == nil {
			dto.Headers = headers
		}
	}
	return dto
}

func (a *App) newWailsEventEmitter() mock.EventEmitter {
	return func(name string, data ...any) {
		if a.ctx == nil {
			return
		}
		runtime.EventsEmit(a.ctx, name, data...)
	}
}

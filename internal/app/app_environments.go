package app

import (
	"fmt"
	"spectra-desktop/internal/domain"
)

type EnvironmentDTO struct {
	ID        string            `json:"id"`
	ProjectID string            `json:"projectID"`
	Name      string            `json:"name"`
	Vars      map[string]string `json:"vars"`
	SortOrder int               `json:"sortOrder"`
}

type SaveEnvironmentInput struct {
	ID        string            `json:"id,omitempty"`
	ProjectID string            `json:"projectID"`
	Name      string            `json:"name"`
	Vars      map[string]string `json:"vars,omitempty"`
	SortOrder int               `json:"sortOrder,omitempty"`
}

func (a *App) ListEnvironments(projectID string) ([]EnvironmentDTO, error) {
	if projectID == "" {
		return []EnvironmentDTO{}, nil
	}
	envs, err := a.envs.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]EnvironmentDTO, 0, len(envs))
	for _, e := range envs {
		out = append(out, envToDTO(e))
	}
	return out, nil
}

func (a *App) SaveEnvironment(input SaveEnvironmentInput) (*EnvironmentDTO, error) {
	if input.ProjectID == "" {
		return nil, fmt.Errorf("project id required")
	}
	if input.Vars == nil {
		input.Vars = map[string]string{}
	}
	env, err := a.envs.Save(a.ctx, domain.EnvironmentInput{
		ID:        input.ID,
		ProjectID: input.ProjectID,
		Name:      input.Name,
		Vars:      input.Vars,
		SortOrder: input.SortOrder,
	})
	if err != nil || env == nil {
		return nil, err
	}
	dto := envToDTO(*env)
	return &dto, nil
}

func (a *App) DeleteEnvironment(id string) error {
	return a.envs.Delete(a.ctx, id)
}

func (a *App) SetActiveEnvironment(projectID, envID string) error {
	return a.projects.UpdateActiveEnvironment(a.ctx, projectID, envID)
}
func envToDTO(e domain.Environment) EnvironmentDTO {
	return EnvironmentDTO{
		ID:        e.ID,
		ProjectID: e.ProjectID,
		Name:      e.Name,
		Vars:      e.Vars,
		SortOrder: e.SortOrder,
	}
}

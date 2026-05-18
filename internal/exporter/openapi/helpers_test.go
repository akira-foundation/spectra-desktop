package openapi

import "spectra-desktop/internal/domain"

func newProject() *domain.Project {
	return &domain.Project{
		ID:   "p1",
		Name: "Demo API",
	}
}

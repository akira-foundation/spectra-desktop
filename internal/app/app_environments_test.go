package app

import (
	"testing"

	"spectra-desktop/internal/domain"
)

func TestEnvToDTO_CopiesFields(t *testing.T) {
	e := domain.Environment{
		ID:        "e1",
		ProjectID: "p1",
		Name:      "Local",
		Vars:      map[string]string{"K": "V"},
		SortOrder: 7,
	}
	dto := envToDTO(e)
	if dto.ID != "e1" || dto.ProjectID != "p1" || dto.Name != "Local" || dto.SortOrder != 7 {
		t.Fatalf("scalars: %+v", dto)
	}
	if dto.Vars["K"] != "V" {
		t.Fatalf("vars: %v", dto.Vars)
	}
}

func TestEnvToDTO_NilVarsRemainsNil(t *testing.T) {
	dto := envToDTO(domain.Environment{ID: "x"})
	if dto.Vars != nil {
		t.Fatalf("expected nil vars, got %v", dto.Vars)
	}
}

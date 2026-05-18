package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

func TestModel_ProjectRoundTrip(t *testing.T) {
	s := newStorage(t)
	ctx := context.Background()
	now := time.Now().UTC()

	in := model.Project{
		ID:        uuid.NewString(),
		Name:      "x",
		Path:      "/p",
		Framework: "laravel",
		Status:    string(domain.ProjectStatusConnected),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := s.DB.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}

	var out model.Project
	if err := s.DB.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Name != in.Name || out.Path != in.Path || out.Status != in.Status {
		t.Fatalf("mismatch: %+v vs %+v", out, in)
	}
}

func TestModel_EndpointRoundTripViaHelpers(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewEndpointRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "rt")
	in := []core.Endpoint{{
		Method:     core.MethodPut,
		Path:       "/things/{id}",
		Handler:    "ThingController@update",
		Middleware: []string{"auth"},
		Parameters: []core.Parameter{{Name: "id", In: "path", Required: true}},
		Tags:       []string{"things"},
	}}
	if err := repo.Replace(ctx, p.ID, in); err != nil {
		t.Fatalf("replace: %v", err)
	}

	out, _ := repo.List(ctx, p.ID)
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
	if out[0].Method != core.MethodPut || out[0].Path != "/things/{id}" {
		t.Fatalf("mismatch: %+v", out[0])
	}
	if len(out[0].Parameters) != 1 || out[0].Parameters[0].Name != "id" {
		t.Fatalf("parameters round-trip failed: %+v", out[0].Parameters)
	}
}

func TestModel_RequestHistoryRoundTrip(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "rt-h")
	in := model.RequestHistory{
		ID:             uuid.NewString(),
		ProjectID:      p.ID,
		EndpointID:     "ep",
		Method:         "GET",
		URL:            "/x",
		ResponseStatus: 200,
		CreatedAt:      time.Now().UTC(),
	}
	if _, err := s.DB.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.RequestHistory
	if err := s.DB.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.URL != in.URL || out.ResponseStatus != in.ResponseStatus {
		t.Fatalf("mismatch: %+v vs %+v", out, in)
	}
}

func TestModel_MockOverrideRoundTrip(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "rt-m")
	in := model.MockOverride{
		ID:         uuid.NewString(),
		ProjectID:  p.ID,
		EndpointID: "ep",
		Enabled:    true,
		Status:     200,
		Body:       "{}",
		Source:     string(domain.MockSourceAuto),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if _, err := s.DB.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.MockOverride
	if err := s.DB.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Body != "{}" || !out.Enabled {
		t.Fatalf("mismatch: %+v", out)
	}
}

package model_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/repository/model"
)

func TestRoundTrip_CapturedValue(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	in := model.CapturedValue{
		ProjectID:   pid,
		Name:        "token",
		Value:       "abc",
		EndpointKey: "ep",
		CapturedAt:  time.Now().UTC().Truncate(time.Second),
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.CapturedValue
	if err := db.NewSelect().Model(&out).Where("project_id = ? AND name = ?", pid, "token").Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Value != "abc" || out.EndpointKey != "ep" {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_Collection(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	c := model.Collection{
		ID: uuid.NewString(), ProjectID: pid, Name: "c", Description: "d",
		SortOrder: 1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&c).Exec(ctx); err != nil {
		t.Fatalf("insert collection: %v", err)
	}
	item := model.CollectionItem{
		ID: uuid.NewString(), CollectionID: c.ID, EndpointID: "ep",
		SortOrder: 1, SkipOnFailure: 1, IterateDataset: 0,
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&item).Exec(ctx); err != nil {
		t.Fatalf("insert item: %v", err)
	}
	var out model.CollectionItem
	if err := db.NewSelect().Model(&out).Where("id = ?", item.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.CollectionID != c.ID || out.SkipOnFailure != 1 {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_Environment(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.Environment{
		ID: uuid.NewString(), ProjectID: pid, Name: "dev",
		VarsJSON: `{"K":"V"}`, SortOrder: 1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.Environment
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.VarsJSON != `{"K":"V"}` || out.Name != "dev" {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_MockOverride(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.MockOverride{
		ID: uuid.NewString(), ProjectID: pid, EndpointID: "ep",
		Enabled: true, Status: 201, LatencyMs: 10, Body: "{}",
		Source: "auto", CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.MockOverride
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if !out.Enabled || out.Status != 201 || out.LatencyMs != 10 {
		t.Fatalf("mismatch: %+v", out)
	}
}

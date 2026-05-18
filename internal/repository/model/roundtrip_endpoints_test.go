package model_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/repository/model"
)

func TestRoundTrip_EndpointCapture(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.EndpointCapture{
		ID:          uuid.NewString(),
		ProjectID:   pid,
		EndpointKey: "k",
		Name:        "n",
		Source:      "json",
		Path:        "data.id",
		SortOrder:   2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.EndpointCapture
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Name != in.Name || out.Source != in.Source || out.Path != in.Path || out.SortOrder != in.SortOrder {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_EndpointDataset(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.EndpointDataset{
		ProjectID: pid, EndpointKey: "ep", RowsJSON: `[{"a":1}]`,
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.EndpointDataset
	if err := db.NewSelect().Model(&out).Where("project_id = ? AND endpoint_key = ?", pid, "ep").Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.RowsJSON != in.RowsJSON {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_EndpointSnapshot(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.EndpointSnapshot{
		ID: uuid.NewString(), ProjectID: pid, Hash: "h",
		PayloadJSON: "[]", EndpointCount: 3, ScannedAt: now, CreatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.EndpointSnapshot
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Hash != "h" || out.EndpointCount != 3 {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_EndpointTest(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.EndpointTest{
		ID: uuid.NewString(), ProjectID: pid, EndpointKey: "ep",
		Name: "n", Kind: "status", JSONPath: "$.x", Op: "eq", Expected: "1",
		SortOrder: 1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.EndpointTest
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Kind != "status" || out.JSONPath != "$.x" || out.Op != "eq" {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_Endpoint(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.Endpoint{
		ID: uuid.NewString(), ProjectID: pid, Method: "GET", Path: "/x",
		Middleware: `["auth"]`, Parameters: "[]", Tags: "[]",
		ScannedAt: now, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.Endpoint
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Method != "GET" || out.Path != "/x" || out.Middleware != `["auth"]` {
		t.Fatalf("mismatch: %+v", out)
	}
}

package model_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"spectra-desktop/internal/repository/model"
)

func TestRoundTrip_License(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	in := model.License{
		ID: "local", CustomerEmail: "a@b.c", Plan: "pro", Status: "active",
		CancelAtPeriodEnd: true, GracePeriod: false, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).On("CONFLICT (id) DO UPDATE").Exec(ctx); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	var out model.License
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.CustomerEmail != "a@b.c" || !out.CancelAtPeriodEnd || out.Plan != "pro" {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_UsageBufferEntry(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	in := model.UsageBufferEntry{
		ID: uuid.NewString(), Feature: "scans", Amount: 2,
		OccurredAt: now, Flushed: false, CreatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.UsageBufferEntry
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Feature != "scans" || out.Amount != 2 || out.Flushed {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_ProjectAccount_NilExpires(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.ProjectAccount{
		ID: uuid.NewString(), ProjectID: pid, Label: "default",
		Kind: "bearer", APIKeyIn: "header", IsDefault: true,
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.ProjectAccount
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if !out.IsDefault || out.Kind != "bearer" || out.ExpiresAt != nil {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_ProjectAccount_WithExpires(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	exp := now.Add(time.Hour)
	in := model.ProjectAccount{
		ID: uuid.NewString(), ProjectID: pid, Label: "x",
		Kind: "oauth2", APIKeyIn: "header", ExpiresAt: &exp,
		CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.ProjectAccount
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.ExpiresAt == nil || !out.ExpiresAt.Equal(exp) {
		t.Fatalf("expires mismatch: %+v", out.ExpiresAt)
	}
}

func TestRoundTrip_ProjectAuth(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.ProjectAuth{
		ProjectID: pid, Scheme: "bearer", Token: "tok",
		TokenPath: "data.token", CapturedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.ProjectAuth
	if err := db.NewSelect().Model(&out).Where("project_id = ?", pid).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Scheme != "bearer" || out.Token != "tok" || out.ExpiresAt != nil {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_RequestHistory(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	in := model.RequestHistory{
		ID: uuid.NewString(), ProjectID: pid, Method: "GET", URL: "/x",
		ResponseStatus: 200, CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.RequestHistory
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.URL != "/x" || out.ResponseStatus != 200 {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_ScratchRequest(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	pid := insertProject(t, db)
	now := time.Now().UTC().Truncate(time.Second)
	in := model.ScratchRequest{
		ID: uuid.NewString(), ProjectID: pid, Name: "n",
		Method: "POST", URL: "/x", HeadersJSON: "{}", Body: "{}",
		SortOrder: 1, CreatedAt: now, UpdatedAt: now,
	}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.ScratchRequest
	if err := db.NewSelect().Model(&out).Where("id = ?", in.ID).Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Method != "POST" || out.URL != "/x" {
		t.Fatalf("mismatch: %+v", out)
	}
}

func TestRoundTrip_Setting(t *testing.T) {
	db := newDB(t)
	ctx := context.Background()
	in := model.Setting{Key: "k", Value: "v", UpdatedAt: time.Now().UTC().Truncate(time.Second)}
	if _, err := db.NewInsert().Model(&in).Exec(ctx); err != nil {
		t.Fatalf("insert: %v", err)
	}
	var out model.Setting
	if err := db.NewSelect().Model(&out).Where("key = ?", "k").Scan(ctx); err != nil {
		t.Fatalf("select: %v", err)
	}
	if out.Value != "v" {
		t.Fatalf("mismatch: %+v", out)
	}
}

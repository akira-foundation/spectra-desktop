package repository

import (
	"context"
	"testing"

	"spectra-desktop/internal/domain"
)

func TestCollectionRepository_CreateGetList(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCollectionRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "col")
	c, err := repo.Create(ctx, domain.Collection{ProjectID: p.ID, Name: "Suite", Description: "d"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.ID == "" || c.CreatedAt.IsZero() {
		t.Fatalf("expected populated created collection: %+v", c)
	}
	got, err := repo.Get(ctx, c.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil || got.Name != "Suite" {
		t.Fatalf("unexpected: %+v", got)
	}
	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("want 1, got %d", len(list))
	}
}

func TestCollectionRepository_Get_MissingReturnsNil(t *testing.T) {
	s := newStorage(t)
	repo := NewCollectionRepository(s.DB)
	got, err := repo.Get(context.Background(), "nope")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestCollectionRepository_Update(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCollectionRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "col")
	c, _ := repo.Create(ctx, domain.Collection{ProjectID: p.ID, Name: "Old"})
	c.Name = "New"
	c.Description = "x"
	c.SortOrder = 7
	if err := repo.Update(ctx, *c); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, _ := repo.Get(ctx, c.ID)
	if got == nil || got.Name != "New" || got.Description != "x" || got.SortOrder != 7 {
		t.Fatalf("update lost: %+v", got)
	}
}

func TestCollectionRepository_Delete(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCollectionRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "col")
	c, _ := repo.Create(ctx, domain.Collection{ProjectID: p.ID, Name: "x"})
	if err := repo.Delete(ctx, c.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := repo.Get(ctx, c.ID)
	if got != nil {
		t.Fatalf("expected gone, got %+v", got)
	}
}

func TestCollectionRepository_ReplaceItems(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCollectionRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "col")
	c, _ := repo.Create(ctx, domain.Collection{ProjectID: p.ID, Name: "x"})

	items := []domain.CollectionItem{
		{EndpointID: "e1", BodyOverride: "{}", SkipOnFailure: true},
		{EndpointID: "e2", IterateDataset: true},
	}
	if err := repo.ReplaceItems(ctx, c.ID, items); err != nil {
		t.Fatalf("replace: %v", err)
	}
	got, _ := repo.Get(ctx, c.ID)
	if len(got.Items) != 2 {
		t.Fatalf("want 2 items, got %d", len(got.Items))
	}
	if !got.Items[0].SkipOnFailure || got.Items[0].SortOrder != 0 {
		t.Fatalf("first item flags: %+v", got.Items[0])
	}
	if !got.Items[1].IterateDataset || got.Items[1].SortOrder != 1 {
		t.Fatalf("second item flags: %+v", got.Items[1])
	}

	if err := repo.ReplaceItems(ctx, c.ID, nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	got, _ = repo.Get(ctx, c.ID)
	if len(got.Items) != 0 {
		t.Fatalf("expected items cleared, got %d", len(got.Items))
	}
}

func TestCollectionRepository_ListWithItems(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCollectionRepository(s.DB)
	ctx := context.Background()

	p := seedProject(t, projects, "col")
	c1, _ := repo.Create(ctx, domain.Collection{ProjectID: p.ID, Name: "a", SortOrder: 0})
	c2, _ := repo.Create(ctx, domain.Collection{ProjectID: p.ID, Name: "b", SortOrder: 1})
	if err := repo.ReplaceItems(ctx, c1.ID, []domain.CollectionItem{{EndpointID: "e1"}}); err != nil {
		t.Fatalf("items c1: %v", err)
	}
	if err := repo.ReplaceItems(ctx, c2.ID, []domain.CollectionItem{{EndpointID: "e2"}, {EndpointID: "e3"}}); err != nil {
		t.Fatalf("items c2: %v", err)
	}
	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || len(list[0].Items) != 1 || len(list[1].Items) != 2 {
		t.Fatalf("aggregated items wrong: %+v", list)
	}
}

func TestCollectionRepository_List_EmptyProject(t *testing.T) {
	s := newStorage(t)
	projects := NewProjectRepository(s.DB)
	repo := NewCollectionRepository(s.DB)
	ctx := context.Background()
	p := seedProject(t, projects, "col")
	list, err := repo.List(ctx, p.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if list == nil || len(list) != 0 {
		t.Fatalf("expected empty non-nil slice, got %+v", list)
	}
}

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

type CollectionRepository struct {
	db *bun.DB
}

func NewCollectionRepository(db *bun.DB) *CollectionRepository {
	return &CollectionRepository{db: db}
}

var _ domain.CollectionRepository = (*CollectionRepository)(nil)

func (r *CollectionRepository) List(ctx context.Context, projectID string) ([]domain.Collection, error) {
	var rows []model.Collection
	err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		OrderExpr("sort_order ASC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []domain.Collection{}, nil
	}
	ids := make([]string, 0, len(rows))
	for _, c := range rows {
		ids = append(ids, c.ID)
	}
	var items []model.CollectionItem
	if err := r.db.NewSelect().
		Model(&items).
		Where("collection_id IN (?)", bun.In(ids)).
		OrderExpr("collection_id, sort_order ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	itemsByColl := map[string][]domain.CollectionItem{}
	for _, it := range items {
		itemsByColl[it.CollectionID] = append(itemsByColl[it.CollectionID], toCollectionItemDomain(it))
	}
	out := make([]domain.Collection, 0, len(rows))
	for _, c := range rows {
		dc := toCollectionDomain(c)
		dc.Items = itemsByColl[c.ID]
		out = append(out, dc)
	}
	return out, nil
}

func (r *CollectionRepository) Get(ctx context.Context, id string) (*domain.Collection, error) {
	row := new(model.Collection)
	err := r.db.NewSelect().Model(row).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	var items []model.CollectionItem
	if err := r.db.NewSelect().
		Model(&items).
		Where("collection_id = ?", id).
		OrderExpr("sort_order ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	dc := toCollectionDomain(*row)
	for _, it := range items {
		dc.Items = append(dc.Items, toCollectionItemDomain(it))
	}
	return &dc, nil
}

func (r *CollectionRepository) Create(ctx context.Context, c domain.Collection) (*domain.Collection, error) {
	now := time.Now().UTC()
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	row := model.Collection{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		Name:        c.Name,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := r.db.NewInsert().Model(&row).Exec(ctx); err != nil {
		return nil, err
	}
	c.CreatedAt = now
	c.UpdatedAt = now
	return &c, nil
}

func (r *CollectionRepository) Update(ctx context.Context, c domain.Collection) error {
	_, err := r.db.NewUpdate().
		Model((*model.Collection)(nil)).
		Set("name = ?", c.Name).
		Set("description = ?", c.Description).
		Set("sort_order = ?", c.SortOrder).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", c.ID).
		Exec(ctx)
	return err
}

func (r *CollectionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Collection)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *CollectionRepository) ReplaceItems(ctx context.Context, collectionID string, items []domain.CollectionItem) error {
	now := time.Now().UTC()
	return r.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().
			Model((*model.CollectionItem)(nil)).
			Where("collection_id = ?", collectionID).
			Exec(ctx); err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		rows := make([]model.CollectionItem, 0, len(items))
		for i, it := range items {
			id := it.ID
			if id == "" {
				id = uuid.NewString()
			}
			skip := 0
			if it.SkipOnFailure {
				skip = 1
			}
			rows = append(rows, model.CollectionItem{
				ID:              id,
				CollectionID:    collectionID,
				EndpointID:      it.EndpointID,
				SortOrder:       i,
				BodyOverride:    it.BodyOverride,
				HeadersOverride: it.HeadersOverride,
				SkipOnFailure:   skip,
				CreatedAt:       now,
				UpdatedAt:       now,
			})
		}
		_, err := tx.NewInsert().Model(&rows).Exec(ctx)
		return err
	})
}

func toCollectionDomain(row model.Collection) domain.Collection {
	return domain.Collection{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		Name:        row.Name,
		Description: row.Description,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toCollectionItemDomain(row model.CollectionItem) domain.CollectionItem {
	return domain.CollectionItem{
		ID:              row.ID,
		CollectionID:    row.CollectionID,
		EndpointID:      row.EndpointID,
		SortOrder:       row.SortOrder,
		BodyOverride:    row.BodyOverride,
		HeadersOverride: row.HeadersOverride,
		SkipOnFailure:   row.SkipOnFailure != 0,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

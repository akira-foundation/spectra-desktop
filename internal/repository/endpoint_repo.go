package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

type EndpointRepository struct {
	db *bun.DB
}

func NewEndpointRepository(db *bun.DB) *EndpointRepository {
	return &EndpointRepository{db: db}
}

var _ domain.EndpointRepository = (*EndpointRepository)(nil)

func (r *EndpointRepository) List(ctx context.Context, projectID string) ([]core.Endpoint, error) {
	var rows []model.Endpoint
	if err := r.db.NewSelect().
		Model(&rows).
		Where("project_id = ?", projectID).
		OrderExpr("path ASC, method ASC").
		Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]core.Endpoint, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToCore())
	}
	return out, nil
}

func (r *EndpointRepository) GetByID(ctx context.Context, id string) (*core.Endpoint, error) {
	var row model.Endpoint
	err := r.db.NewSelect().Model(&row).Where("id = ?", id).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ep := row.ToCore()
	return &ep, nil
}

func (r *EndpointRepository) Replace(ctx context.Context, projectID string, endpoints []core.Endpoint) error {
	now := time.Now().UTC()
	return r.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var existing []model.Endpoint
		if err := tx.NewSelect().
			Model(&existing).
			Column("method", "path", "auth_role_override", "token_path_override").
			Where("project_id = ?", projectID).
			Scan(ctx); err != nil {
			return err
		}
		overrides := make(map[string]struct {
			Role      string
			TokenPath string
		}, len(existing))
		for _, e := range existing {
			if e.AuthRoleOverride == "" && e.TokenPathOverride == "" {
				continue
			}
			overrides[e.Method+" "+e.Path] = struct {
				Role      string
				TokenPath string
			}{e.AuthRoleOverride, e.TokenPathOverride}
		}

		if _, err := tx.NewDelete().
			Model((*model.Endpoint)(nil)).
			Where("project_id = ?", projectID).
			Exec(ctx); err != nil {
			return err
		}
		if len(endpoints) == 0 {
			return nil
		}
		rows := make([]model.Endpoint, 0, len(endpoints))
		for i := range endpoints {
			endpoints[i].ID = projectID + "#" + strconv.Itoa(i)
			if o, ok := overrides[string(endpoints[i].Method)+" "+endpoints[i].Path]; ok {
				endpoints[i].AuthRoleOverride = core.AuthRole(o.Role)
				endpoints[i].TokenPathOverride = o.TokenPath
			}
			rows = append(rows, model.EndpointFromCore(projectID, endpoints[i], now, now))
		}
		_, err := tx.NewInsert().Model(&rows).Exec(ctx)
		return err
	})
}

func (r *EndpointRepository) UpdateAuthOverride(ctx context.Context, endpointID string, role core.AuthRole, tokenPath string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Endpoint)(nil)).
		Set("auth_role_override = ?", string(role)).
		Set("token_path_override = ?", tokenPath).
		Set("updated_at = ?", time.Now().UTC()).
		Where("id = ?", endpointID).
		Exec(ctx)
	return err
}

func (r *EndpointRepository) DeleteByProject(ctx context.Context, projectID string) error {
	_, err := r.db.NewDelete().
		Model((*model.Endpoint)(nil)).
		Where("project_id = ?", projectID).
		Exec(ctx)
	return err
}

func (r *EndpointRepository) Stats(ctx context.Context, projectID string) (domain.ProjectStats, error) {
	var rows []model.Endpoint
	if err := r.db.NewSelect().
		Model(&rows).
		Column("handler", "middleware", "scanned_at").
		Where("project_id = ?", projectID).
		Scan(ctx); err != nil {
		return domain.ProjectStats{}, err
	}

	stats := domain.ProjectStats{Routes: len(rows)}
	if len(rows) == 0 {
		return stats, nil
	}

	controllers := make(map[string]struct{})
	middleware := make(map[string]struct{})
	var latest time.Time
	for _, row := range rows {
		if row.Handler != "" {
			controllers[normalizeController(row.Handler)] = struct{}{}
		}
		var mws []string
		if err := json.Unmarshal([]byte(row.Middleware), &mws); err == nil {
			for _, m := range mws {
				middleware[m] = struct{}{}
			}
		}
		if row.ScannedAt.After(latest) {
			latest = row.ScannedAt
		}
	}
	stats.Controllers = len(controllers)
	stats.Middleware = len(middleware)
	if !latest.IsZero() {
		stats.LastScannedAt = &latest
	}
	return stats, nil
}

func normalizeController(handler string) string {
	for i := 0; i < len(handler); i++ {
		if handler[i] == '@' {
			return handler[:i]
		}
	}
	return handler
}

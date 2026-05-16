package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/domain"
	"spectra-desktop/internal/repository/model"
)

const licenseRowID = "local"

type LicenseRepository struct {
	db *bun.DB
}

func NewLicenseRepository(db *bun.DB) *LicenseRepository {
	return &LicenseRepository{db: db}
}

var _ domain.LicenseRepository = (*LicenseRepository)(nil)

func (r *LicenseRepository) Get(ctx context.Context) (*domain.License, error) {
	var row model.License
	err := r.db.NewSelect().Model(&row).Where("id = ?", licenseRowID).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return licenseToDomain(row), nil
}

func (r *LicenseRepository) Save(ctx context.Context, license domain.License) error {
	row := licenseToModel(license)
	row.ID = licenseRowID
	row.UpdatedAt = time.Now().UTC()
	_, err := r.db.NewInsert().
		Model(&row).
		On("CONFLICT (id) DO UPDATE").
		Set("customer_id = EXCLUDED.customer_id").
		Set("customer_email = EXCLUDED.customer_email").
		Set("customer_name = EXCLUDED.customer_name").
		Set("access_token_enc = EXCLUDED.access_token_enc").
		Set("plan = EXCLUDED.plan").
		Set("cycle = EXCLUDED.cycle").
		Set("status = EXCLUDED.status").
		Set("valid_until = EXCLUDED.valid_until").
		Set("activated_at = EXCLUDED.activated_at").
		Set("last_verified_at = EXCLUDED.last_verified_at").
		Set("license_key_id = EXCLUDED.license_key_id").
		Set("license_algorithm = EXCLUDED.license_algorithm").
		Set("license_payload = EXCLUDED.license_payload").
		Set("license_signature = EXCLUDED.license_signature").
		Set("features_json = EXCLUDED.features_json").
		Set("device_id = EXCLUDED.device_id").
		Set("cancel_at_period_end = EXCLUDED.cancel_at_period_end").
		Set("cancel_at = EXCLUDED.cancel_at").
		Set("target_plan = EXCLUDED.target_plan").
		Set("grace_period = EXCLUDED.grace_period").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}

func (r *LicenseRepository) Clear(ctx context.Context) error {
	cleared := domain.License{ID: licenseRowID, Status: "inactive", FeaturesJSON: "{}"}
	return r.Save(ctx, cleared)
}

type UsageBufferRepository struct {
	db *bun.DB
}

func NewUsageBufferRepository(db *bun.DB) *UsageBufferRepository {
	return &UsageBufferRepository{db: db}
}

var _ domain.UsageBufferRepository = (*UsageBufferRepository)(nil)

func (r *UsageBufferRepository) Append(ctx context.Context, entry domain.UsageBufferEntry) error {
	row := model.UsageBufferEntry{
		ID:         entry.ID,
		Feature:    entry.Feature,
		Amount:     entry.Amount,
		OccurredAt: entry.OccurredAt,
		Flushed:    entry.Flushed,
		CreatedAt:  time.Now().UTC(),
	}
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)
	return err
}

func (r *UsageBufferRepository) PendingBatch(ctx context.Context, limit int) ([]domain.UsageBufferEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows []model.UsageBufferEntry
	err := r.db.NewSelect().
		Model(&rows).
		Where("flushed = FALSE").
		OrderExpr("occurred_at ASC").
		Limit(limit).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.UsageBufferEntry, len(rows))
	for i, row := range rows {
		out[i] = domain.UsageBufferEntry{
			ID:         row.ID,
			Feature:    row.Feature,
			Amount:     row.Amount,
			OccurredAt: row.OccurredAt,
			Flushed:    row.Flushed,
			CreatedAt:  row.CreatedAt,
		}
	}
	return out, nil
}

func (r *UsageBufferRepository) MarkFlushed(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.NewUpdate().
		Model((*model.UsageBufferEntry)(nil)).
		Set("flushed = TRUE").
		Where("id IN (?)", bun.In(ids)).
		Exec(ctx)
	return err
}

func licenseToDomain(row model.License) *domain.License {
	return &domain.License{
		ID:                row.ID,
		CustomerID:        row.CustomerID,
		CustomerEmail:     row.CustomerEmail,
		CustomerName:      row.CustomerName,
		AccessTokenEnc:    row.AccessTokenEnc,
		Plan:              row.Plan,
		Cycle:             row.Cycle,
		Status:            row.Status,
		ValidUntil:        row.ValidUntil,
		ActivatedAt:       row.ActivatedAt,
		LastVerifiedAt:    row.LastVerifiedAt,
		LicenseKeyID:      row.LicenseKeyID,
		LicenseAlgorithm:  row.LicenseAlgorithm,
		LicensePayload:    row.LicensePayload,
		LicenseSignature:  row.LicenseSignature,
		FeaturesJSON:      row.FeaturesJSON,
		DeviceID:          row.DeviceID,
		CancelAtPeriodEnd: row.CancelAtPeriodEnd,
		CancelAt:          row.CancelAt,
		TargetPlan:        row.TargetPlan,
		GracePeriod:       row.GracePeriod,
		UpdatedAt:         row.UpdatedAt,
	}
}

func licenseToModel(l domain.License) model.License {
	return model.License{
		ID:                l.ID,
		CustomerID:        l.CustomerID,
		CustomerEmail:     l.CustomerEmail,
		CustomerName:      l.CustomerName,
		AccessTokenEnc:    l.AccessTokenEnc,
		Plan:              l.Plan,
		Cycle:             l.Cycle,
		Status:            l.Status,
		ValidUntil:        l.ValidUntil,
		ActivatedAt:       l.ActivatedAt,
		LastVerifiedAt:    l.LastVerifiedAt,
		LicenseKeyID:      l.LicenseKeyID,
		LicenseAlgorithm:  l.LicenseAlgorithm,
		LicensePayload:    l.LicensePayload,
		LicenseSignature:  l.LicenseSignature,
		FeaturesJSON:      l.FeaturesJSON,
		DeviceID:          l.DeviceID,
		CancelAtPeriodEnd: l.CancelAtPeriodEnd,
		CancelAt:          l.CancelAt,
		TargetPlan:        l.TargetPlan,
		GracePeriod:       l.GracePeriod,
		UpdatedAt:         l.UpdatedAt,
	}
}

package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
)

const offlineGraceWindow int64 = 7 * 24 * 60 * 60

type FeatureAccess struct {
	Allowed    bool
	Reason     string
	Plan       string
	Status     string
	Grace      bool
	Remaining  uint64
	Unlimited  bool
	HasFeature bool
}

type Gate struct {
	client *Client
	usage  domain.UsageBufferRepository
}

func NewGate(client *Client, usage domain.UsageBufferRepository) *Gate {
	return &Gate{client: client, usage: usage}
}

func (g *Gate) Check(ctx context.Context, feature string) (FeatureAccess, error) {
	if g.client == nil {
		return FeatureAccess{Allowed: true, Unlimited: true}, nil
	}

	license, payload, err := g.client.VerifyLocal(ctx)
	if err != nil {
		return FeatureAccess{}, err
	}
	if license == nil || payload == nil {
		return FeatureAccess{Allowed: false, Reason: "no_license"}, nil
	}

	access := FeatureAccess{
		Plan:   license.Plan,
		Status: license.Status,
		Grace:  license.GracePeriod,
	}

	now := time.Now().UTC()
	if billingsdk.IsExpired(*payload, now) && !billingsdk.IsInGrace(*payload, offlineGraceWindow, now) {
		access.Reason = "license_expired"
		return access, nil
	}

	consumed, err := g.localConsumption(ctx, feature)
	if err != nil {
		return access, err
	}

	remaining, unlimited, ok := billingsdk.ComputeRemaining(*payload, feature, consumed)
	access.HasFeature = ok
	access.Remaining = remaining
	access.Unlimited = unlimited

	switch {
	case !ok:
		access.Reason = "feature_not_in_plan"
	case unlimited:
		access.Allowed = true
	case remaining > 0:
		access.Allowed = true
	default:
		access.Reason = "limit_reached"
	}
	return access, nil
}

func (g *Gate) Require(ctx context.Context, feature string) error {
	access, err := g.Check(ctx, feature)
	if err != nil {
		return err
	}
	if !access.Allowed {
		return &GateDenied{Feature: feature, Reason: access.Reason, Plan: access.Plan}
	}
	return nil
}

type GateDenied struct {
	Feature string
	Reason  string
	Plan    string
}

func (e *GateDenied) Error() string {
	return fmt.Sprintf("billing: feature %q denied (%s, plan=%s)", e.Feature, e.Reason, e.Plan)
}

var ErrFeatureNotAllowed = errors.New("billing: feature not allowed")

func (g *Gate) localConsumption(ctx context.Context, feature string) (uint64, error) {
	if g.usage == nil {
		return 0, nil
	}
	rows, err := g.usage.PendingBatch(ctx, 1000)
	if err != nil {
		return 0, err
	}
	var total uint64
	for _, row := range rows {
		if row.Feature == feature && row.Amount > 0 {
			total += uint64(row.Amount)
		}
	}
	return total, nil
}

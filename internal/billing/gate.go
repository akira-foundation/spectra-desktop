package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"spectra-desktop/internal/domain"
)

type FeatureAccess struct {
	Allowed  bool
	Reason   string
	Plan     string
	Status   string
	Grace    bool
}

type Gate struct {
	repo domain.LicenseRepository
}

func NewGate(repo domain.LicenseRepository) *Gate {
	return &Gate{repo: repo}
}

func (g *Gate) Check(ctx context.Context, feature string) (FeatureAccess, error) {
	if g.repo == nil {
		return FeatureAccess{Allowed: true}, nil
	}
	license, err := g.repo.Get(ctx)
	if err != nil {
		return FeatureAccess{}, err
	}
	if license == nil {
		return FeatureAccess{Allowed: false, Reason: "no_license"}, nil
	}

	features := map[string]bool{}
	if license.FeaturesJSON != "" {
		_ = json.Unmarshal([]byte(license.FeaturesJSON), &features)
	}
	allowed := features[feature]
	access := FeatureAccess{
		Allowed: allowed,
		Plan:    license.Plan,
		Status:  license.Status,
		Grace:   license.GracePeriod,
	}
	if !allowed {
		access.Reason = "feature_not_in_plan"
	}
	if license.Status != "active" {
		access.Allowed = false
		access.Reason = "license_" + license.Status
	}
	return access, nil
}

func (g *Gate) Require(ctx context.Context, feature string) error {
	access, err := g.Check(ctx, feature)
	if err != nil {
		return err
	}
	if !access.Allowed {
		return fmt.Errorf("billing: feature %q not allowed: %s", feature, access.Reason)
	}
	return nil
}

var ErrFeatureNotAllowed = errors.New("billing: feature not allowed")

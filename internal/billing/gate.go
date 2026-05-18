package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
)

const offlineGraceWindow = 7 * 24 * time.Hour

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
	inner  *billingsdk.Gate
}

func NewGate(client *Client, usage domain.UsageBufferRepository) *Gate {
	g := &Gate{client: client, usage: usage}
	g.inner = billingsdk.NewGate(billingsdk.GateOptions{
		Loader:           g.loadLicense,
		LocalConsumption: g.localConsumption,
		GraceWindow:      offlineGraceWindow,
	})
	return g
}

func (g *Gate) Check(ctx context.Context, feature string) (FeatureAccess, error) {
	if g.client == nil {
		return FeatureAccess{Allowed: true, Unlimited: true, HasFeature: true}, nil
	}
	access, err := g.inner.Check(ctx, feature)
	if err != nil {
		return FeatureAccess{}, err
	}
	return adaptAccess(access), nil
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

func (g *Gate) loadLicense(ctx context.Context) (*billingsdk.SignedLicense, *billingsdk.LicenseSnapshotPayload, error) {
	license, payload, err := g.client.VerifyLocal(ctx)
	if err != nil {
		return nil, nil, err
	}
	if license == nil || license.LicensePayload == "" || payload == nil {
		return nil, nil, nil
	}
	signed := &billingsdk.SignedLicense{
		KeyID:      license.LicenseKeyID,
		Algorithm:  license.LicenseAlgorithm,
		Payload:    license.LicensePayload,
		Signature:  license.LicenseSignature,
		ValidUntil: license.ValidUntil,
	}
	return signed, payload, nil
}

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

func adaptAccess(a billingsdk.FeatureAccess) FeatureAccess {
	status, grace := mapLicenseState(a.State)
	return FeatureAccess{
		Allowed:    a.Allowed,
		Reason:     a.Reason,
		Plan:       a.Plan,
		Status:     status,
		Grace:      grace,
		Remaining:  a.Remaining,
		Unlimited:  a.Unlimited,
		HasFeature: a.HasFeature,
	}
}

func mapLicenseState(state billingsdk.LicenseState) (string, bool) {
	switch state {
	case billingsdk.LicenseStateActive:
		return "active", false
	case billingsdk.LicenseStateGrace:
		return "active", true
	case billingsdk.LicenseStateExpired:
		return "expired", false
	case billingsdk.LicenseStateInvalid:
		return "invalid", false
	}
	return "", false
}

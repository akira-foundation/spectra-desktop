package app

import (
	"encoding/json"
	"fmt"
	"spectra-desktop/internal/billing"
)

func (a *App) BillingPlans() (map[string]any, error) {
	if a.billing == nil {
		return nil, fmt.Errorf("billing: not configured")
	}
	plans, err := a.billing.SDK().Plans(a.ctx)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(plans)
	if err != nil {
		return nil, err
	}
	out := map[string]any{}
	_ = json.Unmarshal(raw, &out)
	return out, nil
}

func (a *App) BillingOauthProviders() ([]string, error) {
	if a.billing == nil {
		return nil, fmt.Errorf("billing: not configured")
	}
	resp, err := a.billing.SDK().ListOauthProviders(a.ctx, billing.ProductSlug)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(resp.Providers))
	for _, p := range resp.Providers {
		out = append(out, string(p.Provider))
	}
	return out, nil
}

type FeatureAccessDTO struct {
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason,omitempty"`
	Plan       string `json:"plan,omitempty"`
	Status     string `json:"status,omitempty"`
	Grace      bool   `json:"grace"`
	Remaining  uint64 `json:"remaining"`
	Unlimited  bool   `json:"unlimited"`
	HasFeature bool   `json:"hasFeature"`
}

func (a *App) BillingCheckFeature(feature string) (*FeatureAccessDTO, error) {
	if a.billingGate == nil {
		return &FeatureAccessDTO{Allowed: true, Unlimited: true, HasFeature: true}, nil
	}
	access, err := a.billingGate.Check(a.ctx, feature)
	if err != nil {
		return nil, err
	}
	return &FeatureAccessDTO{
		Allowed:    access.Allowed,
		Reason:     access.Reason,
		Plan:       access.Plan,
		Status:     access.Status,
		Grace:      access.Grace,
		Remaining:  access.Remaining,
		Unlimited:  access.Unlimited,
		HasFeature: access.HasFeature,
	}, nil
}

type BillingTrackUsageInput struct {
	Feature string `json:"feature"`
	Amount  int    `json:"amount,omitempty"`
}

func (a *App) BillingTrackUsage(input BillingTrackUsageInput) error {
	if a.usage == nil {
		return nil
	}
	amount := input.Amount
	if amount <= 0 {
		amount = 1
	}
	return a.usage.Track(a.ctx, input.Feature, amount)
}

func (a *App) BillingFlushUsage() error {
	if a.usage == nil {
		return nil
	}
	return a.usage.Flush(a.ctx)
}

func (a *App) BillingPortal(returnURL string) (string, error) {
	if a.billing == nil {
		return "", fmt.Errorf("billing: not configured")
	}
	link, err := a.billing.SDK().BillingPortal(a.ctx, returnURL)
	if err != nil {
		return "", err
	}
	return link.URL, nil
}

package billing

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
)

func (h signedHarness) keyring() *Keyring {
	return &Keyring{keys: map[string]ed25519.PublicKey{"test": h.pub}}
}

func storedLicense(signed billingsdk.SignedLicense) *domain.License {
	return &domain.License{
		ID:               "local",
		Plan:             "pro",
		Status:           "active",
		LicenseKeyID:     signed.KeyID,
		LicenseAlgorithm: signed.Algorithm,
		LicensePayload:   signed.Payload,
		LicenseSignature: signed.Signature,
		ValidUntil:       signed.ValidUntil,
		FeaturesJSON:     "{}",
	}
}

func newGateClient(repo domain.LicenseRepository, kr *Keyring) *Client {
	return &Client{
		sdk:     billingsdk.NewClient("http://invalid.local", ProductSlug, "secret"),
		repo:    repo,
		keyring: kr,
	}
}

func TestGateCheck_AllowedWhenBoolFeatureEnabled(t *testing.T) {
	h := newSignedHarness(t)
	payload := billingsdk.LicenseSnapshotPayload{
		KeyID:      "test",
		PlanKey:    "pro",
		Features:   map[string]bool{"export": true},
		Usage:      map[string]billingsdk.UsageFeatureState{"export": {Type: "bool", Enabled: true}},
		IssuedAt:   time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		ValidUntil: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	signed := h.sign(t, payload)
	repo := &stubLicenseRepo{license: storedLicense(signed)}
	gate := NewGate(newGateClient(repo, h.keyring()), &stubUsageRepo{})

	access, err := gate.Check(context.Background(), "export")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !access.Allowed || !access.HasFeature || !access.Unlimited {
		t.Fatalf("expected allowed/unlimited bool feature, got %+v", access)
	}
}

func TestGateCheck_DeniedWhenExpiredOutsideGrace(t *testing.T) {
	h := newSignedHarness(t)
	past := time.Now().UTC().Add(-30 * 24 * time.Hour)
	payload := billingsdk.LicenseSnapshotPayload{
		KeyID:      "test",
		PlanKey:    "pro",
		Features:   map[string]bool{"export": true},
		Usage:      map[string]billingsdk.UsageFeatureState{"export": {Type: "bool", Enabled: true}},
		IssuedAt:   past.Add(-time.Hour).Format(time.RFC3339),
		ValidUntil: past.Format(time.RFC3339),
	}
	signed := h.sign(t, payload)
	repo := &stubLicenseRepo{license: storedLicense(signed)}
	gate := NewGate(newGateClient(repo, h.keyring()), &stubUsageRepo{})

	access, err := gate.Check(context.Background(), "export")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if access.Allowed {
		t.Fatalf("expected denied, got allowed")
	}
	if access.Reason != "license_expired" {
		t.Fatalf("expected reason license_expired, got %q", access.Reason)
	}
}

func TestGateCheck_DeniedWhenBoolFeatureDisabled(t *testing.T) {
	h := newSignedHarness(t)
	payload := billingsdk.LicenseSnapshotPayload{
		KeyID:      "test",
		Features:   map[string]bool{"export": false},
		Usage:      map[string]billingsdk.UsageFeatureState{"export": {Type: "bool", Enabled: false}},
		IssuedAt:   time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		ValidUntil: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	signed := h.sign(t, payload)
	repo := &stubLicenseRepo{license: storedLicense(signed)}
	gate := NewGate(newGateClient(repo, h.keyring()), &stubUsageRepo{})

	access, err := gate.Check(context.Background(), "export")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if access.Allowed {
		t.Fatalf("expected denied")
	}
	if access.Reason != "" {
		// bool-disabled lands in the default branch as limit_reached because remaining=0
		if access.Reason != "limit_reached" {
			t.Fatalf("unexpected reason %q", access.Reason)
		}
	}
	if !access.HasFeature {
		t.Fatalf("expected HasFeature true for known feature")
	}
}

func TestGateCheck_DeniedWhenCounterLimitReached(t *testing.T) {
	h := newSignedHarness(t)
	payload := billingsdk.LicenseSnapshotPayload{
		KeyID: "test",
		Usage: map[string]billingsdk.UsageFeatureState{
			"requests": {Type: "counter", Allowance: 100, ConsumedAtIssue: 90},
		},
		IssuedAt:   time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		ValidUntil: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	signed := h.sign(t, payload)
	repo := &stubLicenseRepo{license: storedLicense(signed)}
	usage := &stubUsageRepo{rows: []domain.UsageBufferEntry{
		{ID: "a", Feature: "requests", Amount: 7},
		{ID: "b", Feature: "requests", Amount: 5},
		{ID: "c", Feature: "other", Amount: 99},
	}}
	gate := NewGate(newGateClient(repo, h.keyring()), usage)

	access, err := gate.Check(context.Background(), "requests")
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if access.Allowed {
		t.Fatalf("expected denied, got allowed (remaining=%d)", access.Remaining)
	}
	if access.Reason != "limit_reached" {
		t.Fatalf("expected reason limit_reached, got %q", access.Reason)
	}
}

func TestGateRequire_WrapsDenialInTypedError(t *testing.T) {
	h := newSignedHarness(t)
	payload := billingsdk.LicenseSnapshotPayload{
		KeyID:      "test",
		PlanKey:    "pro",
		Usage:      map[string]billingsdk.UsageFeatureState{"export": {Type: "bool", Enabled: false}},
		IssuedAt:   time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		ValidUntil: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	signed := h.sign(t, payload)
	repo := &stubLicenseRepo{license: storedLicense(signed)}
	gate := NewGate(newGateClient(repo, h.keyring()), &stubUsageRepo{})

	err := gate.Require(context.Background(), "export")
	if err == nil {
		t.Fatalf("expected denial error")
	}
	var denied *GateDenied
	if !errors.As(err, &denied) {
		t.Fatalf("expected *GateDenied, got %T: %v", err, err)
	}
	if denied.Feature != "export" || denied.Plan != "pro" {
		t.Fatalf("denied carries wrong context: %+v", denied)
	}
}

func TestGateRequire_NilWhenAllowed(t *testing.T) {
	h := newSignedHarness(t)
	payload := billingsdk.LicenseSnapshotPayload{
		KeyID:      "test",
		Usage:      map[string]billingsdk.UsageFeatureState{"export": {Type: "bool", Enabled: true}},
		IssuedAt:   time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		ValidUntil: time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	signed := h.sign(t, payload)
	repo := &stubLicenseRepo{license: storedLicense(signed)}
	gate := NewGate(newGateClient(repo, h.keyring()), &stubUsageRepo{})

	if err := gate.Require(context.Background(), "export"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

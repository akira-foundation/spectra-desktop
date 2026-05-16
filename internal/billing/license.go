package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
)

const offlineGraceSeconds int64 = 7 * 24 * 60 * 60

type ActivationInput struct {
	DeviceName  string
	AppVersion  string
	Fingerprint string
}

func (c *Client) RequestOTP(ctx context.Context, email, deviceFP, appVersion string) error {
	return c.SDK().RequestOTP(ctx, billingsdk.OtpRequestPayload{
		Email:      email,
		DeviceFP:   deviceFP,
		Platform:   runtime.GOOS,
		AppVersion: appVersion,
	})
}

func (c *Client) VerifyOTP(ctx context.Context, email, code, deviceFP string) (*billingsdk.OtpVerifyResponse, error) {
	resp, err := c.SDK().VerifyOTP(ctx, billingsdk.OtpVerifyPayload{
		Email:    email,
		Code:     code,
		DeviceFP: deviceFP,
	})
	if err != nil {
		return nil, err
	}
	if err := c.PersistAccessToken(ctx, resp.AccessToken); err != nil {
		return nil, err
	}
	if err := c.persistCustomerFromOtp(ctx, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) ActivateLicense(ctx context.Context, input ActivationInput) (*domain.License, error) {
	if input.Fingerprint == "" {
		return nil, errors.New("billing: fingerprint required")
	}
	platform := runtime.GOOS
	deviceName := optional(input.DeviceName)
	appVersion := optional(input.AppVersion)

	resp, err := c.SDK().LicenseActivate(ctx, billingsdk.LicenseActivatePayload{
		Product:     ProductSlug,
		DeviceType:  "desktop",
		Platform:    &platform,
		DeviceName:  deviceName,
		AppVersion:  appVersion,
		Fingerprint: input.Fingerprint,
	})
	if err != nil {
		return nil, err
	}

	payload, err := c.keyring.Verify(resp.License)
	if err != nil {
		return nil, fmt.Errorf("billing: verify license: %w", err)
	}
	if payload.FingerprintHash != input.Fingerprint {
		return nil, fmt.Errorf("billing: license fingerprint mismatch")
	}

	return c.persistActivation(ctx, resp, payload)
}

func (c *Client) RefreshLicense(ctx context.Context, fingerprint string) (*domain.License, error) {
	resp, err := c.SDK().LicenseRefresh(ctx, billingsdk.LicenseRefreshPayload{
		Product:     ProductSlug,
		Fingerprint: fingerprint,
	})
	if err != nil {
		return nil, err
	}

	payload, err := c.keyring.Verify(resp.License)
	if err != nil {
		return nil, err
	}
	return c.persistActivation(ctx, resp, payload)
}

func (c *Client) VerifyLocal(ctx context.Context) (*domain.License, *billingsdk.LicenseSnapshotPayload, error) {
	license, err := c.repo.Get(ctx)
	if err != nil {
		return nil, nil, err
	}
	if license == nil || license.LicensePayload == "" {
		return license, nil, nil
	}
	signed := billingsdk.SignedLicense{
		KeyID:      license.LicenseKeyID,
		Algorithm:  license.LicenseAlgorithm,
		Payload:    license.LicensePayload,
		Signature:  license.LicenseSignature,
		ValidUntil: license.ValidUntil,
	}
	payload, err := c.keyring.Verify(signed)
	if err != nil {
		return license, nil, err
	}
	now := time.Now().UTC()
	expired := billingsdk.IsExpired(*payload, now)
	inGrace := billingsdk.IsInGrace(*payload, offlineGraceSeconds, now)
	switch {
	case !expired:
		license.Status = "active"
		license.GracePeriod = false
	case inGrace:
		license.Status = "active"
		license.GracePeriod = true
	default:
		license.Status = "expired"
		license.GracePeriod = false
	}
	license.LastVerifiedAt = now.Format(time.RFC3339)
	if err := c.repo.Save(ctx, *license); err != nil {
		return license, payload, err
	}
	return license, payload, nil
}

func (c *Client) persistActivation(ctx context.Context, resp *billingsdk.LicenseActivateResponse, payload *billingsdk.LicenseSnapshotPayload) (*domain.License, error) {
	license, err := c.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	if license == nil {
		license = &domain.License{ID: "local", FeaturesJSON: "{}"}
	}
	features, _ := json.Marshal(resp.Features)
	now := time.Now().UTC()
	license.Plan = resp.Plan
	license.Status = "active"
	license.ValidUntil = payload.ValidUntil
	license.ActivatedAt = payload.IssuedAt
	license.LastVerifiedAt = now.Format(time.RFC3339)
	license.LicenseKeyID = resp.License.KeyID
	license.LicenseAlgorithm = resp.License.Algorithm
	license.LicensePayload = resp.License.Payload
	license.LicenseSignature = resp.License.Signature
	license.FeaturesJSON = string(features)
	license.DeviceID = resp.Device.ID
	license.GracePeriod = false
	if err := c.repo.Save(ctx, *license); err != nil {
		return nil, err
	}
	return license, nil
}

func (c *Client) persistCustomerFromOtp(ctx context.Context, resp *billingsdk.OtpVerifyResponse) error {
	license, err := c.repo.Get(ctx)
	if err != nil {
		return err
	}
	if license == nil {
		license = &domain.License{ID: "local", Status: "inactive", FeaturesJSON: "{}"}
	}
	license.CustomerID = resp.Customer.ID
	license.CustomerEmail = resp.Customer.Email
	return c.repo.Save(ctx, *license)
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

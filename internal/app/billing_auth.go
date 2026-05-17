package app

import (
	"fmt"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"spectra-desktop/internal/billing"
	"spectra-desktop/internal/version"
)

func (a *App) BillingIsConfigured() bool {
	return a.billing != nil
}

func (a *App) BillingIsAuthenticated() bool {
	if a.billing == nil {
		return false
	}
	return a.billing.HasCustomerToken()
}

func (a *App) BillingGetLicense() (*LicenseDTO, error) {
	if a.licenseRepo == nil {
		return nil, nil
	}
	license, err := a.licenseRepo.Get(a.ctx)
	if err != nil || license == nil {
		return nil, err
	}
	return licenseToDTO(*license), nil
}

func (a *App) BillingVerifyLicense() (*LicenseDTO, error) {
	if a.billing == nil {
		return nil, nil
	}
	license, _, err := a.billing.VerifyLocal(a.ctx)
	if err != nil {
		wruntime.EventsEmit(a.ctx, "billing:license-error", err.Error())
		return nil, err
	}
	if license == nil {
		return nil, nil
	}
	dto := licenseToDTO(*license)
	wruntime.EventsEmit(a.ctx, "billing:license-changed", dto)
	return dto, nil
}

type BillingOtpRequestInput struct {
	Email string `json:"email"`
}

func (a *App) BillingRequestOTP(input BillingOtpRequestInput) error {
	if a.billing == nil {
		return fmt.Errorf("billing: not configured")
	}
	identity, err := a.machineIdentity()
	if err != nil {
		return err
	}
	return a.billing.RequestOTP(a.ctx, input.Email, identity.ID, version.Version)
}

type BillingOtpVerifyInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

func (a *App) BillingVerifyOTP(input BillingOtpVerifyInput) (*LicenseDTO, error) {
	if a.billing == nil {
		return nil, fmt.Errorf("billing: not configured")
	}
	identity, err := a.machineIdentity()
	if err != nil {
		return nil, err
	}
	if _, err := a.billing.VerifyOTP(a.ctx, input.Email, input.Code, identity.ID); err != nil {
		return nil, err
	}
	return a.BillingGetLicense()
}

func (a *App) BillingOauthLogin(provider string) (*OauthLoginResult, error) {
	if a.billing == nil {
		return nil, fmt.Errorf("billing: not configured")
	}
	openBrowser := func(url string) error {
		wruntime.BrowserOpenURL(a.ctx, url)
		return nil
	}
	result, err := a.billing.StartOauthLogin(a.ctx, provider, openBrowser)
	if err != nil {
		return nil, err
	}
	out := &OauthLoginResult{
		CustomerID:            result.Customer.ID,
		CustomerEmail:         result.Customer.Email,
		RequiresPlanSelection: result.RequiresPlanSelection,
	}
	if result.Customer.Name != nil {
		out.CustomerName = *result.Customer.Name
	}
	if result.Entitlement != nil {
		ent := &OauthEntitlement{Source: result.Entitlement.Source}
		if result.Entitlement.PlanKey != nil {
			ent.PlanKey = *result.Entitlement.PlanKey
		}
		if result.Entitlement.EndsAt != nil {
			ent.EndsAt = *result.Entitlement.EndsAt
		}
		out.Entitlement = ent
	}
	wruntime.EventsEmit(a.ctx, "billing:session-changed", out)
	return out, nil
}

func (a *App) BillingCancelOauth() {
	billing.CancelPendingOauth()
}

type BillingActivationInput struct {
	DeviceName string `json:"deviceName,omitempty"`
}

func (a *App) BillingActivateLicense(input BillingActivationInput) (*LicenseDTO, error) {
	if a.billing == nil {
		return nil, fmt.Errorf("billing: not configured")
	}
	identity, err := a.machineIdentity()
	if err != nil {
		return nil, err
	}
	deviceName := input.DeviceName
	if deviceName == "" {
		deviceName = defaultDeviceName()
	}
	license, err := a.billing.ActivateLicense(a.ctx, billing.ActivationInput{
		DeviceName:  deviceName,
		AppVersion:  version.Version,
		Fingerprint: identity.ID,
	})
	if err != nil {
		wruntime.EventsEmit(a.ctx, "billing:license-error", err.Error())
		return nil, err
	}
	dto := licenseToDTO(*license)
	wruntime.EventsEmit(a.ctx, "billing:license-changed", dto)
	return dto, nil
}

func (a *App) BillingRefreshLicense() (*LicenseDTO, error) {
	if a.billing == nil {
		return nil, fmt.Errorf("billing: not configured")
	}
	identity, err := a.machineIdentity()
	if err != nil {
		return nil, err
	}
	license, err := a.billing.RefreshLicense(a.ctx, identity.ID)
	if err != nil {
		return nil, err
	}
	dto := licenseToDTO(*license)
	wruntime.EventsEmit(a.ctx, "billing:license-changed", dto)
	return dto, nil
}

func (a *App) BillingLogout() error {
	if a.billing == nil {
		return nil
	}
	if err := a.billing.ClearSession(a.ctx); err != nil {
		return err
	}
	wruntime.EventsEmit(a.ctx, "billing:session-changed", nil)
	wruntime.EventsEmit(a.ctx, "billing:license-changed", nil)
	return nil
}

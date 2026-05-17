package app

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"spectra-desktop/internal/billing"
	"spectra-desktop/internal/domain"
)

type LicenseDTO struct {
	CustomerID        string          `json:"customerId"`
	CustomerEmail     string          `json:"customerEmail"`
	CustomerName      string          `json:"customerName"`
	Plan              string          `json:"plan"`
	Cycle             string          `json:"cycle"`
	Status            string          `json:"status"`
	ValidUntil        string          `json:"validUntil"`
	ActivatedAt       string          `json:"activatedAt"`
	LastVerifiedAt    string          `json:"lastVerifiedAt"`
	Features          map[string]bool `json:"features"`
	DeviceID          string          `json:"deviceId"`
	CancelAtPeriodEnd bool            `json:"cancelAtPeriodEnd"`
	CancelAt          string          `json:"cancelAt,omitempty"`
	TargetPlan        string          `json:"targetPlan,omitempty"`
	GracePeriod       bool            `json:"gracePeriod"`
}

type OauthLoginResult struct {
	CustomerID            string            `json:"customerId"`
	CustomerEmail         string            `json:"customerEmail"`
	CustomerName          string            `json:"customerName"`
	Entitlement           *OauthEntitlement `json:"entitlement,omitempty"`
	RequiresPlanSelection bool              `json:"requiresPlanSelection"`
}

type OauthEntitlement struct {
	PlanKey string `json:"planKey,omitempty"`
	Source  string `json:"source"`
	EndsAt  string `json:"endsAt,omitempty"`
}

func licenseToDTO(l domain.License) *LicenseDTO {
	features := map[string]bool{}
	if l.FeaturesJSON != "" {
		_ = json.Unmarshal([]byte(l.FeaturesJSON), &features)
	}
	return &LicenseDTO{
		CustomerID:        l.CustomerID,
		CustomerEmail:     l.CustomerEmail,
		CustomerName:      l.CustomerName,
		Plan:              l.Plan,
		Cycle:             l.Cycle,
		Status:            l.Status,
		ValidUntil:        l.ValidUntil,
		ActivatedAt:       l.ActivatedAt,
		LastVerifiedAt:    l.LastVerifiedAt,
		Features:          features,
		DeviceID:          l.DeviceID,
		CancelAtPeriodEnd: l.CancelAtPeriodEnd,
		CancelAt:          l.CancelAt,
		TargetPlan:        l.TargetPlan,
		GracePeriod:       l.GracePeriod,
	}
}

func defaultDeviceName() string {
	return fmt.Sprintf("Spectra %s · %s", runtime.GOOS, time.Now().UTC().Format("2006-01-02"))
}

func (a *App) machineIdentity() (*billing.MachineIdentity, error) {
	if a.machineID == nil {
		return nil, fmt.Errorf("billing: machine identity unavailable")
	}
	return a.machineID, nil
}

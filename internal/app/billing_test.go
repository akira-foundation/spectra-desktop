package app

import (
	"strings"
	"testing"

	"spectra-desktop/internal/domain"
)

func TestLicenseToDTO_ParsesFeaturesJSON(t *testing.T) {
	l := domain.License{
		CustomerID:   "cust_1",
		Plan:         "pro",
		Status:       "active",
		FeaturesJSON: `{"archives_export_import":true,"x":false}`,
	}
	dto := licenseToDTO(l)
	if dto.CustomerID != "cust_1" || dto.Plan != "pro" || dto.Status != "active" {
		t.Fatalf("scalars: %+v", dto)
	}
	if !dto.Features["archives_export_import"] || dto.Features["x"] {
		t.Fatalf("features: %v", dto.Features)
	}
}

func TestLicenseToDTO_EmptyFeaturesJSONYieldsEmptyMap(t *testing.T) {
	dto := licenseToDTO(domain.License{})
	if dto.Features == nil {
		t.Fatal("features map should be non-nil")
	}
	if len(dto.Features) != 0 {
		t.Fatalf("expected empty: %v", dto.Features)
	}
}

func TestLicenseToDTO_BadJSONFeaturesYieldsEmpty(t *testing.T) {
	dto := licenseToDTO(domain.License{FeaturesJSON: "not json"})
	if len(dto.Features) != 0 {
		t.Fatalf("expected empty on bad json, got %v", dto.Features)
	}
}

func TestDefaultDeviceName_IncludesGOOS(t *testing.T) {
	got := defaultDeviceName()
	if !strings.HasPrefix(got, "Spectra ") {
		t.Fatalf("got %q", got)
	}
	if !strings.Contains(got, " · ") {
		t.Fatalf("missing separator: %q", got)
	}
}

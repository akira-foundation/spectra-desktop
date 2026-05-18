package domain

import "testing"

func TestSettingsKey_Constants(t *testing.T) {
	if SettingActiveProjectID != "active_project_id" {
		t.Fatalf("SettingActiveProjectID drifted: %q", SettingActiveProjectID)
	}
	if SettingPHPBinaryPath != "php_binary_path" {
		t.Fatalf("SettingPHPBinaryPath drifted: %q", SettingPHPBinaryPath)
	}
}

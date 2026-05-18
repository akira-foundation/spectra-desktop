package domain

import (
	"testing"
	"time"
)

func TestProjectStatus_Constants(t *testing.T) {
	want := map[ProjectStatus]string{
		ProjectStatusConnected:    "connected",
		ProjectStatusDisconnected: "disconnected",
		ProjectStatusSyncing:      "syncing",
		ProjectStatusError:        "error",
	}
	for k, exp := range want {
		if string(k) != exp {
			t.Fatalf("status %q != %q", string(k), exp)
		}
	}
}

func TestAPIFilterMode_Constants(t *testing.T) {
	if APIFilterModeAuto != "auto" || APIFilterModeMiddleware != "middleware" ||
		APIFilterModePrefix != "prefix" || APIFilterModeAll != "all" {
		t.Fatalf("filter mode constants drifted")
	}
}

func TestProject_ZeroAndOptional(t *testing.T) {
	var p Project
	if p.LastSyncedAt != nil {
		t.Fatalf("LastSyncedAt should be nil")
	}
	now := time.Now().UTC()
	p.LastSyncedAt = &now
	if !p.LastSyncedAt.Equal(now) {
		t.Fatalf("LastSyncedAt round trip failed")
	}
}

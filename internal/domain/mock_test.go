package domain

import "testing"

func TestMockSource_Constants(t *testing.T) {
	want := map[MockSource]string{
		MockSourceAuto:      "auto",
		MockSourceHistory:   "history",
		MockSourceCustom:    "custom",
		MockSourceGenerated: "generated",
		MockSourceNoMatch:   "no-match",
	}
	for k, exp := range want {
		if string(k) != exp {
			t.Fatalf("source %q != %q", string(k), exp)
		}
	}
}

func TestMockOverride_ZeroValue(t *testing.T) {
	var m MockOverride
	if m.Enabled || m.Status != 0 || m.LatencyMs != 0 {
		t.Fatalf("zero override not empty: %+v", m)
	}
}

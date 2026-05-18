package core

import "testing"

func TestDetectionResult_ZeroValue(t *testing.T) {
	var r DetectionResult
	if r.Detected || r.Confidence != 0 || r.Version != "" || r.Markers != nil {
		t.Fatalf("zero detection not empty: %+v", r)
	}
}

func TestDriverDefaults_ZeroValue(t *testing.T) {
	var d DriverDefaults
	if d.BaseURL != "" || d.Ports != nil {
		t.Fatalf("zero defaults not empty: %+v", d)
	}
}

func TestDriverCapabilities_ZeroValue(t *testing.T) {
	var c DriverCapabilities
	if c.ScanRoutes || c.HasModels || c.Stats != nil {
		t.Fatalf("zero capabilities not empty: %+v", c)
	}
}

func TestFormattedException_Construction(t *testing.T) {
	e := FormattedException{
		Message: "boom",
		Class:   "X",
		File:    "f.go",
		Line:    1,
		Trace:   []FormattedTraceFrame{{File: "f.go", Line: 2, Function: "F"}},
		Extra:   map[string]any{"k": "v"},
	}
	if len(e.Trace) != 1 || e.Trace[0].Function != "F" {
		t.Fatalf("trace bad: %+v", e.Trace)
	}
	if e.Extra["k"] != "v" {
		t.Fatalf("extra bad: %+v", e.Extra)
	}
}

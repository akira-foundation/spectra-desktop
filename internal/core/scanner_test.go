package core

import (
	"context"
	"errors"
	"testing"
)

type fakeDriver struct {
	name   string
	result DetectionResult
	eps    []Endpoint
	err    error
}

func (f *fakeDriver) Name() string                     { return f.name }
func (f *fakeDriver) Detect(string) DetectionResult    { return f.result }
func (f *fakeDriver) Defaults() DriverDefaults         { return DriverDefaults{} }
func (f *fakeDriver) Capabilities() DriverCapabilities { return DriverCapabilities{} }
func (f *fakeDriver) Scan(_ context.Context, _ string) ([]Endpoint, error) {
	return f.eps, f.err
}

func TestNewScanner_Empty(t *testing.T) {
	s := NewScanner()
	if s == nil || len(s.Drivers()) != 0 {
		t.Fatalf("expected empty scanner")
	}
}

func TestScanner_Register(t *testing.T) {
	s := NewScanner()
	s.Register(&fakeDriver{name: "a"})
	s.Register(&fakeDriver{name: "b"})
	if len(s.Drivers()) != 2 {
		t.Fatalf("expected 2 drivers, got %d", len(s.Drivers()))
	}
}

func TestScanner_ResolveByName_Found(t *testing.T) {
	s := NewScanner()
	d := &fakeDriver{name: "laravel"}
	s.Register(d)
	got, err := s.ResolveByName("laravel")
	if err != nil || got != d {
		t.Fatalf("resolve by name: %v %v", got, err)
	}
}

func TestScanner_ResolveByName_NotFound(t *testing.T) {
	s := NewScanner()
	s.Register(&fakeDriver{name: "x"})
	if _, err := s.ResolveByName("missing"); !errors.Is(err, ErrNoDriver) {
		t.Fatalf("expected ErrNoDriver, got %v", err)
	}
}

func TestScanner_Resolve_PicksHighestConfidence(t *testing.T) {
	s := NewScanner()
	low := &fakeDriver{name: "low", result: DetectionResult{Detected: true, Confidence: 0.3}}
	high := &fakeDriver{name: "high", result: DetectionResult{Detected: true, Confidence: 0.9}}
	off := &fakeDriver{name: "off", result: DetectionResult{Detected: false, Confidence: 1.0}}
	s.Register(low)
	s.Register(off)
	s.Register(high)
	got, res, err := s.Resolve("/x")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if got != high || res.Confidence != 0.9 {
		t.Fatalf("expected high driver, got %v", got)
	}
}

func TestScanner_Resolve_NoMatch(t *testing.T) {
	s := NewScanner()
	s.Register(&fakeDriver{name: "x", result: DetectionResult{Detected: false}})
	_, _, err := s.Resolve("/x")
	if !errors.Is(err, ErrNoDriver) {
		t.Fatalf("expected ErrNoDriver, got %v", err)
	}
}

func TestScanner_Scan_Delegates(t *testing.T) {
	s := NewScanner()
	want := []Endpoint{{Method: MethodGet, Path: "/x"}}
	s.Register(&fakeDriver{name: "x", result: DetectionResult{Detected: true, Confidence: 1.0}, eps: want})
	got, err := s.Scan(context.Background(), "/p")
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(got) != 1 || got[0].Path != "/x" {
		t.Fatalf("scan result: %+v", got)
	}
}

func TestScanner_Scan_NoDriver(t *testing.T) {
	s := NewScanner()
	if _, err := s.Scan(context.Background(), "/p"); !errors.Is(err, ErrNoDriver) {
		t.Fatalf("expected ErrNoDriver, got %v", err)
	}
}

func TestScanner_Scan_PropagatesError(t *testing.T) {
	s := NewScanner()
	boom := errors.New("boom")
	s.Register(&fakeDriver{name: "x", result: DetectionResult{Detected: true, Confidence: 1.0}, err: boom})
	if _, err := s.Scan(context.Background(), "/p"); !errors.Is(err, boom) {
		t.Fatalf("expected boom, got %v", err)
	}
}

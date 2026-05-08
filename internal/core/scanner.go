package core

import (
	"context"
	"errors"
)

var ErrNoDriver = errors.New("no driver matched project")

type Scanner struct {
	drivers []FrameworkDriver
}

func NewScanner() *Scanner {
	return &Scanner{drivers: make([]FrameworkDriver, 0)}
}

func (s *Scanner) Register(d FrameworkDriver) {
	s.drivers = append(s.drivers, d)
}

func (s *Scanner) Drivers() []FrameworkDriver {
	return s.drivers
}

func (s *Scanner) ResolveByName(name string) (FrameworkDriver, error) {
	for _, d := range s.drivers {
		if d.Name() == name {
			return d, nil
		}
	}
	return nil, ErrNoDriver
}

func (s *Scanner) Resolve(projectPath string) (FrameworkDriver, DetectionResult, error) {
	var best FrameworkDriver
	var bestResult DetectionResult
	for _, d := range s.drivers {
		r := d.Detect(projectPath)
		if r.Detected && r.Confidence > bestResult.Confidence {
			best = d
			bestResult = r
		}
	}
	if best == nil {
		return nil, DetectionResult{}, ErrNoDriver
	}
	return best, bestResult, nil
}

func (s *Scanner) Scan(ctx context.Context, projectPath string) ([]Endpoint, error) {
	driver, _, err := s.Resolve(projectPath)
	if err != nil {
		return nil, err
	}
	return driver.Scan(ctx, projectPath)
}

package laravel

import (
	"os"
	"path/filepath"

	"spectra-desktop/internal/core"
)

var detectionMarkers = []string{
	"artisan",
	"composer.json",
	filepath.Join("routes", "web.php"),
	filepath.Join("routes", "api.php"),
	filepath.Join("app", "Http", "Kernel.php"),
}

func detect(projectPath string) core.DetectionResult {
	if projectPath == "" {
		return core.DetectionResult{}
	}
	found := make([]string, 0, len(detectionMarkers))
	for _, marker := range detectionMarkers {
		full := filepath.Join(projectPath, marker)
		if _, err := os.Stat(full); err == nil {
			found = append(found, marker)
		}
	}
	if len(found) == 0 {
		return core.DetectionResult{}
	}
	confidence := float64(len(found)) / float64(len(detectionMarkers))
	return core.DetectionResult{
		Detected:   confidence >= 0.4,
		Confidence: confidence,
		Markers:    found,
	}
}

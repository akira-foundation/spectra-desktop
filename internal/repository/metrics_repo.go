package repository

import (
	"github.com/uptrace/bun"
)

type MetricsRepository struct {
	db *bun.DB
}

func NewMetricsRepository(db *bun.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func percentile(sorted []int, p float64) int {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func shortURL(raw string) string {
	for i := 0; i < len(raw); i++ {
		if i+2 < len(raw) && raw[i] == '/' && raw[i+1] == '/' {
			for j := i + 2; j < len(raw); j++ {
				if raw[j] == '/' {
					return raw[j:]
				}
			}
			return "/"
		}
	}
	return raw
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

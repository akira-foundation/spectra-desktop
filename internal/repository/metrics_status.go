package repository

import (
	"context"
	"sort"
	"time"

	"spectra-desktop/internal/repository/model"
)

type StatusBucket struct {
	Bucket string
	Count  int
}

type MethodShare struct {
	Method  string
	Count   int
	Percent float64
}

func (r *MetricsRepository) StatusBuckets(ctx context.Context, projectID string, since time.Time) ([]StatusBucket, error) {
	var rows []model.RequestHistory
	q := r.db.NewSelect().
		Model(&rows).
		Column("response_status", "error").
		Where("project_id = ?", projectID)
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	counts := map[string]int{"2xx": 0, "3xx": 0, "4xx": 0, "5xx": 0, "err": 0}
	for _, row := range rows {
		if row.Error != "" {
			counts["err"]++
			continue
		}
		switch {
		case row.ResponseStatus >= 200 && row.ResponseStatus < 300:
			counts["2xx"]++
		case row.ResponseStatus >= 300 && row.ResponseStatus < 400:
			counts["3xx"]++
		case row.ResponseStatus >= 400 && row.ResponseStatus < 500:
			counts["4xx"]++
		case row.ResponseStatus >= 500:
			counts["5xx"]++
		default:
			counts["err"]++
		}
	}
	return []StatusBucket{
		{"2xx", counts["2xx"]},
		{"3xx", counts["3xx"]},
		{"4xx", counts["4xx"]},
		{"5xx", counts["5xx"]},
		{"err", counts["err"]},
	}, nil
}

func (r *MetricsRepository) MethodShare(ctx context.Context, projectID string, since time.Time) ([]MethodShare, error) {
	var rows []model.RequestHistory
	q := r.db.NewSelect().
		Model(&rows).
		Column("method").
		Where("project_id = ?", projectID)
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	counts := map[string]int{}
	total := 0
	for _, row := range rows {
		m := row.Method
		if m == "" {
			continue
		}
		counts[m]++
		total++
	}
	out := make([]MethodShare, 0, len(counts))
	for m, c := range counts {
		pct := 0.0
		if total > 0 {
			pct = float64(c) / float64(total)
		}
		out = append(out, MethodShare{Method: m, Count: c, Percent: pct})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out, nil
}

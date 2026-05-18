package repository

import (
	"context"
	"sort"
	"time"

	"spectra-desktop/internal/repository/model"
)

type LatencyStats struct {
	Count int
	Avg   int
	Min   int
	Max   int
	P50   int
	P95   int
	P99   int
}

type LatencyPoint struct {
	Day   time.Time
	AvgMs int
	Count int
}

type EndpointLatencySeries struct {
	EndpointID string
	Method     string
	Path       string
	Points     []LatencyPoint
	AvgMs      int
}

func (r *MetricsRepository) LatencyStats(ctx context.Context, projectID string, since time.Time) (LatencyStats, error) {
	var rows []model.RequestHistory
	q := r.db.NewSelect().
		Model(&rows).
		Column("duration_ms").
		Where("project_id = ?", projectID).
		Where("error = ''").
		Where("duration_ms > 0")
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	err := q.Scan(ctx)
	if err != nil {
		return LatencyStats{}, err
	}
	if len(rows) == 0 {
		return LatencyStats{}, nil
	}
	values := make([]int, len(rows))
	for i, row := range rows {
		values[i] = row.DurationMs
	}
	sort.Ints(values)
	sum := 0
	for _, v := range values {
		sum += v
	}
	return LatencyStats{
		Count: len(values),
		Avg:   sum / len(values),
		Min:   values[0],
		Max:   values[len(values)-1],
		P50:   percentile(values, 0.50),
		P95:   percentile(values, 0.95),
		P99:   percentile(values, 0.99),
	}, nil
}

func (r *MetricsRepository) LatencyOverTime(ctx context.Context, projectID string, days int, topN int) ([]EndpointLatencySeries, error) {
	if days <= 0 {
		days = 7
	}
	if topN <= 0 {
		topN = 5
	}
	since := time.Now().UTC().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	var rows []model.RequestHistory
	err := r.db.NewSelect().
		Model(&rows).
		Column("endpoint_id", "method", "url", "duration_ms", "created_at", "error").
		Where("project_id = ?", projectID).
		Where("created_at >= ?", since).
		Where("endpoint_id != ''").
		Where("error = ''").
		Where("duration_ms > 0").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	type bucket struct {
		method string
		path   string
		days   map[string]struct {
			total int
			count int
		}
		grandTotal int
		grandCount int
	}
	byEP := map[string]*bucket{}
	for _, row := range rows {
		b, ok := byEP[row.EndpointID]
		if !ok {
			b = &bucket{method: row.Method, path: shortURL(row.URL), days: map[string]struct {
				total int
				count int
			}{}}
			byEP[row.EndpointID] = b
		}
		key := row.CreatedAt.UTC().Format("2006-01-02")
		entry := b.days[key]
		entry.total += row.DurationMs
		entry.count++
		b.days[key] = entry
		b.grandTotal += row.DurationMs
		b.grandCount++
	}
	out := make([]EndpointLatencySeries, 0, len(byEP))
	for id, b := range byEP {
		points := make([]LatencyPoint, 0, days)
		for i := 0; i < days; i++ {
			day := since.AddDate(0, 0, i)
			key := day.Format("2006-01-02")
			d := b.days[key]
			avg := 0
			if d.count > 0 {
				avg = d.total / d.count
			}
			points = append(points, LatencyPoint{Day: day, AvgMs: avg, Count: d.count})
		}
		avg := 0
		if b.grandCount > 0 {
			avg = b.grandTotal / b.grandCount
		}
		out = append(out, EndpointLatencySeries{
			EndpointID: id,
			Method:     b.method,
			Path:       b.path,
			Points:     points,
			AvgMs:      avg,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AvgMs > out[j].AvgMs })
	if len(out) > topN {
		out = out[:topN]
	}
	return out, nil
}

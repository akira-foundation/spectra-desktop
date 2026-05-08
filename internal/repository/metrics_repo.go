package repository

import (
	"context"
	"sort"
	"time"

	"github.com/uptrace/bun"

	"spectra-desktop/internal/repository/model"
)

type MetricsRepository struct {
	db *bun.DB
}

func NewMetricsRepository(db *bun.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

type StatusBucket struct {
	Bucket string
	Count  int
}

type LatencyStats struct {
	Count int
	Avg   int
	Min   int
	Max   int
	P50   int
	P95   int
	P99   int
}

type DailyVolume struct {
	Day   time.Time
	Count int
}

type EndpointMetric struct {
	EndpointID string
	Method     string
	Path       string
	Count      int
	Errors     int
	AvgMs      int
}

func (r *MetricsRepository) StatusBuckets(ctx context.Context, projectID string) ([]StatusBucket, error) {
	var rows []model.RequestHistory
	err := r.db.NewSelect().
		Model(&rows).
		Column("response_status", "error").
		Where("project_id = ?", projectID).
		Scan(ctx)
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
	buckets := []StatusBucket{
		{"2xx", counts["2xx"]},
		{"3xx", counts["3xx"]},
		{"4xx", counts["4xx"]},
		{"5xx", counts["5xx"]},
		{"err", counts["err"]},
	}
	return buckets, nil
}

func (r *MetricsRepository) LatencyStats(ctx context.Context, projectID string) (LatencyStats, error) {
	var rows []model.RequestHistory
	err := r.db.NewSelect().
		Model(&rows).
		Column("duration_ms").
		Where("project_id = ?", projectID).
		Where("error = ''").
		Where("duration_ms > 0").
		Scan(ctx)
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

func percentile(sorted []int, p float64) int {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func (r *MetricsRepository) DailyVolume(ctx context.Context, projectID string, days int) ([]DailyVolume, error) {
	if days <= 0 {
		days = 7
	}
	since := time.Now().UTC().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	var rows []model.RequestHistory
	err := r.db.NewSelect().
		Model(&rows).
		Column("created_at").
		Where("project_id = ?", projectID).
		Where("created_at >= ?", since).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	bucket := map[string]int{}
	for _, row := range rows {
		key := row.CreatedAt.UTC().Format("2006-01-02")
		bucket[key]++
	}
	out := make([]DailyVolume, 0, days)
	for i := 0; i < days; i++ {
		day := since.AddDate(0, 0, i)
		key := day.Format("2006-01-02")
		out = append(out, DailyVolume{Day: day, Count: bucket[key]})
	}
	return out, nil
}

func (r *MetricsRepository) EndpointMetrics(ctx context.Context, projectID string) ([]EndpointMetric, error) {
	var rows []model.RequestHistory
	err := r.db.NewSelect().
		Model(&rows).
		Column("endpoint_id", "method", "url", "duration_ms", "response_status", "error").
		Where("project_id = ?", projectID).
		Where("endpoint_id != ''").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	type agg struct {
		method string
		path   string
		count  int
		errors int
		total  int
	}
	bucket := map[string]*agg{}
	for _, row := range rows {
		key := row.EndpointID
		a, ok := bucket[key]
		if !ok {
			a = &agg{method: row.Method, path: shortURL(row.URL)}
			bucket[key] = a
		}
		a.count++
		a.total += row.DurationMs
		if row.Error != "" || row.ResponseStatus >= 400 {
			a.errors++
		}
	}
	out := make([]EndpointMetric, 0, len(bucket))
	for id, a := range bucket {
		avg := 0
		if a.count > 0 {
			avg = a.total / a.count
		}
		out = append(out, EndpointMetric{
			EndpointID: id,
			Method:     a.method,
			Path:       a.path,
			Count:      a.count,
			Errors:     a.errors,
			AvgMs:      avg,
		})
	}
	return out, nil
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

package repository

import (
	"context"
	"sort"
	"time"

	"spectra-desktop/internal/repository/model"
)

type EndpointMetric struct {
	EndpointID string
	Method     string
	Path       string
	Count      int
	Errors     int
	AvgMs      int
}

type FlakyEndpoint struct {
	EndpointID string
	Method     string
	Path       string
	Total      int
	Successes  int
	Failures   int
	FlakeScore float64
}

type FailureSeries struct {
	EndpointID string
	Method     string
	Path       string
	Failures   int
	Points     []UsagePoint
}

func (r *MetricsRepository) EndpointMetrics(ctx context.Context, projectID string, since time.Time) ([]EndpointMetric, error) {
	var rows []model.RequestHistory
	q := r.db.NewSelect().
		Model(&rows).
		Column("endpoint_id", "method", "url", "duration_ms", "response_status", "error").
		Where("project_id = ?", projectID).
		Where("endpoint_id != ''")
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	err := q.Scan(ctx)
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

func (r *MetricsRepository) FlakyEndpoints(ctx context.Context, projectID string, minRuns int) ([]FlakyEndpoint, error) {
	if minRuns <= 0 {
		minRuns = 3
	}
	var rows []model.RequestHistory
	err := r.db.NewSelect().
		Model(&rows).
		Column("endpoint_id", "method", "url", "response_status", "error").
		Where("project_id = ?", projectID).
		Where("endpoint_id != ''").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	type agg struct {
		method    string
		path      string
		total     int
		successes int
		failures  int
	}
	byEP := map[string]*agg{}
	for _, row := range rows {
		a, ok := byEP[row.EndpointID]
		if !ok {
			a = &agg{method: row.Method, path: shortURL(row.URL)}
			byEP[row.EndpointID] = a
		}
		a.total++
		fail := row.Error != "" || row.ResponseStatus >= 400
		if fail {
			a.failures++
		} else {
			a.successes++
		}
	}
	out := make([]FlakyEndpoint, 0)
	for id, a := range byEP {
		if a.total < minRuns || a.successes == 0 || a.failures == 0 {
			continue
		}
		failRate := float64(a.failures) / float64(a.total)
		flake := 1.0 - 2.0*absFloat(failRate-0.5)
		out = append(out, FlakyEndpoint{
			EndpointID: id,
			Method:     a.method,
			Path:       a.path,
			Total:      a.total,
			Successes:  a.successes,
			Failures:   a.failures,
			FlakeScore: flake,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].FlakeScore == out[j].FlakeScore {
			return out[i].Total > out[j].Total
		}
		return out[i].FlakeScore > out[j].FlakeScore
	})
	return out, nil
}

func (r *MetricsRepository) FailuresOverTime(ctx context.Context, projectID string, days int, topN int) ([]FailureSeries, error) {
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
		Column("endpoint_id", "method", "url", "created_at", "response_status", "error").
		Where("project_id = ?", projectID).
		Where("created_at >= ?", since).
		Where("endpoint_id != ''").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	type bucket struct {
		method string
		path   string
		days   map[string]int
		total  int
	}
	byEP := map[string]*bucket{}
	for _, row := range rows {
		fail := row.Error != "" || row.ResponseStatus >= 400
		if !fail {
			continue
		}
		b, ok := byEP[row.EndpointID]
		if !ok {
			b = &bucket{method: row.Method, path: shortURL(row.URL), days: map[string]int{}}
			byEP[row.EndpointID] = b
		}
		key := row.CreatedAt.UTC().Format("2006-01-02")
		b.days[key]++
		b.total++
	}
	out := make([]FailureSeries, 0, len(byEP))
	for id, b := range byEP {
		points := make([]UsagePoint, 0, days)
		for i := 0; i < days; i++ {
			day := since.AddDate(0, 0, i)
			key := day.Format("2006-01-02")
			points = append(points, UsagePoint{Day: day, Count: b.days[key]})
		}
		out = append(out, FailureSeries{
			EndpointID: id, Method: b.method, Path: b.path, Failures: b.total, Points: points,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Failures > out[j].Failures })
	if len(out) > topN {
		out = out[:topN]
	}
	return out, nil
}

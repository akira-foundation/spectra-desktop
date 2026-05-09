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
	buckets := []StatusBucket{
		{"2xx", counts["2xx"]},
		{"3xx", counts["3xx"]},
		{"4xx", counts["4xx"]},
		{"5xx", counts["5xx"]},
		{"err", counts["err"]},
	}
	return buckets, nil
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

type HourlyCell struct {
	Day   int
	Hour  int
	Count int
}

func (r *MetricsRepository) HourlyHeatmap(ctx context.Context, projectID string, days int) ([]HourlyCell, error) {
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
	cells := map[[2]int]int{}
	for _, row := range rows {
		t := row.CreatedAt.UTC()
		dow := int(t.Weekday())
		hour := t.Hour()
		cells[[2]int{dow, hour}]++
	}
	out := make([]HourlyCell, 0, 7*24)
	for d := 0; d < 7; d++ {
		for h := 0; h < 24; h++ {
			out = append(out, HourlyCell{Day: d, Hour: h, Count: cells[[2]int{d, h}]})
		}
	}
	return out, nil
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
		// flakiness peaks at 50/50
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

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

type UsagePoint struct {
	Day   time.Time
	Count int
}

type EndpointUsageSeries struct {
	EndpointID string
	Method     string
	Path       string
	Total      int
	Points     []UsagePoint
}

func (r *MetricsRepository) UsageOverTime(ctx context.Context, projectID string, days int, topN int) ([]EndpointUsageSeries, error) {
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
		Column("endpoint_id", "method", "url", "created_at").
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
		b, ok := byEP[row.EndpointID]
		if !ok {
			b = &bucket{method: row.Method, path: shortURL(row.URL), days: map[string]int{}}
			byEP[row.EndpointID] = b
		}
		key := row.CreatedAt.UTC().Format("2006-01-02")
		b.days[key]++
		b.total++
	}
	out := make([]EndpointUsageSeries, 0, len(byEP))
	for id, b := range byEP {
		points := make([]UsagePoint, 0, days)
		for i := 0; i < days; i++ {
			day := since.AddDate(0, 0, i)
			key := day.Format("2006-01-02")
			points = append(points, UsagePoint{Day: day, Count: b.days[key]})
		}
		out = append(out, EndpointUsageSeries{
			EndpointID: id, Method: b.method, Path: b.path, Total: b.total, Points: points,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Total > out[j].Total })
	if len(out) > topN {
		out = out[:topN]
	}
	return out, nil
}

type MethodShare struct {
	Method  string
	Count   int
	Percent float64
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

type FailureSeries struct {
	EndpointID string
	Method     string
	Path       string
	Failures   int
	Points     []UsagePoint
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

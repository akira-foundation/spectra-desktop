package repository

import (
	"context"
	"sort"
	"time"

	"spectra-desktop/internal/repository/model"
)

type DailyVolume struct {
	Day   time.Time
	Count int
}

type HourlyCell struct {
	Day   int
	Hour  int
	Count int
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

package app

import "time"

type DashboardMetrics struct {
	StatusBuckets []StatusBucketDTO   `json:"statusBuckets"`
	Latency       LatencyDTO          `json:"latency"`
	Volume        []VolumePoint       `json:"volume"`
	TotalRuns     int                 `json:"totalRuns"`
	ErrorRate     float64             `json:"errorRate"`
	TopSlow       []EndpointMetricDTO `json:"topSlow"`
	TopFailing    []EndpointMetricDTO `json:"topFailing"`
	TopUsed       []EndpointMetricDTO `json:"topUsed"`
}

type StatusBucketDTO struct {
	Bucket string `json:"bucket"`
	Count  int    `json:"count"`
}

type LatencyDTO struct {
	Count int `json:"count"`
	Avg   int `json:"avg"`
	Min   int `json:"min"`
	Max   int `json:"max"`
	P50   int `json:"p50"`
	P95   int `json:"p95"`
	P99   int `json:"p99"`
}

type VolumePoint struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type EndpointMetricDTO struct {
	EndpointID string  `json:"endpointID"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Count      int     `json:"count"`
	Errors     int     `json:"errors"`
	AvgMs      int     `json:"avgMs"`
	ErrorRate  float64 `json:"errorRate"`
}

func (a *App) GetDashboardMetrics(projectID string, volumeDays int) (*DashboardMetrics, error) {
	if projectID == "" {
		return nil, nil
	}
	if volumeDays <= 0 {
		volumeDays = 7
	}
	since := time.Now().UTC().AddDate(0, 0, -volumeDays+1).Truncate(24 * time.Hour)
	out := &DashboardMetrics{
		StatusBuckets: []StatusBucketDTO{},
		Volume:        []VolumePoint{},
		TopSlow:       []EndpointMetricDTO{},
		TopFailing:    []EndpointMetricDTO{},
		TopUsed:       []EndpointMetricDTO{},
	}

	buckets, err := a.metrics.StatusBuckets(a.ctx, projectID, since)
	if err != nil {
		return nil, err
	}
	total := 0
	errs := 0
	for _, b := range buckets {
		out.StatusBuckets = append(out.StatusBuckets, StatusBucketDTO{Bucket: b.Bucket, Count: b.Count})
		total += b.Count
		if b.Bucket == "4xx" || b.Bucket == "5xx" || b.Bucket == "err" {
			errs += b.Count
		}
	}
	out.TotalRuns = total
	if total > 0 {
		out.ErrorRate = float64(errs) / float64(total)
	}

	lat, err := a.metrics.LatencyStats(a.ctx, projectID, since)
	if err != nil {
		return nil, err
	}
	out.Latency = LatencyDTO{
		Count: lat.Count,
		Avg:   lat.Avg,
		Min:   lat.Min,
		Max:   lat.Max,
		P50:   lat.P50,
		P95:   lat.P95,
		P99:   lat.P99,
	}

	volume, err := a.metrics.DailyVolume(a.ctx, projectID, volumeDays)
	if err != nil {
		return nil, err
	}
	for _, v := range volume {
		out.Volume = append(out.Volume, VolumePoint{Day: v.Day.Format("2006-01-02"), Count: v.Count})
	}

	endpointStats, err := a.metrics.EndpointMetrics(a.ctx, projectID, since)
	if err != nil {
		return nil, err
	}
	all := make([]EndpointMetricDTO, 0, len(endpointStats))
	for _, e := range endpointStats {
		rate := 0.0
		if e.Count > 0 {
			rate = float64(e.Errors) / float64(e.Count)
		}
		all = append(all, EndpointMetricDTO{
			EndpointID: e.EndpointID,
			Method:     e.Method,
			Path:       e.Path,
			Count:      e.Count,
			Errors:     e.Errors,
			AvgMs:      e.AvgMs,
			ErrorRate:  rate,
		})
	}
	out.TopSlow = topNBy(all, func(a, b EndpointMetricDTO) bool { return a.AvgMs > b.AvgMs }, 5)
	out.TopUsed = topNBy(all, func(a, b EndpointMetricDTO) bool { return a.Count > b.Count }, 5)
	failing := []EndpointMetricDTO{}
	for _, e := range all {
		if e.Errors > 0 {
			failing = append(failing, e)
		}
	}
	out.TopFailing = topNBy(failing, func(a, b EndpointMetricDTO) bool { return a.ErrorRate > b.ErrorRate }, 5)

	return out, nil
}

package repository

import (
	"context"
	"time"
)

type CoverageStats struct {
	TotalEndpoints  int
	UsedEndpoints   int
	UnusedEndpoints []string
	StaleEndpoints  []StaleEndpoint
}

type StaleEndpoint struct {
	EndpointID string
	Method     string
	Path       string
	LastSeen   time.Time
	DaysAgo    int
}

func (r *MetricsRepository) Coverage(ctx context.Context, projectID string, staleAfterDays int) (CoverageStats, error) {
	if staleAfterDays <= 0 {
		staleAfterDays = 30
	}
	type epRow struct {
		EndpointID string     `bun:"endpoint_id"`
		LatestAt   *time.Time `bun:"latest_at"`
	}
	var seen []epRow
	if err := r.db.NewRaw(
		`SELECT endpoint_id, MAX(created_at) AS latest_at
		 FROM request_history
		 WHERE project_id = ? AND endpoint_id != ''
		 GROUP BY endpoint_id`,
		projectID,
	).Scan(ctx, &seen); err != nil {
		return CoverageStats{}, err
	}
	seenMap := map[string]time.Time{}
	for _, r := range seen {
		if r.LatestAt != nil {
			seenMap[r.EndpointID] = *r.LatestAt
		}
	}
	return CoverageStats{
		TotalEndpoints: 0,
		UsedEndpoints:  len(seenMap),
	}, nil
}

func (r *MetricsRepository) LastSeenByEndpoint(ctx context.Context, projectID string) (map[string]time.Time, error) {
	type row struct {
		EndpointID string     `bun:"endpoint_id"`
		LatestAt   *time.Time `bun:"latest_at"`
	}
	var rows []row
	if err := r.db.NewRaw(
		`SELECT endpoint_id, MAX(created_at) AS latest_at
		 FROM request_history
		 WHERE project_id = ? AND endpoint_id != ''
		 GROUP BY endpoint_id`,
		projectID,
	).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	out := make(map[string]time.Time, len(rows))
	for _, r := range rows {
		if r.LatestAt != nil {
			out[r.EndpointID] = *r.LatestAt
		}
	}
	return out, nil
}

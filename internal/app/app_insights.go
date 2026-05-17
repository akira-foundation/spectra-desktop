package app

import (
	"strings"
	"time"
)

type LatencyPointDTO struct {
	Day   string `json:"day"`
	AvgMs int    `json:"avgMs"`
	Count int    `json:"count"`
}

type EndpointLatencySeriesDTO struct {
	EndpointID string            `json:"endpointID"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	AvgMs      int               `json:"avgMs"`
	Points     []LatencyPointDTO `json:"points"`
}

type HourlyCellDTO struct {
	Day   int `json:"day"`
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type FlakyEndpointDTO struct {
	EndpointID string  `json:"endpointID"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Total      int     `json:"total"`
	Successes  int     `json:"successes"`
	Failures   int     `json:"failures"`
	FlakeScore float64 `json:"flakeScore"`
}

type EndpointUsageSeriesDTO struct {
	EndpointID string            `json:"endpointID"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Total      int               `json:"total"`
	Points     []LatencyPointDTO `json:"points"`
}

type EndpointFailureSeriesDTO struct {
	EndpointID string            `json:"endpointID"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Failures   int               `json:"failures"`
	Points     []LatencyPointDTO `json:"points"`
}

type MethodShareDTO struct {
	Method  string  `json:"method"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

type EndpointDiscoveryDTO struct {
	EndpointID string `json:"endpointID"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	LastSeen   int64  `json:"lastSeen,omitempty"`
	DaysAgo    int    `json:"daysAgo,omitempty"`
}

type DiscoveryDTO struct {
	TotalEndpoints  int                    `json:"totalEndpoints"`
	UsedEndpoints   int                    `json:"usedEndpoints"`
	Coverage        float64                `json:"coverage"`
	Unused          []EndpointDiscoveryDTO `json:"unused"`
	Stale           []EndpointDiscoveryDTO `json:"stale"`
	TestedEndpoints int                    `json:"testedEndpoints"`
	TestCoverage    float64                `json:"testCoverage"`
	WriteEndpoints  int                    `json:"writeEndpoints"`
	ReadEndpoints   int                    `json:"readEndpoints"`
	AuthRequired    int                    `json:"authRequired"`
	AuthPublic      int                    `json:"authPublic"`
}

func (a *App) GetDiscovery(projectID string, staleAfterDays int) (*DiscoveryDTO, error) {
	if projectID == "" {
		return &DiscoveryDTO{Unused: []EndpointDiscoveryDTO{}, Stale: []EndpointDiscoveryDTO{}}, nil
	}
	if staleAfterDays <= 0 {
		staleAfterDays = 30
	}
	endpoints, err := a.endpoints.List(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	seen, err := a.metrics.LastSeenByEndpoint(a.ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := &DiscoveryDTO{
		TotalEndpoints: len(endpoints),
		UsedEndpoints:  len(seen),
		Unused:         []EndpointDiscoveryDTO{},
		Stale:          []EndpointDiscoveryDTO{},
	}
	if len(endpoints) > 0 {
		out.Coverage = float64(len(seen)) / float64(len(endpoints))
	}
	// build set of endpoint keys with tests or captures
	testedKeys := map[string]bool{}
	for _, e := range endpoints {
		key := endpointTestKey(string(e.Method), e.Path)
		if tests, _ := a.tests.List(a.ctx, projectID, key); len(tests) > 0 {
			testedKeys[key] = true
			continue
		}
		if caps, _ := a.captures.List(a.ctx, projectID, key); len(caps) > 0 {
			testedKeys[key] = true
		}
	}
	out.TestedEndpoints = len(testedKeys)
	if len(endpoints) > 0 {
		out.TestCoverage = float64(len(testedKeys)) / float64(len(endpoints))
	}

	writeMethods := map[string]bool{"POST": true, "PUT": true, "PATCH": true, "DELETE": true}
	for _, e := range endpoints {
		if writeMethods[strings.ToUpper(string(e.Method))] {
			out.WriteEndpoints++
		} else {
			out.ReadEndpoints++
		}
		if e.AuthRole != "" {
			out.AuthRequired++
		} else {
			out.AuthPublic++
		}
	}

	now := time.Now().UTC()
	staleThreshold := now.AddDate(0, 0, -staleAfterDays)
	for _, e := range endpoints {
		ts, ok := seen[e.ID]
		if !ok {
			out.Unused = append(out.Unused, EndpointDiscoveryDTO{
				EndpointID: e.ID, Method: string(e.Method), Path: e.Path,
			})
			continue
		}
		if ts.Before(staleThreshold) {
			out.Stale = append(out.Stale, EndpointDiscoveryDTO{
				EndpointID: e.ID,
				Method:     string(e.Method),
				Path:       e.Path,
				LastSeen:   ts.Unix(),
				DaysAgo:    int(now.Sub(ts).Hours() / 24),
			})
		}
	}
	return out, nil
}

type InsightsDTO struct {
	LatencyOverTime  []EndpointLatencySeriesDTO `json:"latencyOverTime"`
	UsageOverTime    []EndpointUsageSeriesDTO   `json:"usageOverTime"`
	FailuresOverTime []EndpointFailureSeriesDTO `json:"failuresOverTime"`
	HourlyHeatmap    []HourlyCellDTO            `json:"hourlyHeatmap"`
	Flaky            []FlakyEndpointDTO         `json:"flaky"`
	MethodShare      []MethodShareDTO           `json:"methodShare"`
}

func (a *App) GetInsights(projectID string, days int) (*InsightsDTO, error) {
	if projectID == "" {
		return &InsightsDTO{LatencyOverTime: []EndpointLatencySeriesDTO{}, HourlyHeatmap: []HourlyCellDTO{}, Flaky: []FlakyEndpointDTO{}}, nil
	}
	if days <= 0 {
		days = 7
	}
	out := &InsightsDTO{
		LatencyOverTime:  []EndpointLatencySeriesDTO{},
		UsageOverTime:    []EndpointUsageSeriesDTO{},
		FailuresOverTime: []EndpointFailureSeriesDTO{},
		HourlyHeatmap:    []HourlyCellDTO{},
		Flaky:            []FlakyEndpointDTO{},
		MethodShare:      []MethodShareDTO{},
	}
	since := time.Now().UTC().AddDate(0, 0, -days+1).Truncate(24 * time.Hour)
	if shares, err := a.metrics.MethodShare(a.ctx, projectID, since); err == nil {
		for _, s := range shares {
			out.MethodShare = append(out.MethodShare, MethodShareDTO{Method: s.Method, Count: s.Count, Percent: s.Percent})
		}
	}
	if series, err := a.metrics.LatencyOverTime(a.ctx, projectID, days, 5); err == nil {
		for _, s := range series {
			pts := make([]LatencyPointDTO, 0, len(s.Points))
			for _, p := range s.Points {
				pts = append(pts, LatencyPointDTO{Day: p.Day.Format("2006-01-02"), AvgMs: p.AvgMs, Count: p.Count})
			}
			out.LatencyOverTime = append(out.LatencyOverTime, EndpointLatencySeriesDTO{
				EndpointID: s.EndpointID, Method: s.Method, Path: s.Path, AvgMs: s.AvgMs, Points: pts,
			})
		}
	}
	if series, err := a.metrics.UsageOverTime(a.ctx, projectID, days, 5); err == nil {
		for _, s := range series {
			pts := make([]LatencyPointDTO, 0, len(s.Points))
			for _, p := range s.Points {
				pts = append(pts, LatencyPointDTO{Day: p.Day.Format("2006-01-02"), Count: p.Count})
			}
			out.UsageOverTime = append(out.UsageOverTime, EndpointUsageSeriesDTO{
				EndpointID: s.EndpointID, Method: s.Method, Path: s.Path, Total: s.Total, Points: pts,
			})
		}
	}
	if series, err := a.metrics.FailuresOverTime(a.ctx, projectID, days, 5); err == nil {
		for _, s := range series {
			pts := make([]LatencyPointDTO, 0, len(s.Points))
			for _, p := range s.Points {
				pts = append(pts, LatencyPointDTO{Day: p.Day.Format("2006-01-02"), Count: p.Count})
			}
			out.FailuresOverTime = append(out.FailuresOverTime, EndpointFailureSeriesDTO{
				EndpointID: s.EndpointID, Method: s.Method, Path: s.Path, Failures: s.Failures, Points: pts,
			})
		}
	}
	if cells, err := a.metrics.HourlyHeatmap(a.ctx, projectID, days); err == nil {
		for _, c := range cells {
			out.HourlyHeatmap = append(out.HourlyHeatmap, HourlyCellDTO{Day: c.Day, Hour: c.Hour, Count: c.Count})
		}
	}
	if flaky, err := a.metrics.FlakyEndpoints(a.ctx, projectID, 3); err == nil {
		for _, f := range flaky {
			out.Flaky = append(out.Flaky, FlakyEndpointDTO{
				EndpointID: f.EndpointID, Method: f.Method, Path: f.Path,
				Total: f.Total, Successes: f.Successes, Failures: f.Failures, FlakeScore: f.FlakeScore,
			})
		}
	}
	return out, nil
}

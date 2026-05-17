package app

import (
	"encoding/json"
	"fmt"
	"log"
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) RunCollection(id string) (*CollectionRunDTO, error) {
	if id == "" {
		return nil, fmt.Errorf("id required")
	}
	c, err := a.collections.Get(a.ctx, id)
	if err != nil || c == nil {
		return nil, fmt.Errorf("collection not found")
	}
	endpoints, err := a.endpoints.List(a.ctx, c.ProjectID)
	if err != nil {
		return nil, err
	}
	byID := map[string]core.Endpoint{}
	for _, e := range endpoints {
		byID[e.ID] = e
	}
	run := &CollectionRunDTO{
		CollectionID: id,
		StartedAt:    time.Now().UTC().Unix(),
		Items:        make([]CollectionRunItemDTO, 0, len(c.Items)),
	}
	start := time.Now()
	skipRest := false
	total := len(c.Items)
	runtime.EventsEmit(a.ctx, "collection:run:start", map[string]any{
		"collectionID": id,
		"total":        total,
	})
	for idx, it := range c.Items {
		ep, ok := byID[it.EndpointID]
		if !ok {
			missing := CollectionRunItemDTO{
				EndpointID: it.EndpointID,
				Skipped:    true,
				Error:      "endpoint not found",
			}
			run.Items = append(run.Items, missing)
			run.SkipCount++
			runtime.EventsEmit(a.ctx, "collection:run:progress", map[string]any{
				"collectionID": id, "index": idx, "total": total, "item": missing,
			})
			continue
		}
		if skipRest {
			skipped := CollectionRunItemDTO{
				EndpointID: it.EndpointID,
				Method:     string(ep.Method),
				Path:       ep.Path,
				Skipped:    true,
				Error:      "skipped due to previous failure",
			}
			run.Items = append(run.Items, skipped)
			run.SkipCount++
			runtime.EventsEmit(a.ctx, "collection:run:progress", map[string]any{
				"collectionID": id, "index": idx, "total": total, "item": skipped,
			})
			continue
		}
		headers := map[string]string{}
		if it.HeadersOverride != "" {
			_ = json.Unmarshal([]byte(it.HeadersOverride), &headers)
		}
		bodies := []string{it.BodyOverride}
		if it.IterateDataset {
			rowsJSON, derr := a.datasets.Get(a.ctx, c.ProjectID, endpointTestKey(string(ep.Method), ep.Path))
			if derr == nil {
				var rows []json.RawMessage
				if json.Unmarshal([]byte(rowsJSON), &rows) == nil && len(rows) > 0 {
					bodies = make([]string, 0, len(rows))
					for _, r := range rows {
						bodies = append(bodies, string(r))
					}
				}
			}
		}
		anyFailed := false
		for bIdx, body := range bodies {
			input := ExecuteRequestInput{
				ProjectID:  c.ProjectID,
				EndpointID: ep.ID,
				Method:     string(ep.Method),
				Path:       ep.Path,
				Headers:    headers,
				Body:       body,
			}
			resp, sendErr := a.ExecuteRequest(input)
			item := CollectionRunItemDTO{
				EndpointID: ep.ID,
				Method:     string(ep.Method),
				Path:       ep.Path,
			}
			if len(bodies) > 1 {
				item.Path = fmt.Sprintf("%s [#%d]", ep.Path, bIdx+1)
			}
			if sendErr != nil {
				item.Error = sendErr.Error()
				item.Pass = false
			} else if resp != nil {
				item.Status = resp.Status
				item.DurationMs = int(resp.DurationMs)
				results := a.runTestsForRequest(input, resp)
				passed := true
				for _, r := range results {
					item.TestResults = append(item.TestResults, TestResultDTO{
						ID:      r.ID,
						Name:    r.Name,
						Kind:    r.Kind,
						Pass:    r.Pass,
						Message: r.Message,
					})
					if !r.Pass {
						passed = false
					}
				}
				item.Pass = passed && resp.Status < 400
			}
			if item.Pass {
				run.PassCount++
			} else {
				run.FailCount++
				anyFailed = true
			}
			run.Items = append(run.Items, item)
			runtime.EventsEmit(a.ctx, "collection:run:progress", map[string]any{
				"collectionID": id,
				"index":        idx,
				"total":        total,
				"item":         item,
			})
		}
		if anyFailed && it.SkipOnFailure {
			skipRest = true
		}
	}
	run.DurationMs = int(time.Since(start).Milliseconds())
	a.persistCollectionRun(id, run)
	runtime.EventsEmit(a.ctx, "collection:run:done", run)
	return run, nil
}

func (a *App) persistCollectionRun(collectionID string, run *CollectionRunDTO) {
	buf, err := json.Marshal(run)
	if err != nil {
		log.Printf("collection run marshal: %v", err)
		return
	}
	now := time.Now().UTC()
	_, err = a.storage.DB.NewRaw(
		`INSERT INTO collection_runs (collection_id, run_json, started_at, duration_ms, pass_count, fail_count, skip_count, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(collection_id) DO UPDATE SET
		   run_json = excluded.run_json,
		   started_at = excluded.started_at,
		   duration_ms = excluded.duration_ms,
		   pass_count = excluded.pass_count,
		   fail_count = excluded.fail_count,
		   skip_count = excluded.skip_count,
		   updated_at = excluded.updated_at`,
		collectionID, string(buf), time.Unix(run.StartedAt, 0).UTC(), run.DurationMs,
		run.PassCount, run.FailCount, run.SkipCount, now, now,
	).Exec(a.ctx)
	if err != nil {
		log.Printf("collection run persist: %v", err)
	}
}

func (a *App) GetLastCollectionRun(collectionID string) (*CollectionRunDTO, error) {
	if collectionID == "" {
		return nil, nil
	}
	var runJSON string
	err := a.storage.DB.NewRaw(
		`SELECT run_json FROM collection_runs WHERE collection_id = ?`,
		collectionID,
	).Scan(a.ctx, &runJSON)
	if err != nil {
		return nil, nil
	}
	if runJSON == "" {
		return nil, nil
	}
	var run CollectionRunDTO
	if err := json.Unmarshal([]byte(runJSON), &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (a *App) ListLastCollectionRuns(projectID string) (map[string]*CollectionRunDTO, error) {
	if projectID == "" {
		return map[string]*CollectionRunDTO{}, nil
	}
	type row struct {
		CollectionID string `bun:"collection_id"`
		RunJSON      string `bun:"run_json"`
	}
	var rows []row
	err := a.storage.DB.NewRaw(
		`SELECT cr.collection_id, cr.run_json FROM collection_runs cr
		 JOIN collections c ON c.id = cr.collection_id
		 WHERE c.project_id = ?`,
		projectID,
	).Scan(a.ctx, &rows)
	if err != nil {
		return map[string]*CollectionRunDTO{}, nil
	}
	out := map[string]*CollectionRunDTO{}
	for _, r := range rows {
		var run CollectionRunDTO
		if err := json.Unmarshal([]byte(r.RunJSON), &run); err == nil {
			out[r.CollectionID] = &run
		}
	}
	return out, nil
}

func collectionToDTO(c domain.Collection) CollectionDTO {
	items := make([]CollectionItemDTO, 0, len(c.Items))
	for _, it := range c.Items {
		items = append(items, CollectionItemDTO{
			ID:              it.ID,
			EndpointID:      it.EndpointID,
			BodyOverride:    it.BodyOverride,
			HeadersOverride: it.HeadersOverride,
			SkipOnFailure:   it.SkipOnFailure,
			IterateDataset:  it.IterateDataset,
		})
	}
	return CollectionDTO{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		Name:        c.Name,
		Description: c.Description,
		SortOrder:   c.SortOrder,
		Items:       items,
	}
}

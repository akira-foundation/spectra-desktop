package app

import (
	"encoding/json"
	"time"
)

type SnapshotSummary struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"projectID"`
	EndpointCount int       `json:"endpointCount"`
	ScannedAt     time.Time `json:"scannedAt"`
	Added         int       `json:"added"`
	Removed       int       `json:"removed"`
	Changed       int       `json:"changed"`
}

type SnapshotDiffEntry struct {
	Method   string            `json:"method"`
	Path     string            `json:"path"`
	Kind     string            `json:"kind"`
	Changes  []string          `json:"changes,omitempty"`
	AuthRole string            `json:"authRole,omitempty"`
	Handler  string            `json:"handler,omitempty"`
	Previous *snapshotEndpoint `json:"previous,omitempty"`
	Current  *snapshotEndpoint `json:"current,omitempty"`
}

type SnapshotDiff struct {
	ID         string              `json:"id"`
	ScannedAt  time.Time           `json:"scannedAt"`
	PreviousID string              `json:"previousID,omitempty"`
	Added      []SnapshotDiffEntry `json:"added"`
	Removed    []SnapshotDiffEntry `json:"removed"`
	Changed    []SnapshotDiffEntry `json:"changed"`
}

func (a *App) ListSnapshots(projectID string, limit int) ([]SnapshotSummary, error) {
	if projectID == "" {
		return []SnapshotSummary{}, nil
	}
	if limit <= 0 {
		limit = 50
	}
	snaps, err := a.snapshots.List(a.ctx, projectID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SnapshotSummary, 0, len(snaps))
	for i, s := range snaps {
		summary := SnapshotSummary{
			ID:            s.ID,
			ProjectID:     s.ProjectID,
			EndpointCount: s.EndpointCount,
			ScannedAt:     s.ScannedAt,
		}
		if i+1 < len(snaps) {
			diff, err := computeDiff(snaps[i+1].PayloadJSON, getSnapshotPayload(a, s.ID))
			if err == nil {
				summary.Added = len(diff.Added)
				summary.Removed = len(diff.Removed)
				summary.Changed = len(diff.Changed)
			}
		}
		out = append(out, summary)
	}
	return out, nil
}

func getSnapshotPayload(a *App, id string) string {
	s, err := a.snapshots.GetByID(a.ctx, id)
	if err != nil || s == nil {
		return ""
	}
	return s.PayloadJSON
}

func (a *App) GetSnapshotDiff(snapshotID string) (*SnapshotDiff, error) {
	current, err := a.snapshots.GetByID(a.ctx, snapshotID)
	if err != nil || current == nil {
		return nil, err
	}
	previous, err := a.snapshots.Predecessor(a.ctx, current.ProjectID, current.ScannedAt)
	if err != nil {
		return nil, err
	}
	prevPayload := ""
	prevID := ""
	if previous != nil {
		prevPayload = previous.PayloadJSON
		prevID = previous.ID
	}
	diff, err := computeDiff(prevPayload, current.PayloadJSON)
	if err != nil {
		return nil, err
	}
	return &SnapshotDiff{
		ID:         current.ID,
		ScannedAt:  current.ScannedAt,
		PreviousID: prevID,
		Added:      ensureSlice(diff.Added),
		Removed:    ensureSlice(diff.Removed),
		Changed:    ensureSlice(diff.Changed),
	}, nil
}

func ensureSlice(in []SnapshotDiffEntry) []SnapshotDiffEntry {
	if in == nil {
		return []SnapshotDiffEntry{}
	}
	return in
}

func computeDiff(previousJSON, currentJSON string) (struct {
	Added   []SnapshotDiffEntry
	Removed []SnapshotDiffEntry
	Changed []SnapshotDiffEntry
}, error) {
	var out struct {
		Added   []SnapshotDiffEntry
		Removed []SnapshotDiffEntry
		Changed []SnapshotDiffEntry
	}
	prev, err := decodeSnapshotPayload(previousJSON)
	if err != nil {
		return out, err
	}
	cur, err := decodeSnapshotPayload(currentJSON)
	if err != nil {
		return out, err
	}
	prevMap := indexSnapshot(prev)
	curMap := indexSnapshot(cur)

	for key, ep := range curMap {
		old, exists := prevMap[key]
		if !exists {
			added := snapshotDiffEntry(ep, "added", nil)
			cur := ep
			added.Current = &cur
			out.Added = append(out.Added, added)
			continue
		}
		changes := compareEndpoint(old, ep)
		if len(changes) > 0 {
			entry := snapshotDiffEntry(ep, "changed", changes)
			prev := old
			cur := ep
			entry.Previous = &prev
			entry.Current = &cur
			out.Changed = append(out.Changed, entry)
		}
	}
	for key, ep := range prevMap {
		if _, exists := curMap[key]; !exists {
			removed := snapshotDiffEntry(ep, "removed", nil)
			prev := ep
			removed.Previous = &prev
			out.Removed = append(out.Removed, removed)
		}
	}
	return out, nil
}

func decodeSnapshotPayload(s string) ([]snapshotEndpoint, error) {
	if s == "" {
		return nil, nil
	}
	var items []snapshotEndpoint
	if err := json.Unmarshal([]byte(s), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func indexSnapshot(items []snapshotEndpoint) map[string]snapshotEndpoint {
	out := make(map[string]snapshotEndpoint, len(items))
	for _, ep := range items {
		key := ep.Method + " " + ep.Path
		out[key] = ep
	}
	return out
}

func compareEndpoint(a, b snapshotEndpoint) []string {
	changes := []string{}
	if a.Handler != b.Handler {
		changes = append(changes, "handler")
	}
	if a.AuthRole != b.AuthRole {
		changes = append(changes, "authRole")
	}
	if a.SchemaHash != b.SchemaHash {
		changes = append(changes, "schema")
	}
	if !stringSliceEqual(a.Middleware, b.Middleware) {
		changes = append(changes, "middleware")
	}
	return changes
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func snapshotDiffEntry(ep snapshotEndpoint, kind string, changes []string) SnapshotDiffEntry {
	return SnapshotDiffEntry{
		Method:   ep.Method,
		Path:     ep.Path,
		Kind:     kind,
		Changes:  changes,
		AuthRole: ep.AuthRole,
		Handler:  ep.Handler,
	}
}

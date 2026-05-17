package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
	"time"
)

type snapshotEndpoint struct {
	Method       string                `json:"method"`
	Path         string                `json:"path"`
	Handler      string                `json:"handler,omitempty"`
	Middleware   []string              `json:"middleware,omitempty"`
	AuthRole     string                `json:"authRole,omitempty"`
	SchemaHash   string                `json:"schemaHash,omitempty"`
	SchemaFields []snapshotSchemaField `json:"schemaFields,omitempty"`
}

type snapshotSchemaField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
}

func (a *App) recordSnapshot(projectID string, endpoints []core.Endpoint) error {
	items := make([]snapshotEndpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, snapshotEndpoint{
			Method:       string(ep.Method),
			Path:         ep.Path,
			Handler:      ep.Handler,
			Middleware:   ep.Middleware,
			AuthRole:     string(ep.AuthRole),
			SchemaHash:   stableSchemaHash(ep.RequestSchema),
			SchemaFields: extractSchemaFields(ep.RequestSchema),
		})
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	hash := hashString(string(payload))

	latest, _ := a.snapshots.Latest(a.ctx, projectID)
	if latest != nil && latest.Hash == hash {
		return nil
	}

	now := time.Now().UTC()
	snapshot := domain.EndpointSnapshot{
		ProjectID:     projectID,
		Hash:          hash,
		PayloadJSON:   string(payload),
		EndpointCount: len(endpoints),
		ScannedAt:     now,
	}
	if err := a.snapshots.Save(a.ctx, snapshot); err != nil {
		return err
	}
	return a.snapshots.TrimOldest(a.ctx, projectID, 50)
}

func hashString(s string) string {
	if s == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func extractSchemaFields(raw string) []snapshotSchemaField {
	if raw == "" {
		return nil
	}
	var s struct {
		Fields []struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return nil
	}
	out := make([]snapshotSchemaField, 0, len(s.Fields))
	for _, f := range s.Fields {
		out = append(out, snapshotSchemaField{Name: f.Name, Type: f.Type, Required: f.Required})
	}
	return out
}

// stableSchemaHash strips per-scan noise (gofakeit examples) so a
// schema with the same shape produces the same hash across scans.
func stableSchemaHash(raw string) string {
	if raw == "" {
		return ""
	}
	var s struct {
		Source     string `json:"source"`
		Confidence string `json:"confidence"`
		Fields     []struct {
			Name     string   `json:"name"`
			Type     string   `json:"type"`
			Required bool     `json:"required"`
			Rules    []string `json:"rules,omitempty"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return hashString(raw)
	}
	canonical, err := json.Marshal(s)
	if err != nil {
		return hashString(raw)
	}
	return hashString(string(canonical))
}

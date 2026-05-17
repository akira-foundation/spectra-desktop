package app

import (
	"encoding/json"
	"fmt"
	"spectra-desktop/internal/core"
	"strings"
)

type RegenerateFieldInput struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Rules []string `json:"rules,omitempty"`
}

type RegenerateBodyInput struct {
	ProjectID string                 `json:"projectID,omitempty"`
	Body      string                 `json:"body"`
	Fields    []RegenerateFieldInput `json:"fields,omitempty"`
}

func (a *App) RegenerateExampleBody(endpointID string) (string, error) {
	if endpointID == "" {
		return "{}", nil
	}
	ep, err := a.endpoints.GetByID(a.ctx, endpointID)
	if err != nil || ep == nil || ep.RequestSchema == "" {
		return "{}", err
	}
	var raw struct {
		Fields []RegenerateFieldInput `json:"fields"`
	}
	if err := json.Unmarshal([]byte(ep.RequestSchema), &raw); err != nil {
		return "{}", err
	}
	pid, _ := a.endpoints.ProjectIDOf(a.ctx, endpointID)
	return a.regenerateFromFields(raw.Fields, pid)
}

func (a *App) RegenerateBodyValues(input RegenerateBodyInput) (string, error) {
	body := strings.TrimSpace(input.Body)
	if body == "" || body == "{}" {
		return a.regenerateFromFields(input.Fields, input.ProjectID)
	}
	current := orderedMap{}
	if err := json.Unmarshal([]byte(body), &current); err != nil {
		return a.regenerateFromFields(input.Fields, input.ProjectID)
	}
	if len(current.Keys) == 0 {
		return a.regenerateFromFields(input.Fields, input.ProjectID)
	}
	fieldByName := map[string]RegenerateFieldInput{}
	for _, f := range input.Fields {
		fieldByName[f.Name] = f
	}
	out := orderedMap{}
	gen := a.bodyValueGen(input.ProjectID)
	for _, key := range current.Keys {
		oldVal := current.Values[key]
		var inferredType string
		var rules []string
		if f, ok := fieldByName[key]; ok {
			inferredType = f.Type
			rules = f.Rules
		} else {
			inferredType = inferTypeFromValue(oldVal)
		}
		out.Set(key, gen.GenerateValue(key, inferredType, rules))
	}
	return marshalOrdered(out)
}

func (a *App) regenerateFromFields(fields []RegenerateFieldInput, projectID string) (string, error) {
	if len(fields) == 0 {
		return "{}", nil
	}
	gen := a.bodyValueGen(projectID)
	out := orderedMap{}
	for _, f := range fields {
		out.Set(f.Name, gen.GenerateValue(f.Name, f.Type, f.Rules))
	}
	return marshalOrdered(out)
}

func (a *App) bodyValueGen(projectID string) core.BodyValueGen {
	driver := a.driverForProject(projectID)
	if driver != nil {
		if gen, ok := driver.(core.BodyValueGen); ok {
			return gen
		}
	}
	return fallbackBodyValueGen{}
}

func (a *App) driverForProject(projectID string) core.FrameworkDriver {
	if projectID == "" {
		return nil
	}
	project, err := a.projects.GetByID(a.ctx, projectID)
	if err != nil || project == nil {
		return nil
	}
	driver, err := a.scanner.ResolveByName(project.Framework)
	if err != nil {
		return nil
	}
	return driver
}

type fallbackBodyValueGen struct{}

func (fallbackBodyValueGen) GenerateValue(_, fieldType string, _ []string) any {
	switch strings.ToLower(fieldType) {
	case "integer", "int", "number", "numeric":
		return 0
	case "boolean", "bool":
		return false
	case "array":
		return []any{}
	case "object", "map":
		return map[string]any{}
	}
	return ""
}

type orderedMap struct {
	Keys   []string
	Values map[string]any
}

func (m *orderedMap) Set(k string, v any) {
	if m.Values == nil {
		m.Values = map[string]any{}
	}
	if _, exists := m.Values[k]; !exists {
		m.Keys = append(m.Keys, k)
	}
	m.Values[k] = v
}

func (m *orderedMap) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected object")
	}
	m.Values = map[string]any{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("expected string key")
		}
		var v any
		if err := dec.Decode(&v); err != nil {
			return err
		}
		m.Keys = append(m.Keys, key)
		m.Values[key] = v
	}
	return nil
}

func marshalOrdered(m orderedMap) (string, error) {
	var b strings.Builder
	b.WriteString("{\n")
	for i, key := range m.Keys {
		if i > 0 {
			b.WriteString(",\n")
		}
		b.WriteString("  ")
		keyJSON, _ := json.Marshal(key)
		b.Write(keyJSON)
		b.WriteString(": ")
		valJSON, err := json.MarshalIndent(m.Values[key], "  ", "  ")
		if err != nil {
			return "", err
		}
		b.Write(valJSON)
	}
	b.WriteString("\n}")
	return b.String(), nil
}

func inferTypeFromValue(v any) string {
	switch t := v.(type) {
	case bool:
		return "boolean"
	case json.Number:
		if _, err := t.Int64(); err == nil {
			return "integer"
		}
		return "numeric"
	case float64:
		return "numeric"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	case nil:
		return "string"
	default:
		return "string"
	}
}

package openapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"spectra-desktop/internal/core"
	"spectra-desktop/internal/domain"
)

type Spec struct {
	OpenAPI string                 `json:"openapi"`
	Info    Info                   `json:"info"`
	Servers []Server               `json:"servers,omitempty"`
	Paths   map[string]PathItem    `json:"paths"`
	Components *Components         `json:"components,omitempty"`
}

type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Options *Operation `json:"options,omitempty"`
}

type Operation struct {
	OperationID string                  `json:"operationId,omitempty"`
	Summary     string                  `json:"summary,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
	Parameters  []Parameter             `json:"parameters,omitempty"`
	RequestBody *RequestBody            `json:"requestBody,omitempty"`
	Responses   map[string]Response     `json:"responses"`
	Security    []map[string][]string   `json:"security,omitempty"`
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Required    bool    `json:"required"`
	Description string  `json:"description,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

type RequestBody struct {
	Required bool                   `json:"required,omitempty"`
	Content  map[string]MediaType   `json:"content"`
}

type MediaType struct {
	Schema  *Schema     `json:"schema,omitempty"`
	Example interface{} `json:"example,omitempty"`
}

type Response struct {
	Description string                 `json:"description"`
	Content     map[string]MediaType   `json:"content,omitempty"`
}

type Schema struct {
	Type       string             `json:"type,omitempty"`
	Format     string             `json:"format,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Required   []string           `json:"required,omitempty"`
	Items      *Schema            `json:"items,omitempty"`
	Example    interface{}        `json:"example,omitempty"`
}

type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

func Build(project *domain.Project, endpoints []core.Endpoint) *Spec {
	spec := &Spec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:   project.Name,
			Version: "1.0.0",
		},
		Paths: map[string]PathItem{},
		Components: &Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
				},
			},
		},
	}
	if project.BaseURL != "" {
		spec.Servers = []Server{{URL: project.BaseURL}}
	}

	for _, ep := range endpoints {
		oasPath := toOASPath(ep.Path)
		item := spec.Paths[oasPath]

		op := buildOperation(ep)
		switch strings.ToUpper(string(ep.Method)) {
		case "GET":
			item.Get = op
		case "POST":
			item.Post = op
		case "PUT":
			item.Put = op
		case "PATCH":
			item.Patch = op
		case "DELETE":
			item.Delete = op
		case "HEAD":
			item.Head = op
		case "OPTIONS":
			item.Options = op
		default:
			continue
		}
		spec.Paths[oasPath] = item
	}
	return spec
}

func ToJSON(spec *Spec) (string, error) {
	out, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func toOASPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			continue
		}
		if strings.HasPrefix(p, ":") {
			parts[i] = "{" + p[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func buildOperation(ep core.Endpoint) *Operation {
	op := &Operation{
		OperationID: ep.Name,
		Summary:     ep.Name,
		Responses: map[string]Response{
			"200": {Description: "Successful response"},
		},
	}
	if ep.Tags != nil {
		op.Tags = ep.Tags
	}
	for _, param := range extractPathParams(ep.Path) {
		op.Parameters = append(op.Parameters, Parameter{
			Name:     param,
			In:       "path",
			Required: true,
			Schema:   &Schema{Type: "string"},
		})
	}
	if isAuthRequired(ep) {
		op.Security = []map[string][]string{{"bearerAuth": {}}}
	}
	if schema := buildRequestSchema(ep); schema != nil && allowsBody(string(ep.Method)) {
		op.RequestBody = &RequestBody{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {Schema: schema},
			},
		}
	}
	return op
}

func extractPathParams(path string) []string {
	var out []string
	for _, p := range strings.Split(path, "/") {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			out = append(out, p[1:len(p)-1])
			continue
		}
		if strings.HasPrefix(p, ":") {
			out = append(out, p[1:])
		}
	}
	return out
}

type rawField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type rawSchema struct {
	Fields []rawField `json:"fields"`
}

func buildRequestSchema(ep core.Endpoint) *Schema {
	if ep.RequestSchema == "" {
		return nil
	}
	var raw rawSchema
	if err := json.Unmarshal([]byte(ep.RequestSchema), &raw); err != nil {
		return nil
	}
	if len(raw.Fields) == 0 {
		return nil
	}
	props := map[string]*Schema{}
	required := []string{}
	for _, f := range raw.Fields {
		props[f.Name] = mapType(f.Type)
		if f.Required {
			required = append(required, f.Name)
		}
	}
	return &Schema{
		Type:       "object",
		Properties: props,
		Required:   required,
	}
}

func mapType(t string) *Schema {
	switch t {
	case "integer", "numeric":
		return &Schema{Type: "integer"}
	case "boolean":
		return &Schema{Type: "boolean"}
	case "array":
		return &Schema{Type: "array", Items: &Schema{Type: "string"}}
	case "object":
		return &Schema{Type: "object"}
	case "email":
		return &Schema{Type: "string", Format: "email"}
	case "date":
		return &Schema{Type: "string", Format: "date"}
	case "uuid":
		return &Schema{Type: "string", Format: "uuid"}
	case "url":
		return &Schema{Type: "string", Format: "uri"}
	default:
		return &Schema{Type: "string"}
	}
}

func isAuthRequired(ep core.Endpoint) bool {
	for _, m := range ep.Middleware {
		lower := strings.ToLower(m)
		if strings.HasPrefix(lower, "auth") || strings.Contains(lower, "authenticate") {
			return true
		}
	}
	return false
}

func allowsBody(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "DELETE", "OPTIONS":
		return false
	default:
		return true
	}
}

// keep fmt for future yaml support
var _ = fmt.Sprintf

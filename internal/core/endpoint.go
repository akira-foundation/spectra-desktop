package core

type HTTPMethod string

const (
	MethodGet     HTTPMethod = "GET"
	MethodPost    HTTPMethod = "POST"
	MethodPut     HTTPMethod = "PUT"
	MethodPatch   HTTPMethod = "PATCH"
	MethodDelete  HTTPMethod = "DELETE"
	MethodHead    HTTPMethod = "HEAD"
	MethodOptions HTTPMethod = "OPTIONS"
)

type Endpoint struct {
	ID         string            `json:"id"`
	Method     HTTPMethod        `json:"method"`
	Path       string            `json:"path"`
	Name       string            `json:"name,omitempty"`
	Handler    string            `json:"handler,omitempty"`
	Middleware []string          `json:"middleware,omitempty"`
	Parameters []Parameter       `json:"parameters,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Source     EndpointSource    `json:"source"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Framework  string            `json:"framework,omitempty"`
	Confidence float64           `json:"confidence,omitempty"`
}

type Parameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Type     string `json:"type,omitempty"`
	Required bool   `json:"required"`
}

type EndpointSource struct {
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"`
}

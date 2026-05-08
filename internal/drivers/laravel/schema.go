package laravel

type SchemaSource string

const (
	SchemaSourceNone        SchemaSource = ""
	SchemaSourceFormRequest SchemaSource = "form_request"
	SchemaSourceInline      SchemaSource = "inline_validation"
)

type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "high"
	ConfidenceMedium ConfidenceLevel = "medium"
	ConfidenceLow    ConfidenceLevel = "low"
)

type RequestSchema struct {
	Source     SchemaSource    `json:"source"`
	Confidence ConfidenceLevel `json:"confidence"`
	Fields     []InferredField `json:"fields"`
}

type InferredField struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Rules    []string `json:"rules,omitempty"`
	Example  any      `json:"example,omitempty"`
}

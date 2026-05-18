package laravel

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchemaConstants(t *testing.T) {
	if SchemaSourceNone != "" || SchemaSourceFormRequest != "form_request" || SchemaSourceInline != "inline_validation" {
		t.Fatal("schema source constants drifted")
	}
	if ConfidenceHigh != "high" || ConfidenceMedium != "medium" || ConfidenceLow != "low" {
		t.Fatal("confidence constants drifted")
	}
}

func TestRequestSchema_JSONShape(t *testing.T) {
	s := RequestSchema{
		Source:     SchemaSourceFormRequest,
		Confidence: ConfidenceHigh,
		Fields: []InferredField{
			{Name: "email", Type: "email", Required: true, Rules: []string{"required", "email"}, Example: "e@x"},
		},
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	out := string(b)
	for _, want := range []string{`"source":"form_request"`, `"confidence":"high"`, `"name":"email"`, `"required":true`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in %s", want, out)
		}
	}
}

func TestInferredField_OmitsEmpty(t *testing.T) {
	b, err := json.Marshal(InferredField{Name: "x", Type: "string"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "rules") || strings.Contains(string(b), "example") {
		t.Fatalf("unexpected fields: %s", string(b))
	}
}

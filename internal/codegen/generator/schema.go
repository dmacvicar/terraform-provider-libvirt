package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"
)

// SchemaGenerator generates Terraform Plugin Framework schemas.
type SchemaGenerator struct {
	template *template.Template
}

// NewSchemaGenerator creates a new schema generator.
func NewSchemaGenerator(templatePath string) (*SchemaGenerator, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &SchemaGenerator{
		template: tmpl,
	}, nil
}

// NewSchemaGeneratorFromString creates a generator from a template string.
func NewSchemaGeneratorFromString(templateContent string) (*SchemaGenerator, error) {
	tmpl, err := template.New("schema").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &SchemaGenerator{
		template: tmpl,
	}, nil
}

// Generate generates schema code for the given structs.
func (g *SchemaGenerator) Generate(structs []*StructIR) (string, error) {
	var buf bytes.Buffer

	data := map[string]interface{}{
		"Structs": structs,
	}
	if err := g.template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("formatting generated code: %w", err)
	}

	return string(formatted), nil
}

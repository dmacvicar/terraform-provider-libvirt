package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"
)

// ModelGenerator generates Terraform model structs.
type ModelGenerator struct {
	template *template.Template
}

// NewModelGenerator creates a new model generator.
func NewModelGenerator(templatePath string) (*ModelGenerator, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &ModelGenerator{
		template: tmpl,
	}, nil
}

// NewModelGeneratorFromString creates a generator from a template string.
func NewModelGeneratorFromString(templateContent string) (*ModelGenerator, error) {
	tmpl, err := template.New("model").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &ModelGenerator{
		template: tmpl,
	}, nil
}

// Generate generates model code for the given structs.
func (g *ModelGenerator) Generate(structs []*StructIR) (string, error) {
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

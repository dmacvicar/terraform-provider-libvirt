package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"text/template"
)

// ConvertGenerator generates conversion code between XML and Terraform models.
type ConvertGenerator struct {
	template *template.Template
}

// NewConvertGenerator creates a new conversion generator.
func NewConvertGenerator(templatePath string) (*ConvertGenerator, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &ConvertGenerator{
		template: tmpl,
	}, nil
}

// NewConvertGeneratorFromString creates a generator from a template string.
func NewConvertGeneratorFromString(templateContent string) (*ConvertGenerator, error) {
	tmpl, err := template.New("convert").Parse(templateContent)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return &ConvertGenerator{
		template: tmpl,
	}, nil
}

// Generate generates conversion code for the given structs.
func (g *ConvertGenerator) Generate(structs []*StructIR) (string, error) {
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

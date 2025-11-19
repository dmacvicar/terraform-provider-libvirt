package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docregistry"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/parser"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/util/stringutil"
	"libvirt.org/go/libvirtxml"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	outputDir := "internal/generated"

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create reflector
	reflector := parser.NewLibvirtXMLReflector()

	// Define all top-level libvirt resources to generate
	resources := []struct {
		Name string
		Type reflect.Type
	}{
		{"domain", reflect.TypeOf(libvirtxml.Domain{})},
		{"network", reflect.TypeOf(libvirtxml.Network{})},
		{"storage_pool", reflect.TypeOf(libvirtxml.StoragePool{})},
		{"storage_volume", reflect.TypeOf(libvirtxml.StorageVolume{})},
	}

	// Collect all structs from all resources (deduplicated)
	allStructs := make(map[string]*generator.StructIR)

	type resourceIR struct {
		Name string
		Root *generator.StructIR
	}
	var resourceIRs []resourceIR

	for _, resource := range resources {
		// Analyze the struct
		rootIR, err := reflector.ReflectStruct(resource.Type)
		if err != nil {
			return fmt.Errorf("reflecting %s: %w", resource.Name, err)
		}

		// Mark as top-level resource
		rootIR.IsTopLevel = true

		// Collect all structs including nested ones
		structs := collectAllStructs(rootIR)

		// Add to deduplicated map
		for _, s := range structs {
			if existing, exists := allStructs[s.Name]; !exists {
				allStructs[s.Name] = s
			} else if s.IsTopLevel {
				// If this version is top-level, preserve that flag
				existing.IsTopLevel = true
			}
		}
		resourceIRs = append(resourceIRs, resourceIR{Name: resource.Name, Root: rootIR})
	}

	// Load documentation registry and apply descriptions before rendering templates
	docReg, err := docregistry.Load("internal/codegen/docs")
	if err != nil {
		return fmt.Errorf("loading schema docs: %w", err)
	}
	for _, res := range resourceIRs {
		docReg.Apply(stringutil.SnakeCase(res.Name), res.Root)
	}

	// Convert map to slice
	structs := make([]*generator.StructIR, 0, len(allStructs))
	totalFields := 0
	topLevelCount := 0
	for _, s := range allStructs {
		structs = append(structs, s)
		totalFields += len(s.Fields)
		if s.IsTopLevel {
			topLevelCount++
		}
	}

	// Silently generate code

	// Generate one file per struct
	for _, s := range structs {
		fileName := stringutil.SnakeCase(s.Name)

		if err := generateModel([]*generator.StructIR{s}, outputDir, fileName); err != nil {
			return fmt.Errorf("generating model for %s: %w", s.Name, err)
		}

		if err := generateSchema([]*generator.StructIR{s}, outputDir, fileName); err != nil {
			return fmt.Errorf("generating schema for %s: %w", s.Name, err)
		}

		if err := generateConvert([]*generator.StructIR{s}, outputDir, fileName); err != nil {
			return fmt.Errorf("generating conversions for %s: %w", s.Name, err)
		}
	}

	fmt.Printf("Generated %d structs (%d top-level, %d total fields) in %s/\n", len(structs), topLevelCount, totalFields, outputDir)

	return nil
}

func generateModel(structs []*generator.StructIR, outputDir, resourceName string) error {
	templatePath := "internal/codegen/templates/model.go.tmpl"
	gen, err := generator.NewModelGenerator(templatePath)
	if err != nil {
		return err
	}

	code, err := gen.Generate(structs)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_model.gen.go", resourceName))
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("writing model file: %w", err)
	}

	return nil
}

func generateSchema(structs []*generator.StructIR, outputDir, resourceName string) error {
	templatePath := "internal/codegen/templates/schema.go.tmpl"
	gen, err := generator.NewSchemaGenerator(templatePath)
	if err != nil {
		return err
	}

	code, err := gen.Generate(structs)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_schema.gen.go", resourceName))
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("writing schema file: %w", err)
	}

	return nil
}

func generateConvert(structs []*generator.StructIR, outputDir, resourceName string) error {
	templatePath := "internal/codegen/templates/convert.go.tmpl"
	gen, err := generator.NewConvertGenerator(templatePath)
	if err != nil {
		return err
	}

	code, err := gen.Generate(structs)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_convert.gen.go", resourceName))
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("writing convert file: %w", err)
	}

	return nil
}

// collectAllStructs recursively collects all structs including nested ones.
func collectAllStructs(root *generator.StructIR) []*generator.StructIR {
	result := []*generator.StructIR{}
	seen := make(map[string]bool)

	var collect func(*generator.StructIR)
	collect = func(s *generator.StructIR) {
		if seen[s.Name] {
			return
		}
		seen[s.Name] = true
		result = append(result, s)

		// Collect nested structs
		for _, field := range s.Fields {
			if field.IsNested && field.NestedStruct != nil {
				collect(field.NestedStruct)
			}
		}
	}

	collect(root)
	return result
}

package docgen

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docindex"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/docregistry"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/parser"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/util/stringutil"
	"libvirt.org/go/libvirtxml"
)

// Batch represents a group of fields to document together
type Batch struct {
	Fields []FieldContext // Fields to document with their XML context
}

// FieldContext provides context for a field to be documented
type FieldContext struct {
	TFPath            string   // Terraform path (e.g., "domain.memory")
	XMLPath           string   // XML path (e.g., "domain.memory")
	Optional          bool     // Optional in schema
	Required          bool     // Required in schema
	Computed          bool     // Computed in schema
	PresenceBoolean   bool     // Presence-only boolean
	StringToBool      bool     // Yes/no style flags mapped to bool
	StringToBoolTrue  string   // String value for true
	StringToBoolFalse string   // String value for false
	FlattenedAttr     string   // If flattened attribute (e.g., unit/placement)
	ValidValues       []string // Enumerated choices if known
}

// State tracks progress through batch generation
type State struct {
	LastCompletedBatch int    `json:"last_completed_batch"`
	Timestamp          string `json:"timestamp"`
}

// LoadState loads generation state from disk
func LoadState(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &State{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state JSON: %w", err)
	}

	return &state, nil
}

// SaveState saves generation state to disk
func SaveState(state *State, path string) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}

	return nil
}

// BuildBatches creates batches of all fields
func BuildBatches(existing *docregistry.Registry, batchSize int) ([]Batch, error) {
	// Reflect all libvirt resources
	reflector := parser.NewLibvirtXMLReflector()

	resources := []struct {
		Name string
		Type reflect.Type
	}{
		{"domain", reflect.TypeOf(libvirtxml.Domain{})},
		{"network", reflect.TypeOf(libvirtxml.Network{})},
		{"storage_pool", reflect.TypeOf(libvirtxml.StoragePool{})},
		{"storage_volume", reflect.TypeOf(libvirtxml.StorageVolume{})},
	}

	var allFields []FieldContext

	// Collect ALL field paths (skip already documented when registry is provided)
	for _, resource := range resources {
		rootIR, err := reflector.ReflectStruct(resource.Type)
		if err != nil {
			return nil, fmt.Errorf("reflecting %s: %w", resource.Name, err)
		}

		// Collect all paths
		rootPath := stringutil.SnakeCase(resource.Name)
		rootXMLPath := xmlRootPath(resource.Name)
		collectAllFields(rootIR, rootPath, rootXMLPath, &allFields, existing)
	}

	// Sort for deterministic output
	sort.Slice(allFields, func(i, j int) bool {
		return allFields[i].TFPath < allFields[j].TFPath
	})

	// Split into batches
	var batches []Batch
	for i := 0; i < len(allFields); i += batchSize {
		end := i + batchSize
		if end > len(allFields) {
			end = len(allFields)
		}

		batches = append(batches, Batch{
			Fields: allFields[i:end],
		})
	}

	return batches, nil
}

func xmlRootPath(resourceName string) string {
	switch resourceName {
	case "storage_pool":
		return "pool"
	case "storage_volume":
		return "volume"
	default:
		return resourceName
	}
}

func collectAllFields(s *generator.StructIR, tfRoot, xmlRoot string, result *[]FieldContext, existing *docregistry.Registry) {
	if s == nil {
		return
	}

	visited := make(map[string]bool)
	collectAllFieldsRecursive(s, tfRoot, xmlRoot, visited, result, existing)
}

func collectAllFieldsRecursive(s *generator.StructIR, tfPath, xmlPath string, visited map[string]bool, result *[]FieldContext, existing *docregistry.Registry) {
	if s == nil {
		return
	}

	for _, field := range s.Fields {
		fieldTFPath := tfPath + "." + field.TFName
		if visited[fieldTFPath] {
			continue
		}
		visited[fieldTFPath] = true

		// Build XML path
		fieldXMLPath := buildXMLPath(xmlPath, field)

		// Skip if already documented
		if existing != nil {
			if _, ok := existingEntry(existing, fieldTFPath); ok {
				// Still traverse nested to find undocumented children
				if field.IsNested && !field.IsCycle && field.NestedStruct != nil {
					collectAllFieldsRecursive(field.NestedStruct, fieldTFPath, fieldXMLPath, visited, result, existing)
				}
				continue
			}
		}

		*result = append(*result, FieldContext{
			TFPath:            fieldTFPath,
			XMLPath:           fieldXMLPath,
			Optional:          field.IsOptional,
			Required:          field.IsRequired,
			Computed:          field.IsComputed,
			PresenceBoolean:   field.IsPresenceBoolean,
			StringToBool:      field.StringToBool != nil,
			StringToBoolTrue:  stringToBoolValue(field, true),
			StringToBoolFalse: stringToBoolValue(field, false),
			FlattenedAttr:     field.FlattenedAttrName,
			ValidValues:       field.ValidValues,
		})

		// Recurse into nested structs
		if field.IsNested && !field.IsCycle && field.NestedStruct != nil {
			collectAllFieldsRecursive(field.NestedStruct, fieldTFPath, fieldXMLPath, visited, result, existing)
		}
	}
}

func existingEntry(reg *docregistry.Registry, path string) (docregistry.Entry, bool) {
	if reg == nil {
		return docregistry.Entry{}, false
	}
	return reg.EntryForPath(path)
}

func stringToBoolValue(field *generator.FieldIR, trueVal bool) string {
	if field.StringToBool == nil {
		return ""
	}
	if trueVal {
		return field.StringToBool.TrueValue
	}
	return field.StringToBool.FalseValue
}

func buildXMLPath(parent string, field *generator.FieldIR) string {
	elementName := field.XMLName
	if elementName == "" {
		elementName = field.TFName
	}

	if field.IsXMLAttr {
		return parent + ".@" + elementName
	}

	return parent + "." + elementName
}

// GetFullIndex loads and returns the complete documentation index for AI context
func GetFullIndex(indexFile string) (docindex.Index, error) {
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("reading index file: %w", err)
	}

	var index docindex.Index
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("parsing index JSON: %w", err)
	}

	return index, nil
}

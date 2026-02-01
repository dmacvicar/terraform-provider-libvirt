package parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"
	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/util/stringutil"
)

// LibvirtXMLReflector analyzes libvirtxml structs using reflection.
type LibvirtXMLReflector struct {
	// processedTypes tracks types we've already analyzed to prevent infinite recursion
	processedTypes map[string]*generator.StructIR

	// analysisStack tracks the current path of struct analysis for cycle detection
	// e.g., ["DomainDiskSource", "DomainDiskDataStore", "DomainDiskSlice"]
	analysisStack []string
}

// NewLibvirtXMLReflector creates a new reflector.
func NewLibvirtXMLReflector() *LibvirtXMLReflector {
	return &LibvirtXMLReflector{
		processedTypes: make(map[string]*generator.StructIR),
		analysisStack:  []string{},
	}
}

// ReflectStruct analyzes a libvirtxml struct and returns its IR.
func (r *LibvirtXMLReflector) ReflectStruct(structType reflect.Type) (*generator.StructIR, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", structType.Kind())
	}

	// Check if already processed
	typeName := structType.Name()
	if ir, ok := r.processedTypes[typeName]; ok {
		return ir, nil
	}

	// Skip anonymous/empty structs
	if typeName == "" {
		return nil, fmt.Errorf("struct has no name (anonymous struct not supported)")
	}

	ir := &generator.StructIR{
		Name:      typeName,
		GoPackage: structType.PkgPath(),
		Fields:    []*generator.FieldIR{},
	}

	// Store early to handle recursive types
	r.processedTypes[typeName] = ir

	// Push onto analysis stack for cycle detection
	r.analysisStack = append(r.analysisStack, typeName)
	defer func() {
		// Pop from stack when done analyzing
		r.analysisStack = r.analysisStack[:len(r.analysisStack)-1]
	}()

	// Detect value-with-unit pattern
	pattern := r.detectValueWithUnitPattern(structType)
	if pattern != nil {
		ir.ValueWithUnitPattern = pattern
	}

	// Detect boolean-to-presence pattern (empty struct)
	boolPresence := r.detectBooleanToPresence(structType)
	if boolPresence != nil {
		ir.BooleanToPresence = boolPresence
	}

	// Analyze each field
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		if field.Anonymous {
			if field.Type.Kind() == reflect.Struct {
				embeddedIR, err := r.ReflectStruct(field.Type)
				if err != nil {
					return nil, fmt.Errorf("analyzing embedded struct %s: %w", field.Type.Name(), err)
				}
				ir.Fields = append(ir.Fields, embeddedIR.Fields...)
			}
			continue
		}

		fieldIR, err := r.analyzeField(typeName, field)
		if err != nil {
			return nil, fmt.Errorf("analyzing field %s: %w", field.Name, err)
		}

		if fieldIR != nil {
			ir.Fields = append(ir.Fields, fieldIR)
		}
	}

	// Apply universal and resource-specific patterns
	r.applyFieldPatterns(structType.Name(), ir.Fields)

	// Post-process: expand chardata+attribute fields into separate flattened fields
	// Example: <memory unit='KiB'>524288</memory> → memory (value), memory_unit (attr)
	//          <vcpu placement='static'>2</vcpu> → vcpu (value), vcpu_placement (attr)
	expandedFields := []*generator.FieldIR{}
	for _, field := range ir.Fields {
		if field.TFType == "FLATTEN_VALUE_UNIT" && field.NestedStruct != nil && field.NestedStruct.ValueWithUnitPattern != nil {
			pattern := field.NestedStruct.ValueWithUnitPattern

			// Determine TF type for the value field
			var valueTFType string
			if pattern.ValueGoType == "string" {
				valueTFType = "types.String"
			} else {
				valueTFType = "types.Int64" // uint, uint64, int, etc. all map to Int64
			}

			// Create value field (the chardata content)
			valueField := &generator.FieldIR{
				GoName:             field.GoName,
				XMLName:            field.XMLName,
				TFName:             field.TFName,
				GoType:             field.GoType,
				TFType:             valueTFType,
				IsXMLAttr:          field.IsXMLAttr,
				IsPointer:          field.IsPointer,
				IsOptional:         field.IsOptional,
				IsRequired:         field.IsRequired,
				IsComputed:         field.IsComputed,
				OmitEmpty:          field.OmitEmpty,
				PreserveUserIntent: field.PreserveUserIntent,
				NestedStruct:       field.NestedStruct, // Keep reference for conversion
				IsFlattenedValue:   true,               // Mark as flattened value
			}
			expandedFields = append(expandedFields, valueField)

			// Create a field for each attribute
			for _, attrInfo := range pattern.AttributeFields {
				// Determine TF type for the attribute
				var attrTFType string
				switch attrInfo.GoType {
				case "string":
					attrTFType = "types.String"
				case "uint", "uint64", "int", "int64":
					attrTFType = "types.Int64"
				default:
					attrTFType = "types.String" // Default to string
				}

				attrField := &generator.FieldIR{
					GoName:             field.GoName + attrInfo.GoName,                              // e.g., "MemoryUnit", "VCPUPlacement"
					XMLName:            field.XMLName,                                               // Same XML element
					TFName:             field.TFName + "_" + stringutil.SnakeCase(attrInfo.XMLName), // e.g., "memory_unit", "vcpu_placement"
					GoType:             field.GoType,
					TFType:             attrTFType,
					IsXMLAttr:          true,
					IsPointer:          field.IsPointer,
					IsOptional:         true, // Attributes are always optional
					IsRequired:         false,
					IsComputed:         false,
					OmitEmpty:          true,
					PreserveUserIntent: true,
					NestedStruct:       field.NestedStruct, // Keep reference for conversion
					IsFlattenedAttr:    true,               // Mark as flattened attribute
					FlattenedAttrName:  attrInfo.XMLName,   // Store XML attribute name
				}
				expandedFields = append(expandedFields, attrField)
			}
		} else {
			expandedFields = append(expandedFields, field)
		}
	}
	ir.Fields = expandedFields

	// Apply patterns again after field expansion
	r.applyFieldPatterns(structType.Name(), ir.Fields)

	return ir, nil
}

// applyFieldPatterns applies universal and resource-specific field patterns
func (r *LibvirtXMLReflector) applyFieldPatterns(structName string, fields []*generator.FieldIR) {
	for _, field := range fields {
		// Universal patterns (apply to ALL resources)
		switch field.TFName {
		case "uuid", "id", "key":
			field.IsComputed = true
			field.IsOptional = false
			field.IsRequired = false
			field.PlanModifier = "UseStateForUnknown"

		case "name":
			// At resource root level, name is required and immutable
			if structName == "StoragePool" || structName == "Domain" || structName == "Network" || structName == "StorageVolume" {
				field.IsRequired = true
				field.IsOptional = false
				field.PlanModifier = "RequiresReplace"
			}

		case "type":
			// At resource root level, type is required and immutable
			if structName == "StoragePool" || structName == "Domain" || structName == "Network" {
				field.IsRequired = true
				field.IsOptional = false
				field.PlanModifier = "RequiresReplace"
			}
		}

		// Resource-specific patterns
		if structName == "StoragePool" {
			// Skip unit fields - they stay optional
			if field.IsFlattenedUnit {
				continue
			}

			switch field.TFName {
			case "capacity":
				field.IsComputed = true
				field.IsOptional = false
				field.IsRequired = false
				field.PlanModifier = "UseStateForUnknown"

			case "allocation", "available":
				field.IsComputed = true
				field.IsOptional = false
				field.IsRequired = false
			}
		}

		if structName == "StorageVolume" {
			// Skip unit fields - they stay optional
			if field.IsFlattenedUnit {
				continue
			}

			switch field.TFName {
			case "capacity", "allocation", "physical":
				field.IsComputed = true
				field.IsOptional = false
				field.IsRequired = false
			}
		}
	}
}

func (r *LibvirtXMLReflector) analyzeField(structName string, field reflect.StructField) (*generator.FieldIR, error) {
	// Skip XMLName fields (used by encoding/xml)
	if field.Name == "XMLName" {
		return nil, nil
	}

	xmlTag := field.Tag.Get("xml")
	if xmlTag == "" {
		// Skip fields without XML tags
		return nil, nil
	}

	isDashTag := xmlTag == "-"

	fieldIR := &generator.FieldIR{
		GoName: field.Name,
	}

	// Parse XML tag
	parts := strings.Split(xmlTag, ",")
	fieldIR.XMLName = parts[0]
	if isDashTag {
		fieldIR.XMLName = field.Name
	}

	// Check for attributes and options
	for _, part := range parts[1:] {
		switch part {
		case "attr":
			fieldIR.IsXMLAttr = true
		case "omitempty":
			fieldIR.OmitEmpty = true
			fieldIR.IsOptional = true
		}
	}

	// Analyze Go type
	fieldType := field.Type
	fieldIR.IsPointer = fieldType.Kind() == reflect.Ptr

	if fieldIR.IsPointer {
		fieldType = fieldType.Elem()
		fieldIR.IsOptional = true
	}

	// Determine Go and TF types
	fieldIR.GoType = field.Type.String()

	// Handle slices
	if fieldType.Kind() == reflect.Slice {
		fieldIR.IsList = true
		elemType := fieldType.Elem()

		// Handle pointer to element type (e.g., []*DomainDisk)
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}

		// Store element type information
		fieldIR.ElementGoType = elemType.String()

		// Check if slice contains structs or primitives
		if elemType.Kind() == reflect.Struct {
			// Skip anonymous structs (no name)
			if elemType.Name() == "" {
				return nil, nil
			}

			// Check for circular reference
			if r.isInAnalysisStack(elemType.Name()) {
				fieldIR.IsCycle = true
				fieldIR.CycleNote = r.buildCyclePath(elemType.Name())
				// Don't continue analyzing - break the cycle
				return fieldIR, nil
			}

			fieldIR.IsNested = true
			// Recursively analyze nested struct
			nestedIR, err := r.ReflectStruct(elemType)
			if err != nil {
				return nil, fmt.Errorf("analyzing nested struct in slice %s: %w", elemType.Name(), err)
			}
			fieldIR.NestedStruct = nestedIR
			fieldIR.TFType = "types.List"
		} else {
			// Simple slice ([]string, []int, etc.)
			fieldIR.TFType = "types.List"
		}
	} else if fieldType.Kind() == reflect.Struct {
		// Handle nested structs
		// Skip anonymous structs (no name)
		if fieldType.Name() == "" {
			// Skip this field - anonymous structs not supported
			return nil, nil
		}

		// Check for circular reference
		if r.isInAnalysisStack(fieldType.Name()) {
			fieldIR.IsCycle = true
			fieldIR.CycleNote = r.buildCyclePath(fieldType.Name())
			// Don't continue analyzing - break the cycle
			return fieldIR, nil
		}

		fieldIR.IsNested = true
		// Recursively analyze nested struct
		nestedIR, err := r.ReflectStruct(fieldType)
		if err != nil {
			return nil, fmt.Errorf("analyzing nested struct %s: %w", fieldType.Name(), err)
		}
		fieldIR.NestedStruct = nestedIR

		// Check if nested struct is presence-only (boolean-to-presence pattern)
		if nestedIR.BooleanToPresence != nil && nestedIR.BooleanToPresence.IsPresenceOnly {
			// Convert to boolean - presence of element = true, absence = false
			fieldIR.TFType = "types.Bool"
			fieldIR.IsNested = false
			fieldIR.IsPresenceBoolean = true
			// Keep NestedStruct reference for type name in conversion
		} else if nestedIR.ValueWithUnitPattern != nil && nestedIR.ValueWithUnitPattern.FlattenStrategy == "flatten" {
			// Chardata+attributes pattern: flatten to multiple fields
			// This field will be handled specially - parent struct will get N+1 fields generated
			// (1 for value + N for each attribute)
			fieldIR.TFType = "FLATTEN_VALUE_UNIT" // Special marker (name kept for backwards compat)
			fieldIR.IsNested = false
		} else {
			fieldIR.TFType = "types.Object"
		}
	} else {
		// Map primitive types
		fieldIR.TFType = r.goTypeToTFType(fieldType)
	}

	// Convert to snake_case for Terraform
	fieldIR.TFName = stringutil.SnakeCase(field.Name)

	// For now, assume omitempty, pointer, or list means optional, otherwise required
	// This will be refined with RNG schema information later
	if fieldIR.OmitEmpty || fieldIR.IsPointer || fieldIR.IsList {
		fieldIR.IsOptional = true
		fieldIR.IsRequired = false
		// Optional fields should preserve user intent
		fieldIR.PreserveUserIntent = true
	} else {
		fieldIR.IsOptional = false
		fieldIR.IsRequired = true
	}

	// Detect string-to-bool pattern
	if !fieldIR.IsNested && !fieldIR.IsList {
		stringToBool := r.detectStringToBool(field.Name, fieldType.String())
		if stringToBool != nil {
			fieldIR.StringToBool = stringToBool
			// Override TF type to boolean
			fieldIR.TFType = "types.Bool"
		}
	}

	return fieldIR, nil
}

func (r *LibvirtXMLReflector) goTypeToTFType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "types.String"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "types.Int64"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "types.Int64"
	case reflect.Bool:
		return "types.Bool"
	case reflect.Float32, reflect.Float64:
		return "types.Float64"
	case reflect.Slice:
		// For slices, we'll need to check the element type
		// For now, return list of strings
		return "types.List"
	default:
		return "types.String"
	}
}

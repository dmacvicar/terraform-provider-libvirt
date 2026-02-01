package parser

import (
	"reflect"
	"strings"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"
)

// detectValueWithUnitPattern checks if a struct follows the "chardata with attributes" pattern.
//
// Pattern: Value INSIDE element (chardata), ALL other fields are attributes → FLATTEN
// Examples:
//
//	<memory unit='KiB'>524288</memory>                    → memory, memory_unit
//	<vcpu placement='static' cpuset='1-4'>2</vcpu>        → vcpu, vcpu_placement, vcpu_cpuset
//
// This pattern is ALWAYS flattened regardless of attribute count.
func (r *LibvirtXMLReflector) detectValueWithUnitPattern(structType reflect.Type) *generator.ValueWithUnitPattern {
	var chardataField *struct {
		name   string
		goType string
	}
	var attrFields []generator.AttributeFieldInfo
	hasNonAttrFields := false

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		xmlTag := field.Tag.Get("xml")

		if xmlTag == "" || xmlTag == "-" || field.Name == "XMLName" {
			continue
		}

		// Check for chardata (the value inside the element)
		if strings.Contains(xmlTag, "chardata") {
			chardataField = &struct {
				name   string
				goType string
			}{
				name:   field.Name,
				goType: field.Type.String(),
			}
			continue
		}

		// Check if it's an attribute
		if strings.Contains(xmlTag, "attr") {
			// Extract attribute name from XML tag (e.g., "unit,attr,omitempty" → "unit")
			parts := strings.Split(xmlTag, ",")
			xmlName := parts[0]

			attrFields = append(attrFields, generator.AttributeFieldInfo{
				GoName:  field.Name,
				XMLName: xmlName,
				GoType:  field.Type.String(),
			})
			continue
		}

		// If we have nested elements (not attributes), this pattern doesn't apply
		hasNonAttrFields = true
	}

	// Pattern only applies if:
	// 1. We have chardata (value inside element)
	// 2. All other fields are attributes (no nested elements)
	if chardataField == nil || hasNonAttrFields {
		return nil
	}

	return &generator.ValueWithUnitPattern{
		ValueFieldName:  chardataField.name,
		ValueGoType:     chardataField.goType,
		AttributeFields: attrFields,
		FlattenStrategy: "flatten", // Always flatten this pattern
	}
}

// detectBooleanToPresence checks if a struct is presence-only (empty struct).
// These are used in features where the boolean is represented by element presence.
func (r *LibvirtXMLReflector) detectBooleanToPresence(structType reflect.Type) *generator.BooleanToPresencePattern {
	// Count exported fields (ignoring XMLName). Presence-only structs are completely empty.
	exportedFieldCount := 0
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Name == "XMLName" {
			continue
		}
		if !field.IsExported() {
			continue
		}
		exportedFieldCount++
	}

	if exportedFieldCount == 0 {
		return &generator.BooleanToPresencePattern{
			IsPresenceOnly: true,
		}
	}

	return nil
}

// detectStringToBool detects if a string field should be boolean in TF.
// Common patterns: "yes"/"no", "on"/"off", "enabled"/"disabled"
func (r *LibvirtXMLReflector) detectStringToBool(fieldName string, goType string) *generator.StringToBoolPattern {
	// Only applies to string fields
	if goType != "string" {
		return nil
	}

	// Common field names that are yes/no booleans
	yesNoFields := map[string]bool{
		"readonly":   true,
		"ReadOnly":   true,
		"Readonly":   true,
		"autoport":   true,
		"AutoPort":   true,
		"Autoport":   true,
		"managed":    true,
		"Managed":    true,
		"migratable": true,
		"Migratable": true,
	}

	if yesNoFields[fieldName] {
		return &generator.StringToBoolPattern{
			TrueValue:  "yes",
			FalseValue: "no",
		}
	}

	return nil
}

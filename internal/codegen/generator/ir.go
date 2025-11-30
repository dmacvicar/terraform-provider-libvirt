package generator

// IR (Intermediate Representation) holds the analyzed metadata for code generation.

// StructIR represents a libvirtxml struct to be generated.
type StructIR struct {
	// Name is the libvirtxml struct name (e.g., "DomainTimer")
	Name string

	// GoPackage is the libvirtxml package path
	GoPackage string

	// Fields are the struct fields
	Fields []*FieldIR

	// Description from RNG or comments
	Description string

	// MarkdownDescription holds markdown-formatted docs for schema output
	MarkdownDescription string

	// IsTopLevel indicates this is a top-level resource (Domain, Network, etc.)
	IsTopLevel bool

	// ValueWithUnitPattern indicates this follows the value+unit pattern for flattening
	ValueWithUnitPattern *ValueWithUnitPattern

	// BooleanToPresence indicates this is an empty struct (presence-only)
	BooleanToPresence *BooleanToPresencePattern

	// IsExcluded indicates this struct should not be generated
	IsExcluded bool

	// ExclusionReason explains why this struct is excluded
	ExclusionReason string
}

// ValueWithUnitPattern represents a struct with chardata value + attributes.
// Pattern: Value INSIDE element, ALL other fields are attributes → FLATTEN
// Examples: <memory unit='KiB'>524288</memory> → memory, memory_unit
//
//	<vcpu placement='static' cpuset='1-4'>2</vcpu> → vcpu, vcpu_placement, vcpu_cpuset
type ValueWithUnitPattern struct {
	// ValueFieldName is the field with xml:",chardata" (e.g., "Value")
	ValueFieldName string

	// ValueGoType is the Go type of the value field (e.g., "uint", "uint64")
	ValueGoType string

	// AttributeFields are ALL attribute fields to be flattened
	AttributeFields []AttributeFieldInfo

	// FlattenStrategy is always "flatten" for this pattern
	FlattenStrategy string
}

// AttributeFieldInfo stores information about an attribute field to be flattened
type AttributeFieldInfo struct {
	GoName  string // Go field name (e.g., "Unit", "Placement")
	XMLName string // XML attribute name (e.g., "unit", "placement")
	GoType  string // Go type (e.g., "string", "uint")
}

// BooleanToPresencePattern indicates field is presence-only (empty struct).
type BooleanToPresencePattern struct {
	// IsPresenceOnly is true when the struct has no fields
	IsPresenceOnly bool
}

// StringToBoolPattern indicates string field should be boolean in TF.
type StringToBoolPattern struct {
	// TrueValue is the string representing true (e.g., "yes", "on")
	TrueValue string

	// FalseValue is the string representing false (e.g., "no", "off")
	FalseValue string
}

// FieldIR represents a single field in a struct.
type FieldIR struct {
	// GoName is the field name in Go (e.g., "Name")
	GoName string

	// XMLName is the XML element/attribute name (e.g., "name")
	XMLName string

	// TFName is the Terraform field name (snake_case, e.g., "tick_policy")
	TFName string

	// GoType is the Go type (e.g., "string", "uint64", "*DomainTimerCatchup")
	GoType string

	// TFType is the Terraform type (e.g., "types.String", "types.Int64", "types.Object")
	TFType string

	// IsXMLAttr indicates if this is an XML attribute vs element
	IsXMLAttr bool

	// IsPointer indicates if the Go type is a pointer
	IsPointer bool

	// IsOptional indicates if the field is optional in RNG schema
	IsOptional bool

	// IsRequired indicates if the field is required
	IsRequired bool

	// IsComputed indicates if this is a computed field
	IsComputed bool

	// IsNested indicates this is a nested struct
	IsNested bool

	// IsList indicates this should be a list in Terraform
	IsList bool

	// ElementGoType is the Go type of list elements (e.g., "string", "DomainDisk")
	ElementGoType string

	// NestedStruct is the IR for nested struct (if IsNested)
	NestedStruct *StructIR

	// Description from RNG or comments
	Description string

	// MarkdownDescription holds markdown-formatted docs for schema output
	MarkdownDescription string

	// ValidValues are the allowed values from RNG choice elements
	ValidValues []string

	// OmitEmpty indicates xml:omitempty tag
	OmitEmpty bool

	// IsCycle indicates this field creates a circular reference and should be skipped
	IsCycle bool

	// CycleNote explains which cycle was detected (e.g., "DomainDiskSource → DomainDiskDataStore → DomainDiskSource")
	CycleNote string

	// PreserveUserIntent means this optional field should only be set if user specified it
	PreserveUserIntent bool

	// StringToBool indicates this string field should be boolean in TF
	StringToBool *StringToBoolPattern

	// IsPresenceBoolean indicates this boolean field represents presence/absence of empty struct
	IsPresenceBoolean bool

	// IsFlattenedValue indicates this is the value part of a flattened chardata+attrs pattern
	IsFlattenedValue bool

	// IsFlattenedUnit indicates this is the unit part of a flattened value+unit pair (DEPRECATED: use IsFlattenedAttr)
	IsFlattenedUnit bool

	// IsFlattenedAttr indicates this is an attribute part of a flattened chardata+attrs pattern
	IsFlattenedAttr bool

	// FlattenedAttrName stores the XML attribute this flattened field represents (e.g., "unit", "placement")
	FlattenedAttrName string

	// TFPath stores the full Terraform path (e.g., "domain.cpu.model")
	TFPath string

	// XMLPath stores the full libvirt XML path (e.g., "domain.cpu.model")
	XMLPath string

	// IsExcluded indicates this field should not be generated
	IsExcluded bool

	// ExclusionReason explains why this field is excluded
	ExclusionReason string

	// PlanModifier specifies which plan modifier to use (e.g., "UseStateForUnknown", "RequiresReplace")
	PlanModifier string
}

// SchemaType returns the Terraform schema attribute type.
func (f *FieldIR) SchemaType() string {
	if f.IsList {
		if f.IsNested {
			return "schema.ListNestedAttribute"
		}
		// Simple list types
		switch f.TFType {
		case "types.String":
			return "schema.ListAttribute"
		default:
			return "schema.ListAttribute"
		}
	}

	if f.IsNested {
		return "schema.SingleNestedAttribute"
	}

	switch f.TFType {
	case "types.String":
		return "schema.StringAttribute"
	case "types.Int64":
		return "schema.Int64Attribute"
	case "types.Bool":
		return "schema.BoolAttribute"
	default:
		return "schema.StringAttribute"
	}
}

// ElementType returns the element type for list attributes.
func (f *FieldIR) ElementType() string {
	if !f.IsList {
		return ""
	}

	// For nested structs in lists, we need to determine the type from the struct
	if f.IsNested && f.NestedStruct != nil {
		// Will be handled with AttributeTypes() in schema generation
		return ""
	}

	// For simple types, infer from ElementGoType
	// ElementGoType will be like "string", "int", "uint", etc.
	switch f.ElementGoType {
	case "string":
		return "types.StringType"
	case "int", "int8", "int16", "int32", "int64":
		return "types.Int64Type"
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return "types.Int64Type"
	case "bool":
		return "types.BoolType"
	case "float32", "float64":
		return "types.Float64Type"
	default:
		return "types.StringType"
	}
}

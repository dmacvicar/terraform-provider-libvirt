package policy

import "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"

// ApplyFieldPolicies mutates the IR with Terraform-specific schema/conversion
// semantics after structural reflection is complete.
func ApplyFieldPolicies(structs []*generator.StructIR) {
	for _, s := range structs {
		applyStructPolicies(s)
	}
}

func applyStructPolicies(s *generator.StructIR) {
	for _, field := range s.Fields {
		if field.IsExcluded || field.IsCycle {
			continue
		}

		applyTopLevelIdentityPolicy(s, field)
		applyTopLevelImmutabilityPolicy(s, field)
		applyReportedFieldPolicy(s, field)
		applyReadbackPreservationPolicy(s, field)
	}
}

func applyTopLevelIdentityPolicy(s *generator.StructIR, field *generator.FieldIR) {
	if !s.IsTopLevel {
		return
	}

	switch field.TFName {
	case "uuid", "id", "key":
		field.IsComputed = true
		field.IsOptional = false
		field.IsRequired = false
		field.PlanModifier = "UseStateForUnknown"
	}
}

func applyTopLevelImmutabilityPolicy(s *generator.StructIR, field *generator.FieldIR) {
	if !s.IsTopLevel {
		return
	}

	switch field.TFName {
	case "name":
		field.IsRequired = true
		field.IsOptional = false
		field.IsComputed = false
		field.PlanModifier = "RequiresReplace"
	case "type":
		if s.Name != "StorageVolume" {
			field.IsRequired = true
			field.IsOptional = false
			field.IsComputed = false
			field.PlanModifier = "RequiresReplace"
		}
	}
}

func applyReportedFieldPolicy(s *generator.StructIR, field *generator.FieldIR) {
	if field.IsFlattenedUnit {
		return
	}

	switch s.Name {
	case "StoragePool":
		switch field.TFName {
		case "capacity":
			field.IsComputed = true
			field.IsOptional = false
			field.IsRequired = false
			field.PlanModifier = "UseStateForUnknown"
			field.PreserveUserIntent = false
		case "allocation", "available":
			field.IsComputed = true
			field.IsOptional = false
			field.IsRequired = false
			field.PreserveUserIntent = false
		}
	case "StorageVolume":
		switch field.TFName {
		case "capacity":
			field.IsComputed = true
			field.IsOptional = false
			field.IsRequired = false
		case "allocation", "physical":
			field.IsComputed = true
			field.IsOptional = false
			field.IsRequired = false
			field.PreserveUserIntent = false
		}
	}
}

func applyReadbackPreservationPolicy(s *generator.StructIR, field *generator.FieldIR) {
	switch s.Name {
	case "DomainCPU":
		if field.TFName == "mode" {
			field.PreservePlannedValueOnReadbackOmit = true
		}
	case "DomainGraphicSpice":
		if field.TFName == "listen" {
			field.PreservePlannedValueOnReadbackOmit = true
		}
	}
}

package policy

import "github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"

type fieldPolicy func(*generator.FieldIR)

var fieldPolicies = map[string][]fieldPolicy{
	"StoragePool.capacity": {
		policyComputedReportedField,
		policyUseStateForUnknown,
		policyDisablePreserveUserIntent,
	},
	"StoragePool.allocation": {
		policyComputedReportedField,
		policyDisablePreserveUserIntent,
	},
	"StoragePool.available": {
		policyComputedReportedField,
		policyDisablePreserveUserIntent,
	},
	"StorageVolume.capacity": {
		policyComputedReportedField,
	},
	"StorageVolume.allocation": {
		policyComputedReportedField,
		policyDisablePreserveUserIntent,
	},
	"StorageVolume.physical": {
		policyComputedReportedField,
		policyDisablePreserveUserIntent,
	},
	"DomainCPU.mode": {
		policyPreservePlannedValueOnReadbackOmit,
	},
	"DomainGraphicSpice.listen": {
		policyPreservePlannedValueOnReadbackOmit,
	},
}

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
		applyExplicitFieldPolicies(s, field)
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

func applyExplicitFieldPolicies(s *generator.StructIR, field *generator.FieldIR) {
	if field.IsFlattenedUnit {
		return
	}

	key := fieldOverrideKey(s, field)
	applyPolicies(field, fieldPolicies[key])
}

func fieldOverrideKey(s *generator.StructIR, field *generator.FieldIR) string {
	return s.Name + "." + field.TFName
}

func applyPolicies(field *generator.FieldIR, policies []fieldPolicy) {
	for _, policy := range policies {
		policy(field)
	}
}

func policyComputedReportedField(field *generator.FieldIR) {
	field.IsComputed = true
	field.IsOptional = false
	field.IsRequired = false
}

func policyUseStateForUnknown(field *generator.FieldIR) {
	field.PlanModifier = "UseStateForUnknown"
}

func policyDisablePreserveUserIntent(field *generator.FieldIR) {
	field.PreserveUserIntent = false
}

func policyPreservePlannedValueOnReadbackOmit(field *generator.FieldIR) {
	field.PreservePlannedValueOnReadbackOmit = true
}

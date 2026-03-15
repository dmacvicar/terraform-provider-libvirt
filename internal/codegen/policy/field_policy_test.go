package policy

import (
	"testing"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/codegen/generator"
)

func TestApplyFieldPoliciesMarksTopLevelIdentityComputed(t *testing.T) {
	root := &generator.StructIR{
		Name:       "Network",
		IsTopLevel: true,
		Fields: []*generator.FieldIR{
			{TFName: "id", IsRequired: true},
		},
	}

	ApplyFieldPolicies([]*generator.StructIR{root})

	field := root.Fields[0]
	if !field.IsComputed {
		t.Fatal("expected top-level id to be computed")
	}
	if field.IsRequired {
		t.Fatal("expected top-level id to not be required")
	}
	if field.PlanModifier != "UseStateForUnknown" {
		t.Fatalf("expected top-level id plan modifier UseStateForUnknown, got %q", field.PlanModifier)
	}
}

func TestApplyFieldPoliciesLeavesNestedIdentityUserManaged(t *testing.T) {
	nested := &generator.StructIR{
		Name: "DomainAudio",
		Fields: []*generator.FieldIR{
			{TFName: "id", IsRequired: true},
		},
	}

	ApplyFieldPolicies([]*generator.StructIR{nested})

	field := nested.Fields[0]
	if field.IsComputed {
		t.Fatal("expected nested id to remain user-managed")
	}
	if !field.IsRequired {
		t.Fatal("expected nested id to remain required")
	}
	if field.PlanModifier != "" {
		t.Fatalf("expected nested id to have no plan modifier, got %q", field.PlanModifier)
	}
}

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

func TestApplyFieldPoliciesMarksReadbackPreservationOverrides(t *testing.T) {
	structs := []*generator.StructIR{
		{
			Name: "DomainCPU",
			Fields: []*generator.FieldIR{
				{TFName: "mode"},
			},
		},
		{
			Name: "DomainGraphicSpice",
			Fields: []*generator.FieldIR{
				{TFName: "listen"},
			},
		},
	}

	ApplyFieldPolicies(structs)

	if !structs[0].Fields[0].PreservePlannedValueOnReadbackOmit {
		t.Fatal("expected DomainCPU.mode to preserve planned value on omitted readback")
	}
	if !structs[1].Fields[0].PreservePlannedValueOnReadbackOmit {
		t.Fatal("expected DomainGraphicSpice.listen to preserve planned value on omitted readback")
	}
}

func TestApplyFieldPoliciesMarksReportedFieldOverrides(t *testing.T) {
	structs := []*generator.StructIR{
		{
			Name: "StoragePool",
			Fields: []*generator.FieldIR{
				{TFName: "capacity", IsOptional: true, IsRequired: true, PreserveUserIntent: true},
				{TFName: "allocation", IsOptional: true, IsRequired: true, PreserveUserIntent: true},
			},
		},
		{
			Name: "StorageVolume",
			Fields: []*generator.FieldIR{
				{TFName: "capacity", IsOptional: true, IsRequired: true},
				{TFName: "physical", IsOptional: true, IsRequired: true, PreserveUserIntent: true},
			},
		},
	}

	ApplyFieldPolicies(structs)

	poolCapacity := structs[0].Fields[0]
	if !poolCapacity.IsComputed || poolCapacity.IsOptional || poolCapacity.IsRequired {
		t.Fatal("expected StoragePool.capacity to be computed-only after override")
	}
	if poolCapacity.PlanModifier != "UseStateForUnknown" {
		t.Fatalf("expected StoragePool.capacity plan modifier UseStateForUnknown, got %q", poolCapacity.PlanModifier)
	}
	if poolCapacity.PreserveUserIntent {
		t.Fatal("expected StoragePool.capacity to disable PreserveUserIntent")
	}

	poolAllocation := structs[0].Fields[1]
	if !poolAllocation.IsComputed || poolAllocation.IsOptional || poolAllocation.IsRequired {
		t.Fatal("expected StoragePool.allocation to be computed-only after override")
	}
	if poolAllocation.PreserveUserIntent {
		t.Fatal("expected StoragePool.allocation to disable PreserveUserIntent")
	}

	volumeCapacity := structs[1].Fields[0]
	if !volumeCapacity.IsComputed || volumeCapacity.IsOptional || volumeCapacity.IsRequired {
		t.Fatal("expected StorageVolume.capacity to be computed-only after override")
	}

	volumePhysical := structs[1].Fields[1]
	if !volumePhysical.IsComputed || volumePhysical.IsOptional || volumePhysical.IsRequired {
		t.Fatal("expected StorageVolume.physical to be computed-only after override")
	}
	if volumePhysical.PreserveUserIntent {
		t.Fatal("expected StorageVolume.physical to disable PreserveUserIntent")
	}
}

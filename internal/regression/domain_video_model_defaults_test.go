package regression

import (
	"context"
	"testing"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"libvirt.org/go/libvirtxml"
)

// #1365
func TestDomainVideoModelPreservesOmittedLibvirtDefaults(t *testing.T) {
	ctx := context.Background()

	plannedModel := generated.DomainVideoModelModel{
		Type:       types.StringValue("virtio"),
		Heads:      types.Int64Null(),
		Ram:        types.Int64Null(),
		VRam:       types.Int64Null(),
		VRam64:     types.Int64Null(),
		VGAMem:     types.Int64Null(),
		Primary:    types.StringNull(),
		Blob:       types.StringNull(),
		EDID:       types.StringNull(),
		Accel:      types.ObjectNull(generated.DomainVideoAccelAttributeTypes()),
		Resolution: types.ObjectNull(generated.DomainVideoResolutionAttributeTypes()),
	}
	plannedModelValue, diags := types.ObjectValueFrom(ctx, generated.DomainVideoModelAttributeTypes(), plannedModel)
	if diags.HasError() {
		t.Fatalf("creating planned video model: %v", diags)
	}

	plan := &generated.DomainVideoModel{Model: plannedModelValue}
	xml := &libvirtxml.DomainVideo{
		Model: libvirtxml.DomainVideoModel{
			Type:    "virtio",
			Heads:   1,
			Primary: "yes",
		},
	}

	model, err := generated.DomainVideoFromXML(ctx, xml, plan)
	if err != nil {
		t.Fatalf("converting domain video from XML: %v", err)
	}

	var got generated.DomainVideoModelModel
	if diags := model.Model.As(ctx, &got, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("reading converted video model: %v", diags)
	}

	if !got.Heads.IsNull() {
		t.Errorf("expected omitted heads to remain null, got %s", got.Heads.String())
	}
	if !got.Primary.IsNull() {
		t.Errorf("expected omitted primary to remain null, got %s", got.Primary.String())
	}
}

// #1365
func TestDomainVideoPreservesOmittedModel(t *testing.T) {
	ctx := context.Background()
	plan := &generated.DomainVideoModel{
		Model: types.ObjectNull(generated.DomainVideoModelAttributeTypes()),
	}
	xml := &libvirtxml.DomainVideo{
		Model: libvirtxml.DomainVideoModel{
			Type:    "virtio",
			Heads:   1,
			Primary: "yes",
		},
	}

	model, err := generated.DomainVideoFromXML(ctx, xml, plan)
	if err != nil {
		t.Fatalf("converting domain video from XML: %v", err)
	}

	if !model.Model.IsNull() {
		t.Fatalf("expected omitted model to remain null, got %s", model.Model.String())
	}
}

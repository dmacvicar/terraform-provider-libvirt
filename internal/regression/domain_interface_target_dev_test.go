package regression

import (
	"context"
	"testing"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"libvirt.org/go/libvirtxml"
)

func TestDomainInterfaceTargetDevPreservesPlannedHostGeneratedName(t *testing.T) {
	ctx := context.Background()

	plan := &generated.DomainInterfaceTargetModel{
		Dev: types.StringValue("vnet0"),
	}
	xml := &libvirtxml.DomainInterfaceTarget{
		Dev: "vnet16",
	}

	model, err := generated.DomainInterfaceTargetFromXML(ctx, xml, plan)
	if err != nil {
		t.Fatalf("converting domain interface target from XML: %v", err)
	}

	if got, want := model.Dev.ValueString(), "vnet0"; got != want {
		t.Fatalf("expected planned target dev to be preserved for host-generated name, got %q want %q", got, want)
	}
}

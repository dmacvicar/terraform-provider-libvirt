package regression

import (
	"context"
	"testing"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"libvirt.org/go/libvirtxml"
)

func TestDomainChardevSourceUNIXPathPreservesPlannedValueAgainstRuntimeRewrite(t *testing.T) {
	ctx := context.Background()

	plan := &generated.DomainChardevSourceUNIXModel{
		Path: types.StringValue("/var/lib/libvirt/qemu/channel/target/test.org.qemu.guest_agent.0"),
	}
	xml := &libvirtxml.DomainChardevSourceUNIX{
		Path: "/run/libvirt/qemu/channel/103-test/org.qemu.guest_agent.0",
	}

	model, err := generated.DomainChardevSourceUNIXFromXML(ctx, xml, plan)
	if err != nil {
		t.Fatalf("converting chardev source unix from XML: %v", err)
	}

	want := "/var/lib/libvirt/qemu/channel/target/test.org.qemu.guest_agent.0"
	if got := model.Path.ValueString(); got != want {
		t.Fatalf("expected planned path to be preserved against libvirt runtime rewrite, got %q want %q", got, want)
	}
}

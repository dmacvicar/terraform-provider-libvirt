package regression

import (
	"testing"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestNetworkVLANTagIDIsRequired(t *testing.T) {
	attr := generated.NetworkVLANTagSchemaAttribute()
	nested, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected SingleNestedAttribute, got %T", attr)
	}

	idAttr, ok := nested.Attributes["id"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("expected Int64Attribute for network VLAN tag id, got %T", nested.Attributes["id"])
	}

	if !idAttr.Required {
		t.Fatal("expected network VLAN tag id to be required")
	}
	if idAttr.Computed {
		t.Fatal("expected network VLAN tag id to not be computed")
	}
}

func TestDomainInterfaceVLanTagIDIsRequired(t *testing.T) {
	attr := generated.DomainInterfaceVLanTagSchemaAttribute()
	nested, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected SingleNestedAttribute, got %T", attr)
	}

	idAttr, ok := nested.Attributes["id"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("expected Int64Attribute for domain interface VLAN tag id, got %T", nested.Attributes["id"])
	}

	if !idAttr.Required {
		t.Fatal("expected domain interface VLAN tag id to be required")
	}
	if idAttr.Computed {
		t.Fatal("expected domain interface VLAN tag id to not be computed")
	}
}

func TestDomainAudioIDIsRequired(t *testing.T) {
	attr := generated.DomainAudioSchemaAttribute()
	nested, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected SingleNestedAttribute, got %T", attr)
	}

	idAttr, ok := nested.Attributes["id"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("expected Int64Attribute for domain audio id, got %T", nested.Attributes["id"])
	}

	if !idAttr.Required {
		t.Fatal("expected domain audio id to be required")
	}
	if idAttr.Computed {
		t.Fatal("expected domain audio id to not be computed")
	}
}

func TestDomainCPUMemoryTuneNodeIDIsRequired(t *testing.T) {
	attr := generated.DomainCPUMemoryTuneNodeSchemaAttribute()
	nested, ok := attr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("expected SingleNestedAttribute, got %T", attr)
	}

	idAttr, ok := nested.Attributes["id"].(schema.Int64Attribute)
	if !ok {
		t.Fatalf("expected Int64Attribute for domain CPU memory tune node id, got %T", nested.Attributes["id"])
	}

	if !idAttr.Required {
		t.Fatal("expected domain CPU memory tune node id to be required")
	}
	if idAttr.Computed {
		t.Fatal("expected domain CPU memory tune node id to not be computed")
	}
}

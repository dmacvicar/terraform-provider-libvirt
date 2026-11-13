package provider

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/generated"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// The volume resource overrides `backing_store` to add RequiresReplace. Guard
// that the override keeps every field, and every nested attribute, of the
// generated definition so the plan modifier is the only difference.
func TestVolumeResourceBackingStoreMatchesGeneratedSchema(t *testing.T) {
	r := &VolumeResource{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, resp)

	override, ok := resp.Schema.Attributes["backing_store"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("backing_store is %T, want schema.SingleNestedAttribute", resp.Schema.Attributes["backing_store"])
	}
	generated := mustSingleNestedAttribute(
		generated.StorageVolumeBackingStoreSchemaAttribute(),
		"StorageVolumeBackingStore",
	)

	if override.Optional != generated.Optional {
		t.Errorf("Optional = %v, want %v", override.Optional, generated.Optional)
	}
	if override.Required != generated.Required {
		t.Errorf("Required = %v, want %v", override.Required, generated.Required)
	}
	if override.Description != generated.Description {
		t.Errorf("Description changed:\n got: %q\nwant: %q", override.Description, generated.Description)
	}
	if override.MarkdownDescription != generated.MarkdownDescription {
		t.Errorf("MarkdownDescription changed:\n got: %q\nwant: %q", override.MarkdownDescription, generated.MarkdownDescription)
	}

	if got, want := sortedKeys(override.Attributes), sortedKeys(generated.Attributes); !reflect.DeepEqual(got, want) {
		t.Fatalf("nested attributes = %v, want %v", got, want)
	}
	for name, want := range generated.Attributes {
		got := override.Attributes[name]
		if reflect.TypeOf(got) != reflect.TypeOf(want) {
			t.Errorf("attribute %q is %T, want %T", name, got, want)
		}
	}

	// The one intended difference.
	if len(override.PlanModifiers) != len(generated.PlanModifiers)+1 {
		t.Errorf("PlanModifiers = %d, want generated %d + 1",
			len(override.PlanModifiers), len(generated.PlanModifiers))
	}
}

func sortedKeys(m map[string]schema.Attribute) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

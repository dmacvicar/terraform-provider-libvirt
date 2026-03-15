package parser

import (
	"reflect"
	"slices"
	"testing"

	"libvirt.org/go/libvirtxml"
)

func TestReflectStruct_AnonymousEmbeddedFields(t *testing.T) {
	reflector := NewLibvirtXMLReflector()

	ir, err := reflector.ReflectStruct(reflect.TypeOf(libvirtxml.DomainFeatureHyperVSpinlocks{}))
	if err != nil {
		t.Fatalf("ReflectStruct() error = %v", err)
	}

	fields := make(map[string]bool, len(ir.Fields))
	for _, field := range ir.Fields {
		fields[field.TFName] = true
	}

	if !fields["state"] {
		t.Fatalf("expected embedded field %q to be reflected, got fields %v", "state", keys(fields))
	}

	if !fields["retries"] {
		t.Fatalf("expected field %q to be reflected, got fields %v", "retries", keys(fields))
	}
}

func keys(fields map[string]bool) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

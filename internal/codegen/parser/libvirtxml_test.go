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

func TestReflectStruct_HypervisorNamespaceFields(t *testing.T) {
	reflector := NewLibvirtXMLReflector()

	ir, err := reflector.ReflectStruct(reflect.TypeOf(libvirtxml.Domain{}))
	if err != nil {
		t.Fatalf("ReflectStruct() error = %v", err)
	}

	fields := make(map[string]*generatorFieldView, len(ir.Fields))
	for _, field := range ir.Fields {
		fields[field.TFName] = &generatorFieldView{
			GoName:  field.GoName,
			XMLName: field.XMLName,
		}
	}

	testCases := map[string]generatorFieldView{
		"qemu_commandline":        {GoName: "QEMUCommandline", XMLName: "commandline"},
		"qemu_capabilities":       {GoName: "QEMUCapabilities", XMLName: "capabilities"},
		"qemu_override":           {GoName: "QEMUOverride", XMLName: "override"},
		"qemu_deprecation":        {GoName: "QEMUDeprecation", XMLName: "deprecation"},
		"lxc_namespace":           {GoName: "LXCNamespace", XMLName: "namespace"},
		"bhyve_commandline":       {GoName: "BHyveCommandline", XMLName: "commandline"},
		"vmware_data_center_path": {GoName: "VMWareDataCenterPath", XMLName: "datacenterpath"},
		"xen_commandline":         {GoName: "XenCommandline", XMLName: "commandline"},
	}

	for tfName, want := range testCases {
		got, ok := fields[tfName]
		if !ok {
			t.Fatalf("expected field %q to be reflected, got fields %v", tfName, keysFromViews(fields))
		}
		if got.GoName != want.GoName {
			t.Fatalf("field %q GoName = %q, want %q", tfName, got.GoName, want.GoName)
		}
		if got.XMLName != want.XMLName {
			t.Fatalf("field %q XMLName = %q, want %q", tfName, got.XMLName, want.XMLName)
		}
	}
}

type generatorFieldView struct {
	GoName  string
	XMLName string
}

func keys(fields map[string]bool) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func keysFromViews(fields map[string]*generatorFieldView) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

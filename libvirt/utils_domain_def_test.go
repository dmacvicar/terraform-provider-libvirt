package libvirt

import (
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func init() {
	spew.Config.Indent = "\t"
}

func TestSplitKernelCmdLine(t *testing.T) {
	e := []map[string]string{
		{"foo": "bar"},
		{
			"foo":  "bar",
			"key":  "val",
			"root": "UUID=aa52d618-a2c4-4aad-aeb7-68d9e3a2c91d"},
		{"_": "nosplash rw"}}
	r, err := splitKernelCmdLine("foo=bar foo=bar key=val root=UUID=aa52d618-a2c4-4aad-aeb7-68d9e3a2c91d nosplash rw")
	if !reflect.DeepEqual(r, e) {
		t.Fatalf("got='%s' expected='%s'", spew.Sdump(r), spew.Sdump(e))
	}
	if err != nil {
		t.Error(err)
	}
}

func TestSplitKernelEmptyCmdLine(t *testing.T) {
	var e []map[string]string
	r, err := splitKernelCmdLine("")
	if !reflect.DeepEqual(r, e) {
		t.Fatalf("got='%s' expected='%s'", spew.Sdump(r), spew.Sdump(e))
	}
	if err != nil {
		t.Error(err)
	}
}

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
	e := []map[string]string{{"foo": "bar"}, {"foo": "bar", "key": "val"}}
	r, err := splitKernelCmdLine("foo=bar foo=bar key=val")
	if !reflect.DeepEqual(r, e) {
		t.Fatalf("got='%s' expected='%s'", spew.Sdump(r), spew.Sdump(e))
	}
	if err != nil {
		t.Error(err)
	}
}

func TestSplitKernelInvalidCmdLine(t *testing.T) {
	v := "foo=barfoo=bar"
	r, err := splitKernelCmdLine(v)
	if r != nil {
		t.Fatalf("got='%s' expected='%s'", spew.Sdump(r), err)
	}
	if err == nil {
		t.Errorf("Expected error for parsing '%s'", v)
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

package main

import (
	"bytes"
	"regexp"
	"testing"
)

func TestPrintVersion(t *testing.T) {
	buf := &bytes.Buffer{}
	err := printVersion(buf)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.Bytes()

	re := regexp.MustCompile(`^.*terraform-provider-libvirt.test was not built correctly\nCompiled against library: libvirt [0-9.]*\nUsing library: libvirt [0-9.]*\nRunning hypervisor: .* [0-9.]*\nRunning against daemon: [0-9.]*\n$`)
	if !re.Match(output) {
		t.Fatalf("unexpected output:\n%q", string(output))
	}
}

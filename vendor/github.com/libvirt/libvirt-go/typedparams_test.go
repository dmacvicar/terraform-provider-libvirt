/*
 * This file is part of the libvirt-go project
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * Copyright (c) 2013 Alex Zorin
 * Copyright (C) 2016 Red Hat, Inc.
 *
 */

package libvirt

import (
	"testing"
)

func TestPackUnpack(t *testing.T) {
	in1 := "input1"
	in1s := true
	expect1 := "input1"
	got1 := ""
	got1s := false

	in2 := ""
	in2s := false
	expect2 := "not-changed"
	got2 := expect2
	got2s := false

	in3 := []string{"input3a", "input3b", "input3c", "input3d"}
	in3s := true
	expect3 := []string{"input3a", "input3b", "input3c", "input3d"}
	got3 := []string{}
	got3s := false

	infoin := make(map[string]typedParamsFieldInfo)
	infoin["data1"] = typedParamsFieldInfo{
		set: &in1s,
		s:   &in1,
	}
	infoin["data2"] = typedParamsFieldInfo{
		set: &in2s,
		s:   &in2,
	}
	infoin["data3"] = typedParamsFieldInfo{
		set: &in3s,
		sl:  &in3,
	}

	infoout := make(map[string]typedParamsFieldInfo)
	infoout["data1"] = typedParamsFieldInfo{
		set: &got1s,
		s:   &got1,
	}
	infoout["data2"] = typedParamsFieldInfo{
		set: &got2s,
		s:   &got2,
	}
	infoout["data3"] = typedParamsFieldInfo{
		set: &got3s,
		sl:  &got3,
	}

	params, err := typedParamsPackNew(infoin)
	if err != nil {
		t.Fatal(err)
	}

	nout, err := typedParamsUnpack(*params, infoout)
	if err != nil {
		t.Fatal(err)
	}

	if nout != 5 {
		t.Fatalf("Expected 5 output parameters, not %d", nout)
	}

	if got1 != expect1 {
		t.Fatalf("Expected '%s' but got '%s'", expect1, got1)
	}
	if got2 != expect2 {
		t.Fatalf("Expected '%s' but got '%s'", expect2, got2)
	}
	if len(got3) != len(expect3) {
		t.Fatalf("Expected '%s' but got '%s'", expect3, got3)
	}
	for i := 0; i < len(got3); i++ {
		if len(got3[i]) != len(expect3[i]) {
			t.Fatalf("Expected '%s' but got '%s'", expect3[i], got3[i])
		}
	}
}

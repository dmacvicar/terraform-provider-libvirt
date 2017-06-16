// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"reflect"
	"testing"
)

// Minimally implement the Registrant interface
type registrant struct {
	name string
}

func (t registrant) Name() string {
	return t.name
}

func TestCreateRegister(t *testing.T) {
	type in struct {
		registrants []registrant
	}
	type out struct {
		registrants Registry
	}

	a := registrant{name: "a"}
	b := registrant{name: "b"}
	c := registrant{name: "c"}

	tests := []struct {
		name string
		in   in
		out  out
	}{
		{
			name: "empty",
			in:   in{registrants: []registrant{}},
			out:  out{registrants: Registry{name: "empty", registrants: map[string]Registrant{}}},
		},
		{
			name: "three abc ...",
			in:   in{registrants: []registrant{a, b, c}},
			out:  out{registrants: Registry{name: "three abc ...", registrants: map[string]Registrant{"a": a, "b": b, "c": c}}},
		},
	}

	for i, test := range tests {
		tr := Create(test.name)
		for _, r := range test.in.registrants {
			tr.Register(r)
		}
		if !reflect.DeepEqual(&test.out.registrants, tr) {
			t.Errorf("#%d: bad registrants: want %#v, got %#v", i, &test.out.registrants, tr)
		}
	}
}

func TestGet(t *testing.T) {
	type in struct {
		registrants []registrant
		name        string
	}
	type out struct {
		creator *registrant
	}

	a := registrant{name: "a"}
	b := registrant{name: "b"}
	c := registrant{name: "c"}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{registrants: nil, name: "a"},
			out: out{creator: nil},
		},
		{
			in:  in{registrants: []registrant{a, b, c}, name: "a"},
			out: out{creator: &a},
		},
		{
			in:  in{registrants: []registrant{a, b, c}, name: "c"},
			out: out{creator: &c},
		},
	}

	for i, test := range tests {
		tr := Create("test")
		for _, r := range test.in.registrants {
			tr.Register(r)
		}
		r := tr.Get(test.in.name)
		if r == nil {
			if test.out.creator != nil {
				t.Errorf("#%d: got nil expected %#v", i, r)
			}
		} else if !reflect.DeepEqual(*test.out.creator, r.(registrant)) {
			t.Errorf("#%d: bad registrant: want %#v, got %#v", i, *test.out.creator, r.(registrant))
		}
	}
}

func TestNames(t *testing.T) {
	type in struct {
		registrants []registrant
	}
	type out struct {
		names []string
	}

	a := registrant{name: "a"}
	b := registrant{name: "b"}
	c := registrant{name: "c"}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{registrants: nil},
			out: out{names: []string{}},
		},
		{
			in:  in{registrants: []registrant{a, b, c}},
			out: out{names: []string{"a", "b", "c"}},
		},
	}

	for i, test := range tests {
		tr := Create("test")
		for _, r := range test.in.registrants {
			tr.Register(r)
		}
		names := tr.Names()
		if !reflect.DeepEqual(test.out.names, names) {
			t.Errorf("#%d: bad names: want %#v, got %#v", i, test.out.names, names)
		}
	}
}

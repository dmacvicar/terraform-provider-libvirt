// Copyright 2016 CoreOS, Inc.
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

package types

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestSystemdUnitNameUnmarshalJSON(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		unit SystemdUnitName
		err  error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{data: `"test.service"`},
			out: out{unit: SystemdUnitName("test.service")},
		},
		{
			in:  in{data: `"test.socket"`},
			out: out{unit: SystemdUnitName("test.socket")},
		},
		{
			in:  in{data: `"test.blah"`},
			out: out{err: errors.New("invalid systemd unit extension")},
		},
	}

	for i, test := range tests {
		var unit SystemdUnitName
		err := json.Unmarshal([]byte(test.in.data), &unit)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(test.out.unit, unit) {
			t.Errorf("#%d: bad unit: want %#v, got %#v", i, test.out.unit, unit)
		}
	}
}

func TestNetworkdUnitNameUnmarshalJSON(t *testing.T) {
	type in struct {
		data string
	}
	type out struct {
		unit NetworkdUnitName
		err  error
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{data: `"test.network"`},
			out: out{unit: NetworkdUnitName("test.network")},
		},
		{
			in:  in{data: `"test.link"`},
			out: out{unit: NetworkdUnitName("test.link")},
		},
		{
			in:  in{data: `"test.netdev"`},
			out: out{unit: NetworkdUnitName("test.netdev")},
		},
		{
			in:  in{data: `"test.blah"`},
			out: out{err: errors.New("invalid networkd unit extension")},
		},
	}

	for i, test := range tests {
		var unit NetworkdUnitName
		err := json.Unmarshal([]byte(test.in.data), &unit)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("#%d: bad error: want %v, got %v", i, test.out.err, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(test.out.unit, unit) {
			t.Errorf("#%d: bad unit: want %#v, got %#v", i, test.out.unit, unit)
		}
	}
}

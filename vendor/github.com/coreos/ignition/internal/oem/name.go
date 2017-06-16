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

package oem

import (
	"fmt"
)

// Name is used to identify an OEM. It must be in the set of registered OEMs.
type Name string

func (s Name) String() string {
	return string(s)
}

func (s *Name) Set(val string) error {
	if _, ok := Get(val); !ok {
		return fmt.Errorf("%s is not a valid oem", val)
	}

	*s = Name(val)
	return nil
}

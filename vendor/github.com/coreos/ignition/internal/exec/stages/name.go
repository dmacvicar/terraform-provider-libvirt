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

package stages

import (
	"fmt"
)

// Name is used to identify a StageCreator (which instantiates the stage of
// the same name) from the command line. It must be in the set of registered
// stages.
type Name string

func (s Name) String() string {
	return string(s)
}

func (s *Name) Set(val string) error {
	if stage := Get(val); stage == nil {
		return fmt.Errorf("%s is not a valid stage", val)
	}

	*s = Name(val)
	return nil
}

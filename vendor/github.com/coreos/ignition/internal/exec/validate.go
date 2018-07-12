// Copyright 2017 CoreOS, Inc.
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

package exec

import (
	"io/ioutil"
	"os"

	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/config"
)

func Validate(filename string) (report.Report, error) {
	var err error
	var b []byte
	if filename == "-" {
		b, err = ioutil.ReadAll(os.Stdin)
	} else {
		b, err = ioutil.ReadFile(filename)
	}
	if err != nil {
		return report.Report{}, err
	}

	_, r, err := config.Parse(b)
	return r, err
}

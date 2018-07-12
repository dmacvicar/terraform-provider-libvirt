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

package general

import (
	"fmt"
	"strings"

	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	// Default config is applied
	register.Register(register.PositiveTest, makePreemptTest("D"))
	// Base and default configs are both applied
	register.Register(register.PositiveTest, makePreemptTest("BD"))
	// Base and provider configs are both applied
	register.Register(register.PositiveTest, makePreemptTest("BP"))
	// Provider config is applied; default config is ignored
	register.Register(register.PositiveTest, makePreemptTest("dP"))
	// Base and user configs are applied; provider and default
	// configs are ignored
	register.Register(register.PositiveTest, makePreemptTest("BdUp"))
	// No configs provided; Ignition should still run successfully
	register.Register(register.PositiveTest, makePreemptTest(""))
}

// makePreemptTest returns a config preemption test that executes some
// combination of "b"ase, "d"efault, "u"ser (user.ign), and "p"rovider
// (IGNITION_CONFIG_FILE) configs. Capital letters indicate configs that
// Ignition should apply.
func makePreemptTest(components string) types.Test {
	longnames := map[string]string{
		"b": "base",
		"d": "default",
		"u": "user",
		"p": "provider",
	}
	makeConfig := func(component string) string {
		return fmt.Sprintf(`{
			"ignition": {"version": "2.1.0"},
			"storage": {
				"files": [{
					"filesystem": "root",
					"path": "/ignition/%s",
					"contents": {"source": "data:,%s"}
				}]}
		}`, longnames[component], component)
	}
	enabled := func(component string) bool {
		return strings.Contains(strings.ToLower(components), component)
	}

	var longnameList []string
	for _, component := range strings.Split(strings.ToLower(components), "") {
		longnameList = append(longnameList, longnames[component])
	}
	if len(longnameList) == 0 {
		longnameList = append(longnameList, "no")
	}
	name := "Preemption with " + strings.Join(longnameList, ", ") + " config"

	in := types.GetBaseDisk()
	out := types.GetBaseDisk()

	var config string
	if enabled("p") {
		config = makeConfig("p")
	}

	var systemFiles []types.File
	for _, component := range []string{"b", "d", "u"} {
		if enabled(component) {
			systemFiles = append(systemFiles, types.File{
				Node: types.Node{
					Name: longnames[component] + ".ign",
				},
				Contents: makeConfig(component),
			})
		}
	}

	for component, longname := range longnames {
		in[0].Partitions.AddFiles("ROOT", []types.File{
			{
				Node: types.Node{
					Name:      longname,
					Directory: "ignition",
				},
				Contents: "unset",
			},
		})
		result := "unset"
		if strings.Contains(components, strings.ToUpper(component)) {
			result = component
		}
		out[0].Partitions.AddFiles("ROOT", []types.File{
			{
				Node: types.Node{
					Name:      longname,
					Directory: "ignition",
				},
				Contents: result,
			},
		})
	}

	return types.Test{
		Name:              name,
		In:                in,
		Out:               out,
		Config:            config,
		SystemDirFiles:    systemFiles,
		ConfigShouldBeBad: true,
	}
}

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

package files

import (
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	register.Register(register.PositiveTest, CreateDirectoryOnRoot())
	register.Register(register.PositiveTest, ForceDirCreation())
	register.Register(register.PositiveTest, ForceDirCreationOverNonemptyDir())
}

func CreateDirectoryOnRoot() types.Test {
	name := "Create a Directory on the Root Filesystem"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
	  "ignition": { "version": "2.1.0" },
	  "storage": {
	    "directories": [{
	      "filesystem": "root",
	      "path": "/foo/bar"
	    }]
	  }
	}`
	out[0].Partitions.AddDirectories("ROOT", []types.Directory{
		{
			Node: types.Node{
				Name:      "bar",
				Directory: "foo",
			},
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func ForceDirCreation() types.Test {
	name := "Force Directory Creation"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
	  "ignition": { "version": "2.2.0" },
	  "storage": {
	    "directories": [{
	      "filesystem": "root",
	      "path": "/foo/bar",
		  "overwrite": true
	    }]
	  }
	}`
	in[0].Partitions.AddFiles("ROOT", []types.File{
		{
			Node: types.Node{
				Directory: "foo",
				Name:      "bar",
			},
			Contents: "hello, world",
		},
	})
	out[0].Partitions.AddDirectories("ROOT", []types.Directory{
		{
			Node: types.Node{
				Directory: "foo",
				Name:      "bar",
			},
		},
	})

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func ForceDirCreationOverNonemptyDir() types.Test {
	name := "Force Directory Creation Over Nonempty Directory"
	in := types.GetBaseDisk()
	out := types.GetBaseDisk()
	config := `{
	  "ignition": { "version": "2.2.0" },
	  "storage": {
	    "directories": [{
	      "filesystem": "root",
	      "path": "/foo/bar",
		  "overwrite": true
	    }]
	  }
	}`
	in[0].Partitions.AddFiles("ROOT", []types.File{
		{
			Node: types.Node{
				Directory: "foo/bar",
				Name:      "baz",
			},
			Contents: "hello, world",
		},
	})
	out[0].Partitions.AddDirectories("ROOT", []types.Directory{
		{
			Node: types.Node{
				Directory: "foo",
				Name:      "bar",
			},
		},
	})
	// TODO: add ability to ensure that foo/bar/baz doesn't exist here.

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

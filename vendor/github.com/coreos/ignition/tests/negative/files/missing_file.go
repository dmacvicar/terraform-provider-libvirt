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
	register.Register(register.NegativeTest, MissingRemoteContentsHTTP())
	register.Register(register.NegativeTest, MissingRemoteContentsTFTP())
	register.Register(register.NegativeTest, MissingRemoteContentsOEM())
}

func MissingRemoteContentsHTTP() types.Test {
	name := "Missing File from Remote Contents - HTTP"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": { "version": "2.1.0" },
	  "storage": {
	    "files": [{
	      "filesystem": "root",
	      "path": "/foo/bar",
	      "contents": {
	        "source": "http://127.0.0.1:8080/asdf"
	      }
	    }]
	  }
	}`
	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func MissingRemoteContentsTFTP() types.Test {
	name := "Missing File from Remote Contents - TFTP"
	in := types.GetBaseDisk()
	out := in
	config := `{
          "ignition": { "version": "2.1.0" },
          "storage": {
            "files": [{
              "filesystem": "root",
              "path": "/foo/bar",
              "contents": {
                "source": "tftp://127.0.0.1:69/asdf"
              }
            }]
          }
        }`
	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func MissingRemoteContentsOEM() types.Test {
	name := "Create Files from Remote Contents - OEM"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": { "version": "2.1.0" },
	  "storage": {
	    "files": [{
	      "filesystem": "root",
	      "path": "/foo/bar",
	      "contents": {
	        "source": "oem:///source"
	      }
	    }]
	  }
	}`
	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

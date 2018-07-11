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
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	register.Register(register.NegativeTest, ReplaceConfigWithInvalidHash())
	register.Register(register.NegativeTest, AppendConfigWithInvalidHash())
	register.Register(register.NegativeTest, ReplaceConfigWithMissingFileHTTP())
	register.Register(register.NegativeTest, ReplaceConfigWithMissingFileTFTP())
	register.Register(register.NegativeTest, ReplaceConfigWithMissingFileOEM())
	register.Register(register.NegativeTest, AppendConfigWithMissingFileHTTP())
	register.Register(register.NegativeTest, AppendConfigWithMissingFileTFTP())
	register.Register(register.NegativeTest, AppendConfigWithMissingFileOEM())
	register.Register(register.NegativeTest, VersionOnlyConfig22())
	register.Register(register.NegativeTest, VersionOnlyConfig23())
	register.Register(register.NegativeTest, VersionOnlyConfig24())
}

func ReplaceConfigWithInvalidHash() types.Test {
	name := "Replace Config with Invalid Hash"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": {
	    "version": "2.0.0",
	    "config": {
	      "replace": {
	        "source": "http://127.0.0.1:8080/config",
			"verification": { "hash": "sha512-1a04c76c17079cd99e688ba4f1ba095b927d3fecf2b1e027af361dfeafb548f7f5f6fdd675aaa2563950db441d893ca77b0c3e965cdcb891784af96e330267d7" }
	      }
	    }
	  }
	}`

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func AppendConfigWithInvalidHash() types.Test {
	name := "Append Config with Invalid Hash"
	in := types.GetBaseDisk()
	out := in
	mntDevices := []types.MntDevice{
		{
			Label:        "EFI-SYSTEM",
			Substitution: "$DEVICE",
		},
	}
	config := `{
	  "ignition": {
	    "version": "2.0.0",
	    "config": {
	      "append": [{
	        "source": "http://127.0.0.1:8080/config",
			"verification": { "hash": "sha512-1a04c76c17079cd99e688ba4f1ba095b927d3fecf2b1e027af361dfeafb548f7f5f6fdd675aaa2563950db441d893ca77b0c3e965cdcb891784af96e330267d7" }
	      }]
	    }
	  },
      "storage": {
        "files": [{
          "filesystem": "root",
          "path": "/foo/bar2",
          "contents": { "source": "data:,another%20example%20file%0A" }
        }]
      }
	}`

	return types.Test{
		Name:       name,
		In:         in,
		Out:        out,
		MntDevices: mntDevices,
		Config:     config,
	}
}

func ReplaceConfigWithMissingFileHTTP() types.Test {
	name := "Replace Config with Missing File - HTTP"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.1.0",
	    "config": {
	      "replace": {
	        "source": "http://127.0.0.1:8080/asdf"
	      }
	    }
	  }
	}`

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func ReplaceConfigWithMissingFileTFTP() types.Test {
	name := "Replace Config with Missing File - TFTP"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.1.0",
	    "config": {
	      "replace": {
	        "source": "tftp://127.0.0.1:69/asdf"
	      }
	    }
	  }
	}`

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func ReplaceConfigWithMissingFileOEM() types.Test {
	name := "Replace Config with Missing File - OEM"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.1.0",
	    "config": {
	      "replace": {
	        "source": "oem:///asdf"
	      }
	    }
	  }
	}`

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func AppendConfigWithMissingFileHTTP() types.Test {
	name := "Append Config with Missing File - HTTP"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.1.0",
	    "config": {
	      "append": [{
	        "source": "http://127.0.0.1:8080/asdf"
	      }]
	    }
	  }
	}`

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func AppendConfigWithMissingFileTFTP() types.Test {
	name := "Append Config with Missing File - TFTP"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.1.0",
	    "config": {
	      "append": [{
	        "source": "tftp://127.0.0.1:69/asdf"
	      }]
	    }
	  }
	}`

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func AppendConfigWithMissingFileOEM() types.Test {
	name := "Append Config with Missing File - OEM"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.1.0",
	    "config": {
	      "append": [{
	        "source": "oem:///asdf"
	      }]
	    }
	  }
	}`

	return types.Test{
		Name:   name,
		In:     in,
		Out:    out,
		Config: config,
	}
}

func VersionOnlyConfig22() types.Test {
	name := "Version Only Config 2.2.0-experimental"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.2.0-experimental"
	  }
	}`

	return types.Test{
		Name:              name,
		In:                in,
		Out:               out,
		Config:            config,
		ConfigShouldBeBad: true,
	}
}

func VersionOnlyConfig23() types.Test {
	name := "Version Only Config 2.3.0"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.3.0"
	  }
	}`

	return types.Test{
		Name:              name,
		In:                in,
		Out:               out,
		Config:            config,
		ConfigShouldBeBad: true,
	}
}

func VersionOnlyConfig24() types.Test {
	name := "Version Only Config 2.4.0-experimental"
	in := types.GetBaseDisk()
	out := in
	config := `{
	  "ignition": {
	    "version": "2.4.0-experimental"
	  }
	}`

	return types.Test{
		Name:              name,
		In:                in,
		Out:               out,
		Config:            config,
		ConfigShouldBeBad: true,
	}
}

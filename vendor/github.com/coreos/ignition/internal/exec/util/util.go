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

package util

import (
	"path/filepath"

	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/resource"
)

// Util encapsulates logging and destdir indirection for the util methods.
type Util struct {
	DestDir string // directory prefix to use in applying fs paths.
	Fetcher resource.Fetcher
	*log.Logger
}

// JoinPath returns a path into the context ala filepath.Join(d, args)
func (u Util) JoinPath(path ...string) string {
	return filepath.Join(u.DestDir, filepath.Join(path...))
}

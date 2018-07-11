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
	"github.com/coreos/ignition/internal/config/types"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/registry"
	"github.com/coreos/ignition/internal/resource"
)

// Stage is responsible for actually executing a stage of the configuration.
type Stage interface {
	Run(config types.Config) bool
	Name() string
}

// StageCreator is responsible for instantiating a particular stage given a
// logger and root path under the root partition.
type StageCreator interface {
	Create(logger *log.Logger, root string, f resource.Fetcher) Stage
	Name() string
}

var stages = registry.Create("stages")

func Register(stage StageCreator) {
	stages.Register(stage)
}

func Get(name string) StageCreator {
	if s, ok := stages.Get(name).(StageCreator); ok {
		return s
	}
	return nil
}

func Names() (names []string) {
	return stages.Names()
}

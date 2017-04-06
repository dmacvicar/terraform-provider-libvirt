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

package registry

import (
	"fmt"
	"sort"
)

// Registrant interface implementors may be registered in a Registry
type Registrant interface {
	Name() string
}

type Registry struct {
	name        string
	registrants map[string]Registrant
}

// Create creates a new registry
func Create(name string) *Registry {
	return &Registry{name: name, registrants: map[string]Registrant{}}
}

// Register registers a new registrant to a registry
func (r *Registry) Register(registrant Registrant) {
	if _, ok := r.registrants[registrant.Name()]; ok {
		panic(fmt.Sprintf("%s: registrant %q already registered", r.name, registrant.Name()))
	}
	r.registrants[registrant.Name()] = registrant
}

// Get gets a named registrant from a registry
func (r *Registry) Get(name string) interface{} {
	return r.registrants[name]
}

// Names returns the sorted registrant names
func (r *Registry) Names() []string {
	keys := []string{}
	for key := range r.registrants {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

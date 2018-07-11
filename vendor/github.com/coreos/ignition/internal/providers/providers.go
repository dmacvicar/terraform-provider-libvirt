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

package providers

import (
	"errors"

	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/config/types"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/resource"
)

var (
	ErrNoProvider = errors.New("config provider was not online")
)

type FuncFetchConfig func(f resource.Fetcher) (types.Config, report.Report, error)
type FuncNewFetcher func(logger *log.Logger) (resource.Fetcher, error)

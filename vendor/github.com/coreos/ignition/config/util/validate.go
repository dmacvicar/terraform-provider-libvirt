// Copyright 2018 CoreOS, Inc.
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
	"bytes"
	"reflect"

	"github.com/coreos/ignition/config/validate"
	astjson "github.com/coreos/ignition/config/validate/astjson"
	"github.com/coreos/ignition/config/validate/report"

	json "github.com/ajeddeloh/go-json"
)

func ValidateConfig(rawConfig []byte, config interface{}) report.Report {
	// Unmarshal again to a json.Node to get offset information for building a report
	var ast json.Node
	var r report.Report
	configValue := reflect.ValueOf(config)
	if err := json.Unmarshal(rawConfig, &ast); err != nil {
		r.Add(report.Entry{
			Kind:    report.EntryWarning,
			Message: "Ignition could not unmarshal your config for reporting line numbers. This should never happen. Please file a bug.",
		})
		r.Merge(validate.ValidateWithoutSource(configValue))
	} else {
		r.Merge(validate.Validate(configValue, astjson.FromJsonRoot(ast), bytes.NewReader(rawConfig), true))
	}
	return r
}

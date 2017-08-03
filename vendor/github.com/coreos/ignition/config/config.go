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

package config

import (
	"bytes"
	"errors"
	"reflect"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/v1"
	"github.com/coreos/ignition/config/v2_0"
	"github.com/coreos/ignition/config/v2_1"
	"github.com/coreos/ignition/config/validate"
	astjson "github.com/coreos/ignition/config/validate/astjson"
	"github.com/coreos/ignition/config/validate/report"

	json "github.com/ajeddeloh/go-json"
	"github.com/coreos/go-semver/semver"
	"go4.org/errorutil"
)

var (
	ErrCloudConfig           = errors.New("not a config (found coreos-cloudconfig)")
	ErrEmpty                 = errors.New("not a config (empty)")
	ErrScript                = errors.New("not a config (found coreos-cloudinit script)")
	ErrDeprecated            = errors.New("config format deprecated")
	ErrInvalid               = errors.New("config is not valid")
	ErrUnknownVersion        = errors.New("unsupported config version")
	ErrVersionIndeterminable = errors.New("unable to determine version")
)

// Parse parses the raw config into a types.Config struct and generates a report of any
// errors, warnings, info, and deprecations it encountered
func Parse(rawConfig []byte) (types.Config, report.Report, error) {
	version, err := Version(rawConfig)
	if err != nil && err != ErrVersionIndeterminable {
		// If we can't determine the version, ignore this error so that in the
		// default case of the switch statement we can check for empty configs,
		// cloud configs, and other such things.
		return types.Config{}, report.ReportFromError(err, report.EntryError), err
	}
	switch version {
	case semver.Version{Major: 1}:
		config, err := ParseFromV1(rawConfig)
		if err != nil {
			return types.Config{}, report.ReportFromError(err, report.EntryError), err
		}

		return config, report.ReportFromError(ErrDeprecated, report.EntryDeprecated), nil
	case types.MaxVersion:
		return ParseFromLatest(rawConfig)
	case semver.Version{Major: 2, Minor: 1}:
		return ParseFromV2_1(rawConfig)
	case semver.Version{Major: 2, Minor: 0}:
		return ParseFromV2_0(rawConfig)
	default:
		if isEmpty(rawConfig) {
			return types.Config{}, report.Report{}, ErrEmpty
		} else if isCloudConfig(rawConfig) {
			return types.Config{}, report.Report{}, ErrCloudConfig
		} else if isScript(rawConfig) {
			return types.Config{}, report.Report{}, ErrScript
		}
		return types.Config{}, report.Report{}, ErrUnknownVersion
	}
}

func ParseFromLatest(rawConfig []byte) (types.Config, report.Report, error) {
	var err error
	var config types.Config

	// These errors are fatal and the config should not be further validated
	if err = json.Unmarshal(rawConfig, &config); err == nil {
		versionReport := config.Ignition.Validate()
		if versionReport.IsFatal() {
			return types.Config{}, versionReport, ErrInvalid
		}
	}

	// Handle json syntax and type errors first, since they are fatal but have offset info
	if serr, ok := err.(*json.SyntaxError); ok {
		line, col, highlight := errorutil.HighlightBytePosition(bytes.NewReader(rawConfig), serr.Offset)
		return types.Config{},
			report.Report{
				Entries: []report.Entry{{
					Kind:      report.EntryError,
					Message:   serr.Error(),
					Line:      line,
					Column:    col,
					Highlight: highlight,
				}},
			},
			ErrInvalid
	}

	if terr, ok := err.(*json.UnmarshalTypeError); ok {
		line, col, highlight := errorutil.HighlightBytePosition(bytes.NewReader(rawConfig), terr.Offset)
		return types.Config{},
			report.Report{
				Entries: []report.Entry{{
					Kind:      report.EntryError,
					Message:   terr.Error(),
					Line:      line,
					Column:    col,
					Highlight: highlight,
				}},
			},
			ErrInvalid
	}

	// Handle other fatal errors (i.e. invalid version)
	if err != nil {
		return types.Config{}, report.ReportFromError(err, report.EntryError), err
	}

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

	if r.IsFatal() {
		return types.Config{}, r, ErrInvalid
	}

	return config, r, nil
}

func ParseFromV1(rawConfig []byte) (types.Config, error) {
	config, err := v1.Parse(rawConfig)
	if err != nil {
		return types.Config{}, err
	}

	return TranslateFromV1(config), nil
}

func ParseFromV2_0(rawConfig []byte) (types.Config, report.Report, error) {
	cfg, report, err := v2_0.Parse(rawConfig)
	if err != nil {
		return types.Config{}, report, err
	}

	return TranslateFromV2_0(cfg), report, err
}

func ParseFromV2_1(rawConfig []byte) (types.Config, report.Report, error) {
	cfg, report, err := v2_1.Parse(rawConfig)
	if err != nil {
		return types.Config{}, report, err
	}

	return TranslateFromV2_1(cfg), report, err
}

func Version(rawConfig []byte) (semver.Version, error) {
	var composite struct {
		Version  *int `json:"ignitionVersion"`
		Ignition struct {
			Version *string `json:"version"`
		} `json:"ignition"`
	}

	if json.Unmarshal(rawConfig, &composite) == nil {
		if composite.Ignition.Version != nil {
			v, err := types.Ignition{Version: *composite.Ignition.Version}.Semver()
			if err != nil {
				return semver.Version{}, err
			}
			return *v, nil
		} else if composite.Version != nil {
			return semver.Version{Major: int64(*composite.Version)}, nil
		}
	}

	return semver.Version{}, ErrVersionIndeterminable
}

func isEmpty(userdata []byte) bool {
	return len(userdata) == 0
}

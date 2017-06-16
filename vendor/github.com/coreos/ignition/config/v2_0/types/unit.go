// Copyright 2016 CoreOS, Inc.
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

package types

import (
	"bytes"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/go-systemd/unit"

	"github.com/coreos/ignition/config/validate/report"
)

type SystemdUnit struct {
	Name     SystemdUnitName     `json:"name,omitempty"`
	Enable   bool                `json:"enable,omitempty"`
	Mask     bool                `json:"mask,omitempty"`
	Contents string              `json:"contents,omitempty"`
	DropIns  []SystemdUnitDropIn `json:"dropins,omitempty"`
}

func (u SystemdUnit) Validate() report.Report {
	if err := validateUnitContent(u.Contents); err != nil {
		return report.ReportFromError(err, report.EntryError)
	}

	return report.Report{}
}

type SystemdUnitDropIn struct {
	Name     SystemdUnitDropInName `json:"name,omitempty"`
	Contents string                `json:"contents,omitempty"`
}

func (u SystemdUnitDropIn) Validate() report.Report {
	if err := validateUnitContent(u.Contents); err != nil {
		return report.ReportFromError(err, report.EntryError)
	}

	return report.Report{}
}

type SystemdUnitName string

func (n SystemdUnitName) Validate() report.Report {
	switch path.Ext(string(n)) {
	case ".service", ".socket", ".device", ".mount", ".automount", ".swap", ".target", ".path", ".timer", ".snapshot", ".slice", ".scope":
		return report.Report{}
	default:
		return report.ReportFromError(errors.New("invalid systemd unit extension"), report.EntryError)
	}
}

type SystemdUnitDropInName string

func (n SystemdUnitDropInName) Validate() report.Report {
	switch path.Ext(string(n)) {
	case ".conf":
		return report.Report{}
	default:
		return report.ReportFromError(errors.New("invalid systemd unit drop-in extension"), report.EntryError)
	}
}

type NetworkdUnit struct {
	Name     NetworkdUnitName `json:"name,omitempty"`
	Contents string           `json:"contents,omitempty"`
}

func (u NetworkdUnit) Validate() report.Report {
	if err := validateUnitContent(u.Contents); err != nil {
		return report.ReportFromError(err, report.EntryError)
	}

	return report.Report{}
}

type NetworkdUnitName string

func (n NetworkdUnitName) Validate() report.Report {
	switch path.Ext(string(n)) {
	case ".link", ".netdev", ".network":
		return report.Report{}
	default:
		return report.ReportFromError(errors.New("invalid networkd unit extension"), report.EntryError)
	}
}

func validateUnitContent(content string) error {
	c := bytes.NewBufferString(content)
	_, err := unit.Deserialize(c)
	if err != nil {
		return fmt.Errorf("invalid unit content: %s", err)
	}

	return nil
}

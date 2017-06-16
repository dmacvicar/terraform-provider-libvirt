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
	"encoding/json"
	"errors"
	"path"
)

type SystemdUnit struct {
	Name     SystemdUnitName     `json:"name,omitempty"`
	Enable   bool                `json:"enable,omitempty"`
	Mask     bool                `json:"mask,omitempty"`
	Contents string              `json:"contents,omitempty"`
	DropIns  []SystemdUnitDropIn `json:"dropins,omitempty"`
}

type SystemdUnitDropIn struct {
	Name     SystemdUnitDropInName `json:"name,omitempty"`
	Contents string                `json:"contents,omitempty"`
}

type SystemdUnitName string
type systemdUnitName SystemdUnitName

func (n *SystemdUnitName) UnmarshalJSON(data []byte) error {
	tn := systemdUnitName(*n)
	if err := json.Unmarshal(data, &tn); err != nil {
		return err
	}
	*n = SystemdUnitName(tn)
	return n.AssertValid()
}

func (n SystemdUnitName) AssertValid() error {
	switch path.Ext(string(n)) {
	case ".service", ".socket", ".device", ".mount", ".automount", ".swap", ".target", ".path", ".timer", ".snapshot", ".slice", ".scope":
		return nil
	default:
		return errors.New("invalid systemd unit extension")
	}
}

type SystemdUnitDropInName string
type systemdUnitDropInName SystemdUnitDropInName

func (n *SystemdUnitDropInName) UnmarshalJSON(data []byte) error {
	tn := systemdUnitDropInName(*n)
	if err := json.Unmarshal(data, &tn); err != nil {
		return err
	}
	*n = SystemdUnitDropInName(tn)
	return n.AssertValid()
}

func (n SystemdUnitDropInName) AssertValid() error {
	switch path.Ext(string(n)) {
	case ".conf":
		return nil
	default:
		return errors.New("invalid systemd unit drop-in extension")
	}
}

type NetworkdUnit struct {
	Name     NetworkdUnitName `json:"name,omitempty"`
	Contents string           `json:"contents,omitempty"`
}

type NetworkdUnitName string
type networkdUnitName NetworkdUnitName

func (n *NetworkdUnitName) UnmarshalJSON(data []byte) error {
	tn := networkdUnitName(*n)
	if err := json.Unmarshal(data, &tn); err != nil {
		return err
	}
	*n = NetworkdUnitName(tn)
	return n.AssertValid()
}

func (n NetworkdUnitName) AssertValid() error {
	switch path.Ext(string(n)) {
	case ".link", ".netdev", ".network":
		return nil
	default:
		return errors.New("invalid networkd unit extension")
	}
}

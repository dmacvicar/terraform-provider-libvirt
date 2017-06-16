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

package systemd

import (
	"fmt"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
)

// WaitOnDevices waits for the devices named in devs to be plugged before returning.
func WaitOnDevices(devs []string, stage string) error {
	conn, err := dbus.NewSystemdConnection()
	if err != nil {
		return err
	}

	results := map[string]chan string{}
	for _, dev := range devs {
		unitName := unit.UnitNamePathEscape(dev + ".device")
		results[unitName] = make(chan string, 1)

		if _, err = conn.StartUnit(unitName, "replace", results[unitName]); err != nil {
			return fmt.Errorf("failed starting device unit %s: %v", unitName, err)
		}
	}

	for unitName, result := range results {
		s := <-result

		if s != "done" {
			return fmt.Errorf("device unit %s %s", unitName, s)
		}
	}

	return nil
}

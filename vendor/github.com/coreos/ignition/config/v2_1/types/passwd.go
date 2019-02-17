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

package types

import (
	"errors"

	"github.com/coreos/ignition/config/validate/report"
)

var (
	ErrPasswdCreateDeprecated      = errors.New("the create object has been deprecated in favor of user-level options")
	ErrPasswdCreateAndGecos        = errors.New("cannot use both the create object and the user-level gecos field")
	ErrPasswdCreateAndGroups       = errors.New("cannot use both the create object and the user-level groups field")
	ErrPasswdCreateAndHomeDir      = errors.New("cannot use both the create object and the user-level homeDir field")
	ErrPasswdCreateAndNoCreateHome = errors.New("cannot use both the create object and the user-level noCreateHome field")
	ErrPasswdCreateAndNoLogInit    = errors.New("cannot use both the create object and the user-level noLogInit field")
	ErrPasswdCreateAndNoUserGroup  = errors.New("cannot use both the create object and the user-level noUserGroup field")
	ErrPasswdCreateAndPrimaryGroup = errors.New("cannot use both the create object and the user-level primaryGroup field")
	ErrPasswdCreateAndShell        = errors.New("cannot use both the create object and the user-level shell field")
	ErrPasswdCreateAndSystem       = errors.New("cannot use both the create object and the user-level system field")
	ErrPasswdCreateAndUID          = errors.New("cannot use both the create object and the user-level uid field")
)

func (p PasswdUser) Validate() report.Report {
	r := report.Report{}
	if p.Create != nil {
		r.Add(report.Entry{
			Message: ErrPasswdCreateDeprecated.Error(),
			Kind:    report.EntryWarning,
		})
		addErr := func(err error) {
			r.Add(report.Entry{
				Message: err.Error(),
				Kind:    report.EntryError,
			})
		}
		if p.Gecos != "" {
			addErr(ErrPasswdCreateAndGecos)
		}
		if len(p.Groups) > 0 {
			addErr(ErrPasswdCreateAndGroups)
		}
		if p.HomeDir != "" {
			addErr(ErrPasswdCreateAndHomeDir)
		}
		if p.NoCreateHome {
			addErr(ErrPasswdCreateAndNoCreateHome)
		}
		if p.NoLogInit {
			addErr(ErrPasswdCreateAndNoLogInit)
		}
		if p.NoUserGroup {
			addErr(ErrPasswdCreateAndNoUserGroup)
		}
		if p.PrimaryGroup != "" {
			addErr(ErrPasswdCreateAndPrimaryGroup)
		}
		if p.Shell != "" {
			addErr(ErrPasswdCreateAndShell)
		}
		if p.System {
			addErr(ErrPasswdCreateAndSystem)
		}
		if p.UID != nil {
			addErr(ErrPasswdCreateAndUID)
		}
	}
	return r
}

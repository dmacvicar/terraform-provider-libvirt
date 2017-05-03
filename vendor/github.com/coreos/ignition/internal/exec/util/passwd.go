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
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/coreos/ignition/config/types"

	keys "github.com/coreos/update-ssh-keys/authorized_keys_d"
)

// CreateUser creates the user as described.
func (u Util) CreateUser(c types.PasswdUser) error {
	if c.Create == nil {
		return nil
	}

	cu := c.Create
	args := []string{"--root", u.DestDir}

	if c.PasswordHash != "" {
		args = append(args, "--password", c.PasswordHash)
	} else {
		args = append(args, "--password", "*")
	}

	if cu.UID != nil {
		args = append(args, "--uid",
			strconv.FormatUint(uint64(*cu.UID), 10))
	}

	if cu.Gecos != "" {
		args = append(args, "--comment", fmt.Sprintf("%q", cu.Gecos))
	}

	if cu.HomeDir != "" {
		args = append(args, "--home-dir", cu.HomeDir)
	}

	if cu.NoCreateHome {
		args = append(args, "--no-create-home")
	} else {
		args = append(args, "--create-home")
	}

	if cu.PrimaryGroup != "" {
		args = append(args, "--gid", cu.PrimaryGroup)
	}

	if len(cu.Groups) > 0 {
		args = append(args, "--groups", strings.Join(translateV2_1UsercreateGroupSliceToStringSlice(cu.Groups), ","))
	}

	if cu.NoUserGroup {
		args = append(args, "--no-user-group")
	}

	if cu.System {
		args = append(args, "--system")
	}

	if cu.NoLogInit {
		args = append(args, "--no-log-init")
	}

	if cu.Shell != "" {
		args = append(args, "--shell", cu.Shell)
	}

	args = append(args, c.Name)

	return u.LogCmd(exec.Command("useradd", args...),
		"creating user %q", c.Name)
}

// golang--
func translateV2_1UsercreateGroupSliceToStringSlice(groups []types.UsercreateGroup) []string {
	newGroups := make([]string, len(groups))
	for i, g := range groups {
		newGroups[i] = string(g)
	}
	return newGroups
}

// Add the provided SSH public keys to the user's authorized keys.
func (u Util) AuthorizeSSHKeys(c types.PasswdUser) error {
	if len(c.SSHAuthorizedKeys) == 0 {
		return nil
	}

	return u.LogOp(func() error {
		usr, err := u.userLookup(c.Name)
		if err != nil {
			return fmt.Errorf("unable to lookup user %q", c.Name)
		}

		akd, err := keys.Open(usr, true)
		if err != nil {
			return err
		}
		defer akd.Close()

		// TODO(vc): introduce key names to config?
		// TODO(vc): validate c.SSHAuthorizedKeys well-formedness.
		ks := strings.Join(translateV2_1SSHAuthorizedKeySliceToStringSlice(c.SSHAuthorizedKeys), "\n")
		// XXX(vc): for now ensure the addition is always
		// newline-terminated.  A future version of akd will handle this
		// for us in addition to validating the ssh keys for
		// well-formedness.
		if !strings.HasSuffix(ks, "\n") {
			ks = ks + "\n"
		}

		if err := akd.Add("coreos-ignition", []byte(ks), true, true); err != nil {
			return err
		}

		if err := akd.Sync(); err != nil {
			return err
		}

		return nil
	}, "adding ssh keys to user %q", c.Name)
}

// golang--
func translateV2_1SSHAuthorizedKeySliceToStringSlice(keys []types.SSHAuthorizedKey) []string {
	newKeys := make([]string, len(keys))
	for i, k := range keys {
		newKeys[i] = string(k)
	}
	return newKeys
}

// SetPasswordHash sets the password hash of the specified user.
func (u Util) SetPasswordHash(c types.PasswdUser) error {
	if c.PasswordHash == "" {
		return nil
	}

	args := []string{
		"--root", u.DestDir,
		"--password", c.PasswordHash,
	}

	args = append(args, c.Name)

	return u.LogCmd(exec.Command("usermod", args...),
		"setting password for %q", c.Name)
}

// CreateGroup creates the group as described.
func (u Util) CreateGroup(g types.PasswdGroup) error {
	args := []string{"--root", u.DestDir}

	if g.Gid != nil {
		args = append(args, "--gid",
			strconv.FormatUint(uint64(*g.Gid), 10))
	}

	if g.PasswordHash != "" {
		args = append(args, "--password", g.PasswordHash)
	} else {
		args = append(args, "--password", "*")
	}

	if g.System {
		args = append(args, "--system")
	}

	args = append(args, g.Name)

	return u.LogCmd(exec.Command("groupadd", args...),
		"adding group %q", g.Name)
}

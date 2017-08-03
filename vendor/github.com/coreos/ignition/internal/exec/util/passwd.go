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

// EnsureUser ensures that the user exists as described. If the user does not
// yet exist, they will be created, otherwise the existing user will be
// modified.
func (u Util) EnsureUser(c types.PasswdUser) error {
	exists, err := u.CheckIfUserExists(c)
	if err != nil {
		return err
	}
	if c.Create != nil {
		cu := c.Create
		c.Gecos = cu.Gecos
		c.Groups = translateV2_1UsercreateGroupSliceToPasswdUserGroupSlice(cu.Groups)
		c.HomeDir = cu.HomeDir
		c.NoCreateHome = cu.NoCreateHome
		c.NoLogInit = cu.NoLogInit
		c.NoUserGroup = cu.NoUserGroup
		c.PrimaryGroup = cu.PrimaryGroup
		c.Shell = cu.Shell
		c.System = cu.System
		c.UID = cu.UID
	}
	args := []string{"--root", u.DestDir}

	var cmd string
	if exists {
		cmd = "usermod"

		if c.HomeDir != "" {
			args = append(args, "--home", c.HomeDir, "--move-home")
		}
	} else {
		cmd = "useradd"

		if c.HomeDir != "" {
			args = append(args, "--home-dir", c.HomeDir)
		}

		if c.NoCreateHome {
			args = append(args, "--no-create-home")
		} else {
			args = append(args, "--create-home")
		}

		if c.NoUserGroup {
			args = append(args, "--no-user-group")
		}

		if c.System {
			args = append(args, "--system")
		}

		if c.NoLogInit {
			args = append(args, "--no-log-init")
		}
	}

	if c.PasswordHash != nil {
		if *c.PasswordHash != "" {
			args = append(args, "--password", *c.PasswordHash)
		} else {
			args = append(args, "--password", "*")
		}
	} else if !exists {
		// Set the user's password to "*" if they don't exist yet and one wasn't
		// set to disable password logins
		args = append(args, "--password", "*")
	}

	if c.UID != nil {
		args = append(args, "--uid",
			strconv.FormatUint(uint64(*c.UID), 10))
	}

	if c.Gecos != "" {
		args = append(args, "--comment", fmt.Sprintf("%q", c.Gecos))
	}

	if c.PrimaryGroup != "" {
		args = append(args, "--gid", c.PrimaryGroup)
	}

	if len(c.Groups) > 0 {
		args = append(args, "--groups", strings.Join(translateV2_1PasswdUserGroupSliceToStringSlice(c.Groups), ","))
	}

	if c.Shell != "" {
		args = append(args, "--shell", c.Shell)
	}

	args = append(args, c.Name)

	_, err = u.LogCmd(exec.Command(cmd, args...),
		"creating or modifying user %q", c.Name)
	return err
}

// golang--
func translateV2_1UsercreateGroupSliceToPasswdUserGroupSlice(groups []types.UsercreateGroup) []types.PasswdUserGroup {
	newGroups := make([]types.PasswdUserGroup, len(groups))
	for i, g := range groups {
		newGroups[i] = types.PasswdUserGroup(g)
	}
	return newGroups
}

func (u Util) CheckIfUserExists(c types.PasswdUser) (bool, error) {
	code, err := u.LogCmd(exec.Command("chroot", u.DestDir, "id", c.Name),
		"checking if user %q exists", c.Name)
	if err != nil {
		if code == 1 {
			return false, nil
		}
		u.Logger.Info("error encountered (%+T): %v", err, err)
		return false, err
	}
	return true, nil
}

// golang--
func translateV2_1PasswdUserGroupSliceToStringSlice(groups []types.PasswdUserGroup) []string {
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
	if c.PasswordHash == nil {
		return nil
	}

	pwhash := *c.PasswordHash
	if *c.PasswordHash == "" {
		pwhash = "*"
	}

	args := []string{
		"--root", u.DestDir,
		"--password", pwhash,
	}

	args = append(args, c.Name)

	_, err := u.LogCmd(exec.Command("usermod", args...),
		"setting password for %q", c.Name)
	return err
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

	_, err := u.LogCmd(exec.Command("groupadd", args...),
		"adding group %q", g.Name)
	return err
}

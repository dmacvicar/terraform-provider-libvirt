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

// +build linux

// authorized_keys_d manages a user's ~/.ssh/authorized_keys.d and can produce
// a ~/.ssh/authorized_keys file from the authorized_keys.d contents.
package authorized_keys_d

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/coreos/update-ssh-keys/authorized_keys_d/as_user"
)

const (
	AuthorizedKeysFile = "authorized_keys"
	AuthorizedKeysDir  = "authorized_keys.d"
	PreservedKeysName  = "orig_authorized_keys"
	SSHDir             = ".ssh"

	lockFile  = ".authorized_keys.d.lock"       // In "~/".
	stageFile = ".authorized_keys.d.stage_file" // In "~/.ssh".
	stageDir  = ".authorized_keys.d.stage_dir"  // In "~/.ssh".
)

// SSHAuthorizedKeysDir represents an opened user's authorized_keys.d.
type SSHAuthorizedKeysDir struct {
	path string     // Path to authorized_keys.d directory.
	user *user.User // User of the directory.
	lock *os.File   // Lock file for serializing Open()-Close().
}

// SSHAuthorizedKey represents an opened user's authorized_keys.d/<name> entry.
type SSHAuthorizedKey struct {
	Name     string                // Name given to the key.
	Disabled bool                  // Disabled state of the key.
	Path     string                // Path to the file backing the key.
	origin   *SSHAuthorizedKeysDir // Originating dir for this key.
}

// sshDirPath returns the path to the .ssh dir for the user.
func sshDirPath(u *user.User) string {
	return filepath.Join(u.HomeDir, SSHDir)
}

// authKeysFilePath returns the path to the authorized_keys file for the user.
func authKeysFilePath(u *user.User) string {
	return filepath.Join(sshDirPath(u), AuthorizedKeysFile)
}

// authKeysDirPath returns the path to the authorized_keys.d for the user.
func authKeysDirPath(u *user.User) string {
	return filepath.Join(sshDirPath(u), AuthorizedKeysDir)
}

// lockFilePath returns the path to the lock file for the user.
func lockFilePath(u *user.User) string {
	return filepath.Join(u.HomeDir, lockFile)
}

// stageDirPath returns the path to the staging directory for the user.
func stageDirPath(u *user.User) string {
	return filepath.Join(sshDirPath(u), stageDir)
}

// stageFilePath returns the path to the staging file for the user.
func stageFilePath(u *user.User) string {
	return filepath.Join(sshDirPath(u), stageFile)
}

// opendir opens the authorized keys directory.
func opendir(dir string) (*SSHAuthorizedKeysDir, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("%q is not a directory", dir)
	}
	return &SSHAuthorizedKeysDir{path: dir}, nil
}

// acquireLock locks the lock file for the given user's authorized_keys.d.
// A lock file is created if it doesn't already exist.
// The locking is currently a simple coarse-grained mutex held for the
// Open()-Close() duration, implemented using a lock file in the user's ~/.
func acquireLock(u *user.User) (*os.File, error) {
	f, err := as_user.OpenFile(u, lockFilePath(u),
		syscall.O_CREAT|syscall.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

// createAuthorizedKeysDir creates an authorized keys directory for the user.
// If the user has an authorized_keys file, it is migrated.
func createAuthorizedKeysDir(u *user.User) (*SSHAuthorizedKeysDir, error) {
	td := stageDirPath(u)
	if err := as_user.MkdirAll(u, td, 0700); err != nil {
		return nil, err
	}
	defer os.RemoveAll(td)

	akd, err := opendir(td)
	if err != nil {
		return nil, err
	}
	akd.user = u

	akfb, err := ioutil.ReadFile(authKeysFilePath(u))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	} else if err == nil {
		err = akd.Add(PreservedKeysName, akfb, false, false)
		if err != nil {
			return nil, err
		}
	}
	if err = akd.rename(authKeysDirPath(u)); err != nil {
		return nil, err
	}
	return akd, err
}

// Open opens the authorized keys directory for the supplied user.
// If create is false, Open will fail if no directory exists yet.
// If create is true, Open will create the directory if it doesn't exist,
// preserving the authorized_keys file in the process.
// After a successful open, Close should be called when finished to unlock
// the directory.
func Open(usr *user.User, create bool) (*SSHAuthorizedKeysDir, error) {
	l, err := acquireLock(usr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			l.Close()
		}
	}()

	akd, err := opendir(authKeysDirPath(usr))
	if err != nil && (!create || !os.IsNotExist(err)) {
		return nil, err
	} else if os.IsNotExist(err) {
		akd, err = createAuthorizedKeysDir(usr)
		if err != nil {
			return nil, err
		}
	}

	akd.lock = l
	akd.user = usr
	return akd, nil
}

// Close closes the authorized keys directory.
func (akd *SSHAuthorizedKeysDir) Close() error {
	return akd.lock.Close()
}

// rename renames the authorized_keys dir to the supplied path.
func (akd *SSHAuthorizedKeysDir) rename(to string) error {
	err := as_user.Rename(akd.user, akd.path, to)
	if err != nil {
		return err
	}
	akd.path = to
	return nil
}

// keyPath returns the path to the named key.
func (akd *SSHAuthorizedKeysDir) keyPath(n string) string {
	return filepath.Join(akd.path, n)
}

// WalkKeys iterates across all keys in akd, calling f for each key.
// Iterating stops on error, and the error is propagated out.
func (akd *SSHAuthorizedKeysDir) WalkKeys(f func(*SSHAuthorizedKey) error) error {
	d, err := os.Open(akd.path)
	if err != nil {
		return err
	}

	names, err := d.Readdirnames(0)
	if err != nil {
		return err
	}

	sort.Strings(names)
	for _, n := range names {
		ak, err := akd.Open(n)
		if err != nil {
			return err
		}
		if err := f(ak); err != nil {
			return err
		}
	}

	return nil
}

// Open opens the key at name.
func (akd *SSHAuthorizedKeysDir) Open(name string) (*SSHAuthorizedKey, error) {
	p := akd.keyPath(name)
	fi, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	ak := &SSHAuthorizedKey{
		Name:     name,
		Disabled: (fi.Size() == 0),
		Path:     p,
		origin:   akd,
	}
	return ak, nil
}

// Remove removes the key at name.
func (akd *SSHAuthorizedKeysDir) Remove(name string) error {
	ak, err := akd.Open(name)
	if err != nil {
		return err
	}
	return ak.Remove()
}

// Disable disables the key at name.
func (akd *SSHAuthorizedKeysDir) Disable(name string) error {
	ak, err := akd.Open(name)
	if err != nil {
		return err
	}
	return ak.Disable()
}

// Add adds the supplied key at name.
// replace enables replacing keys already existing at name.
// force enables adding keys to a disabled name, enabling it in the process.
// Names starting wtih ".", and anything containing "/" are disallowed.
func (akd *SSHAuthorizedKeysDir) Add(name string, keys []byte, replace, force bool) error {
	if strings.HasPrefix(name, ".") || strings.Contains(name, "/") {
		return fmt.Errorf(`illegal name`)
	}

	p := akd.keyPath(name)
	fi, err := os.Stat(p)
	if err == nil {
		if fi.Size() > 0 && !replace {
			return fmt.Errorf("key %q already exists", name)
		} else if fi.Size() == 0 && !force {
			return fmt.Errorf("key %q disabled", name)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	ak := &SSHAuthorizedKey{Path: p, origin: akd}
	return ak.Replace(keys)
}

// KeysFilePath returns the backing authorized_keys file path for this
// SSHAuthorizedKeysDir.  This is the file written to by Sync().
func (akd *SSHAuthorizedKeysDir) KeysFilePath() string {
	return authKeysFilePath(akd.user)
}

// KeysDirPath returns the authorized_keys.d directory path for this
// SSHAuthorizedKeysDir.  This is the directory containing the discrete key
// files.
func (akd *SSHAuthorizedKeysDir) KeysDirPath() string {
	return authKeysDirPath(akd.user)
}

// Sync synchronizes the user's ~/.ssh/authorized_keys file with the
// current authorized_keys.d directory state.
func (akd *SSHAuthorizedKeysDir) Sync() error {
	sp := stageFilePath(akd.user)
	sf, err := as_user.OpenFile(akd.user, sp,
		syscall.O_CREAT|syscall.O_TRUNC|syscall.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			sf.Close()
			os.Remove(sp)
		}
	}()

	if err := akd.WalkKeys(func(k *SSHAuthorizedKey) error {
		if !k.Disabled {
			kb, err := ioutil.ReadFile(k.Path)
			if err != nil {
				return err
			}
			kb = append(kb, '\n')
			if _, err := sf.Write(kb); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := sf.Close(); err != nil {
		return err
	}

	err = as_user.Rename(akd.user, sp, authKeysFilePath(akd.user))
	if err != nil {
		return err
	}

	return nil
}

// Remove removes the opened key.
func (ak *SSHAuthorizedKey) Remove() error {
	return os.Remove(ak.Path)
}

// Disable disables the opened key.
func (ak *SSHAuthorizedKey) Disable() error {
	return os.Truncate(ak.Path, 0)
}

// Replace replaces the opened key with the supplied data.
func (ak *SSHAuthorizedKey) Replace(keys []byte) error {
	sp := stageFilePath(ak.origin.user)
	sf, err := as_user.OpenFile(ak.origin.user, sp,
		syscall.O_WRONLY|syscall.O_CREAT|syscall.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer os.Remove(sp)
	if _, err = sf.Write(keys); err != nil {
		return err
	}
	if err := sf.Close(); err != nil {
		return err
	}
	return as_user.Rename(ak.origin.user, sp, ak.Path)
}

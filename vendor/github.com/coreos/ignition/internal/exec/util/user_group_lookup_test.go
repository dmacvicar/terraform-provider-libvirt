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
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/coreos/ignition/internal/log"
)

// tempBase() slaps together a minimal /etc/{passwd,group} for the lookup test.
func tempBase() (string, error) {
	td, err := ioutil.TempDir("", "ign-usr-lookup-test")
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Join(td, "etc"), 0755); err != nil {
		return "", err
	}

	gp := filepath.Join(td, "etc/group")
	err = ioutil.WriteFile(gp, []byte("foo:x:4242:\n"), 0644)
	if err != nil {
		return "", err
	}

	pp := filepath.Join(td, "etc/passwd")
	err = ioutil.WriteFile(pp, []byte("foo:x:44:4242::/home/foo:/bin/false"), 0644)
	if err != nil {
		return "", err
	}

	nsp := filepath.Join(td, "etc/nsswitch.conf")
	err = ioutil.WriteFile(nsp, []byte("passwd: files\ngroup: files\nshadow: files\ngshadow: files\n"), 0644)
	if err != nil {
		return "", err
	}

	return td, nil
}

func TestUserLookup(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("test requires root for chroot(), skipping")
	}

	// perform a user lookup to ensure libnss_files.so is loaded
	// note this assumes /etc/nsswitch.conf invokes files.
	user.Lookup("root")

	td, err := tempBase()
	if err != nil {
		t.Fatalf("temp base error: %v", err)
	}
	defer os.RemoveAll(td)

	logger := log.New()
	defer logger.Close()

	u := &Util{
		DestDir: td,
		Logger:  &logger,
	}

	usr, err := u.userLookup("foo")
	if err != nil {
		t.Fatalf("lookup error: %v", err)
	}

	if usr.Name != "foo" {
		t.Fatalf("unexpected name: %q", usr.Name)
	}

	if usr.Uid != "44" {
		t.Fatalf("unexpected uid: %q", usr.Uid)
	}

	if usr.Gid != "4242" {
		t.Fatalf("unexpected gid: %q", usr.Gid)
	}
}

func TestGroupLookup(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("test requires root for chroot(), skipping")
	}

	td, err := tempBase()
	if err != nil {
		t.Fatalf("temp base error: %v", err)
	}
	defer os.RemoveAll(td)

	logger := log.New()
	defer logger.Close()

	u := &Util{
		DestDir: td,
		Logger:  &logger,
	}

	grp, err := u.groupLookup("foo")
	if err != nil {
		t.Fatalf("lookup error: %v", err)
	}

	if grp.Name != "foo" {
		t.Fatalf("unexpected name: %q", grp.Name)
	}

	if grp.Gid != "4242" {
		t.Fatalf("unexpected gid: %q", grp.Gid)
	}
}

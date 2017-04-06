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

package authorized_keys_d

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

const (
	DirNode = iota
	FileNode
)

type tree []node
type node struct {
	path  string
	typ   int
	mode  os.FileMode
	bytes []byte
}

// create creates the tree rooted at the path where.
// If tree is empty, where will not be created.
func (t tree) create(where string) error {
	if len(t) == 0 {
		os.RemoveAll(where)
		return nil
	}

	for i, n := range t {
		p := filepath.Join(where, n.path)
		switch n.typ {
		case DirNode:
			if err := os.MkdirAll(p, n.mode); err != nil {
				return fmt.Errorf("%d: mkdir error: %v", i, err)
			}
			if err := os.Chmod(p, n.mode); err != nil {
				return fmt.Errorf("%d: chown error: %v", i, err)
			}
		case FileNode:
			err := ioutil.WriteFile(p, n.bytes, n.mode)
			if err != nil {
				return fmt.Errorf("%d: writefile error: %v", i, err)
			}
		default:
			return fmt.Errorf("%d: unknown node type: %v", i, n.typ)
		}
	}

	return nil
}

// testUser populates a user.User for testing.
func testUser(home string) *user.User {
	return &user.User{
		Uid:      strconv.Itoa(os.Getuid()),
		Gid:      strconv.Itoa(os.Getgid()),
		Username: "test",
		Name:     "test",
		HomeDir:  home,
	}
}

// testDir creates a temp directory for testing.
func testDir(t *testing.T) string {
	td, err := ioutil.TempDir("", "akd-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return td
}

func TestOpenClose(t *testing.T) {
	tests := []struct {
		create  bool
		home    tree
		success bool
	}{
		{ // 0: No home, no create, fail.
		},
		{ // 1: No home, yes create, fail.
			create: true,
		},
		{ // 2: Empty home, no create, fail.
			create:  false,
			home:    tree{{path: ".", typ: DirNode, mode: 0700}},
			success: false,
		},
		{ // 3: Empty home, yes create, success.
			create:  true,
			home:    tree{{path: ".", typ: DirNode, mode: 0700}},
			success: true,
		},
		{ // 4: Empty home, no create, fail.
			create:  false,
			home:    tree{{path: ".", typ: DirNode, mode: 0700}},
			success: false,
		},
		{ // 5: Intact home, no create, fail.
			create: false,
			home: tree{
				{path: ".", typ: DirNode, mode: 0700},
				{path: SSHDir, typ: DirNode, mode: 0700},
				{
					path: filepath.Join(SSHDir, AuthorizedKeysFile),
					typ:  FileNode, mode: 0600,
				},
			},
			success: false,
		},
		{ // 6: Intact home, yes create, succeed.
			create: true,
			home: tree{
				{path: ".", typ: DirNode, mode: 0700},
				{path: SSHDir, typ: DirNode, mode: 0700},
				{
					path: filepath.Join(SSHDir, AuthorizedKeysFile),
					typ:  FileNode, mode: 0600,
				},
			},
			success: true,
		},
		{ // 7: Managed home, no create, succeed.
			create: false,
			home: tree{
				{path: ".", typ: DirNode, mode: 0700},
				{path: SSHDir, typ: DirNode, mode: 0700},
				{
					path: filepath.Join(SSHDir, AuthorizedKeysFile),
					typ:  FileNode, mode: 0600,
				},
				{
					path: filepath.Join(SSHDir, AuthorizedKeysDir),
					typ:  DirNode, mode: 0700,
				},
			},
			success: true,
		},
		{ // 8: Managed home, yes create, succeed.
			create: true,
			home: tree{
				{path: ".", typ: DirNode, mode: 0700},
				{path: SSHDir, typ: DirNode, mode: 0700},
				{
					path: filepath.Join(SSHDir, AuthorizedKeysFile),
					typ:  FileNode, mode: 0600,
				},
				{
					path: filepath.Join(SSHDir, AuthorizedKeysDir),
					typ:  DirNode, mode: 0700,
				},
			},
			success: true,
		},
	}

	for i, tt := range tests {
		td := testDir(t)
		if err := tt.home.create(td); err != nil {
			t.Errorf("%d: failed to create home tree: %v", i, err)
			continue
		}

		akd, err := Open(testUser(td), tt.create)
		if (err == nil) != tt.success {
			t.Errorf("%d: expected %v got %v", i, tt.success, bool(err == nil))
			if err != nil {
				t.Errorf("%d: unexpected error: %v", i, err)
			}
		}
		if err == nil {
			if err := akd.Close(); err != nil {
				t.Errorf("%d: close failed: %v", i, err)
			}
		}
		os.RemoveAll(td)
	}
}

func TestAddKey(t *testing.T) {
	tests := []struct {
		name    string
		disable bool
		replace bool
		force   bool
		success bool
	}{
		{ // 0: Doesn't exist, no replace, success.
			name:    "test",
			success: true,
		},
		{ // 1: Exists, no replace, fail.
			name: "test",
		},
		{ // 2: Exists, replace, succeed.
			name:    "test",
			replace: true,
			success: true,
		},
		{ // 2: Exists (disabled), force, succeed.
			name:    "test",
			disable: true,
			force:   true,
			success: true,
		},
		{ // 2: Exists (disabled), !force, fail.
			name:    "test",
			disable: true,
		},
		{ // 3: Leading dot, fail.
			name: ".hidden",
		},
		{ // 4: Leading dot, fail.
			name: ".",
		},
		{ // 5: Leading dot, fail.
			name: "..",
		},
		{ // 6: Contains slash, fail.
			name: "no/slash",
		},
	}

	td := testDir(t)
	akd, err := Open(testUser(td), true)
	if err != nil {
		t.Fatalf("open error: %v", err)
	}
	defer func() {
		akd.Close()
		os.RemoveAll(td)
	}()

	for i, tt := range tests {
		if tt.disable {
			if err := akd.Disable(tt.name); err != nil {
				t.Errorf("%d: disable failed: %v", i, err)
			}
		}
		err := akd.Add(tt.name, []byte("test"), tt.replace, tt.force)
		if (err == nil) != tt.success {
			t.Errorf("%d: expected %v got %v", i, tt.success, bool(err == nil))
		}
	}
}

func TestWalkKeys(t *testing.T) {
	tests := []struct {
		names    []string
		asWalked []string
	}{
		{ // 0: Empty, in-order.
			names: []string{},
		},
		{ // 1: One, in-order.
			names: []string{"0_one"},
		},
		{ // 2: Many, in-order.
			names: []string{"0_one", "1_two", "2_three"},
		},
		{ // 3: Many, out-of-order.
			// XXX: add order reflecting readdir order is fs-specific, harmless to try.
			names:    []string{"0_one", "2_three", "1_two"},
			asWalked: []string{"0_one", "1_two", "2_three"},
		},
	}

	for i, tt := range tests {
		td := testDir(t)
		akd, err := Open(testUser(td), true)
		if err != nil {
			t.Errorf("%d: open error: %v", i, err)
			os.RemoveAll(td)
			continue
		}

		for _, n := range tt.names {
			err := akd.Add(n, []byte("foobar"), false, false)
			if err != nil {
				t.Errorf("%d: add error: %v", i, err)
			}
		}

		walked := []string{}
		if err := akd.WalkKeys(func(k *SSHAuthorizedKey) error {
			walked = append(walked, k.Name)
			return nil
		}); err != nil {
			t.Errorf("%d: walk error: %v", i, err)
		}

		against := tt.names
		if len(tt.asWalked) != 0 {
			against = tt.asWalked
		}

		if !reflect.DeepEqual(walked, against) {
			t.Errorf("%d: inconsistent result: %v vs. %v", i, walked, against)
		}

		akd.Close()
		os.RemoveAll(td)
	}
}

func TestOpenName(t *testing.T) {
	tests := []struct {
		names   []string
		open    string
		success bool
	}{
		{ // 0: Empty, absent, fail.
			open:    "missing",
			success: false,
		},
		{ // 1: Populated, absent, fail.
			names:   []string{"a", "b", "c"},
			open:    "missing",
			success: false,
		},
		{ // 2: Populated, present, succeed.
			names:   []string{"a", "b", "c"},
			open:    "b",
			success: true,
		},
	}

	for i, tt := range tests {
		td := testDir(t)
		akd, err := Open(testUser(td), true)
		if err != nil {
			t.Errorf("%d: open error: %v", i, err)
			os.RemoveAll(td)
			continue
		}

		for _, n := range tt.names {
			err := akd.Add(n, []byte("foobar"), false, false)
			if err != nil {
				t.Errorf("%d: add error: %v", i, err)
			}
		}

		k, err := akd.Open(tt.open)
		if (err == nil) != tt.success {
			t.Errorf("%d: expected %v got %v", i, tt.success, bool(err == nil))
		}
		if err == nil && k.Name != tt.open {
			t.Errorf("%d: expected %v got %v", i, tt.open, k.Name)
		}

		akd.Close()
		os.RemoveAll(td)
	}
}

func TestRemoveName(t *testing.T) {
	tests := []struct {
		names   []string
		remove  string
		success bool
	}{
		{ // 0: Empty, absent, fail.
			remove:  "missing",
			success: false,
		},
		{ // 1: Populated, absent, fail.
			names:   []string{"a", "b", "c"},
			remove:  "missing",
			success: false,
		},
		{ // 2: Populated, present, succeed.
			names:   []string{"a", "b", "c"},
			remove:  "b",
			success: true,
		},
	}

	for i, tt := range tests {
		td := testDir(t)
		akd, err := Open(testUser(td), true)
		if err != nil {
			t.Errorf("%d: open error: %v", i, err)
			os.RemoveAll(td)
			continue
		}

		for _, n := range tt.names {
			err := akd.Add(n, []byte("foobar"), false, false)
			if err != nil {
				t.Errorf("%d: add error: %v", i, err)
			}
		}

		err = akd.Remove(tt.remove)
		if (err == nil) != tt.success {
			t.Errorf("%d: expected %v got %v", i, tt.success, bool(err == nil))
		}

		akd.Close()
		os.RemoveAll(td)
	}
}

func TestDisableName(t *testing.T) {
	name := "test"
	td := testDir(t)
	defer os.RemoveAll(td)
	akd, err := Open(testUser(td), true)
	if err != nil {
		t.Errorf("open error: %v", err)
		return
	}
	defer akd.Close()

	if err := akd.Add(name, []byte("foobar"), false, false); err != nil {
		t.Errorf("add error: %v", err)
		return
	}

	ak, err := akd.Open(name)
	if err != nil {
		t.Errorf("open name error: %v", err)
		return
	}

	if ak.Disabled {
		t.Errorf("unexpectedly disabled")
		return
	}

	if err := akd.Disable(name); err != nil {
		t.Errorf("disable failed: %v", err)
		return
	}

	ak, err = akd.Open(name)
	if err != nil {
		t.Errorf("open name error: %v", err)
		return
	}

	if !ak.Disabled {
		t.Errorf("unexpected enabled")
		return
	}
}

func TestSync(t *testing.T) {
	td := testDir(t)
	defer os.RemoveAll(td)
	akd, err := Open(testUser(td), true)
	if err != nil {
		t.Errorf("open error: %v", err)
		return
	}
	defer akd.Close()

	bytes := []byte{}
	for i := 0; i < 10; i++ {
		b := make([]byte, 512)
		if _, err := rand.Read(b); err != nil {
			t.Errorf("rand read err: %v", err)
			return
		}
		if err := akd.Add(strconv.Itoa(i), b, false, false); err != nil {
			t.Errorf("add error: %v", err)
			return
		}
		bytes = append(bytes, b...)
		bytes = append(bytes, '\n')
	}

	if err := akd.Sync(); err != nil {
		t.Errorf("sync failed: %v", err)
		return
	}

	rb, err := ioutil.ReadFile(akd.KeysFilePath())
	if err != nil {
		t.Errorf("read failed: %v", err)
		return
	}

	if !reflect.DeepEqual(bytes, rb) {
		t.Errorf("inconsistent output")
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		names   []string
		remove  string
		success bool
	}{
		{ // 0: Empty, absent, fail.
			remove:  "missing",
			success: false,
		},
		{ // 1: Populated, absent, fail.
			names:   []string{"a", "b", "c"},
			remove:  "missing",
			success: false,
		},
		{ // 2: Populated, present, succeed.
			names:   []string{"a", "b", "c"},
			remove:  "b",
			success: true,
		},
	}

	for i, tt := range tests {
		td := testDir(t)
		akd, err := Open(testUser(td), true)
		if err != nil {
			t.Errorf("%d: open error: %v", i, err)
			os.RemoveAll(td)
			continue
		}

		for _, n := range tt.names {
			err := akd.Add(n, []byte("foobar"), false, false)
			if err != nil {
				t.Errorf("%d: add error: %v", i, err)
			}
		}

		ak, err := akd.Open(tt.remove)
		if (err == nil) != tt.success {
			t.Errorf("%d: expected %v got %v", i, tt.success, bool(err == nil))
		}
		if err != nil {
			akd.Close()
			os.RemoveAll(td)
			continue
		}

		err = ak.Remove()
		if (err == nil) != tt.success {
			t.Errorf("%d: expected %v got %v", i, tt.success, bool(err == nil))
		}

		akd.Close()
		os.RemoveAll(td)
	}
}

func TestDisable(t *testing.T) {
	name := "test"
	td := testDir(t)
	defer os.RemoveAll(td)
	akd, err := Open(testUser(td), true)
	if err != nil {
		t.Errorf("open error: %v", err)
		return
	}
	defer akd.Close()

	if err := akd.Add(name, []byte("foobar"), false, false); err != nil {
		t.Errorf("add error: %v", err)
		return
	}

	ak, err := akd.Open(name)
	if err != nil {
		t.Errorf("open name error: %v", err)
		return
	}

	if ak.Disabled {
		t.Errorf("unexpectedly disabled")
		return
	}

	if err := ak.Disable(); err != nil {
		t.Errorf("disable failed: %v", err)
		return
	}

	ak, err = akd.Open(name)
	if err != nil {
		t.Errorf("open name error: %v", err)
		return
	}

	if !ak.Disabled {
		t.Errorf("unexpected enabled")
		return
	}
}

func TestReplace(t *testing.T) {
	name := "test"
	td := testDir(t)
	defer os.RemoveAll(td)
	akd, err := Open(testUser(td), true)
	if err != nil {
		t.Errorf("open error: %v", err)
		return
	}
	defer akd.Close()

	if err := akd.Add(name, []byte("foobar"), false, false); err != nil {
		t.Errorf("add error: %v", err)
		return
	}

	ak, err := akd.Open(name)
	if err != nil {
		t.Errorf("open name error: %v", err)
		return
	}

	if err := ak.Replace([]byte("moop")); err != nil {
		t.Errorf("replace failed: %v", err)
		return
	}

	ak, err = akd.Open(name)
	if err != nil {
		t.Errorf("open name error: %v", err)
		return
	}

	b, err := ioutil.ReadFile(ak.Path)
	if err != nil {
		t.Errorf("readfile failed: %v", err)
		return
	}

	if string(b) != "moop" {
		t.Errorf("unexpected results: %q", string(b))
	}
}

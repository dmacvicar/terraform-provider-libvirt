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

package as_user

// #include "as_user.h"
import "C"

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

// TODO(vc): do this at akd.Open() and make our own typed user struct.
// getIds extracts the uid and guid as integers from a user.User.
func getIds(u *user.User) (C.int, C.int, error) {
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid uid: %v", err)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return -1, -1, fmt.Errorf("invalid gid: %v", err)
	}
	return C.int(uid), C.int(gid), nil
}

// OpenFile reimplements os.OpenFile but switching euid/egid first if necessary.
func OpenFile(u *user.User, name string, flags int, perm uint32) (*os.File, error) {
	uid, gid, err := getIds(u)
	if err != nil {
		return nil, err
	}
	args := C.au_open_args_t{
		ids:   C.au_ids_t{uid: uid, gid: gid},
		path:  C.CString(name),
		flags: C.uint(flags),
		mode:  C.uint(perm),
	}

	fd, err := C.au_open(&args)
	if fd < 0 {
		return nil, err
	}
	return os.NewFile(uintptr(fd), name), nil
}

// MkdirAll reimplements os.MkdirAll but switching euid/egid first if necessary.
func MkdirAll(u *user.User, path string, perm uint32) error {
	uid, gid, err := getIds(u)
	if err != nil {
		return err
	}
	args := C.au_mkdir_all_args_t{
		ids:  C.au_ids_t{uid: uid, gid: gid},
		path: C.CString(path),
		mode: C.uint(perm),
	}

	if ret, err := C.au_mkdir_all(&args); ret < 0 {
		return err
	}
	return nil
}

// Rename reimplements os.Rename but switching euid/egid first if necessary.
func Rename(u *user.User, oldpath, newpath string) error {
	uid, gid, err := getIds(u)
	if err != nil {
		return err
	}
	args := C.au_rename_args_t{
		ids:     C.au_ids_t{uid: uid, gid: gid},
		oldpath: C.CString(oldpath),
		newpath: C.CString(newpath),
	}

	if ret, err := C.au_rename(&args); ret < 0 {
		return err
	}
	return nil
}

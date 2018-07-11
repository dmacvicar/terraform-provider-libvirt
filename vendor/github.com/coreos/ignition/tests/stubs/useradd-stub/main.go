// Copyright 2018 CoreOS, Inc.
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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	flagRoot         string
	flagHomeDir      string
	flagCreateHome   bool
	flagNoCreateHome bool
	flagNoUserGroup  bool
	flagSystem       bool
	flagNoLogInit    bool
	flagPassword     string
	flagUid          int
	flagComment      string
	flagGid          int
	flagGroups       string
	flagShell        string
)

func main() {
	flag.StringVar(&flagRoot, "root", "", "Apply changes in the CHROOT_DIR directory and use the configuration files from the CHROOT_DIR directory")
	flag.StringVar(&flagHomeDir, "home-dir", "", "The new user will be created using HOME_DIR as the value for the user's login directory")
	flag.BoolVar(&flagCreateHome, "create-home", false, "Create the user's home directory if it does not exist.")
	flag.BoolVar(&flagNoCreateHome, "no-create-home", false, "Do no create the user's home directory")
	flag.BoolVar(&flagNoUserGroup, "no-user-group", false, "Do not create a group with the same name as the user")
	flag.BoolVar(&flagSystem, "system", false, "Create a system account")
	flag.BoolVar(&flagNoLogInit, "no-log-init", false, "Do not add the user to the lastlog and faillog databases")
	flag.StringVar(&flagPassword, "password", "", "The encrypted password, as returned by crypt")
	flag.IntVar(&flagUid, "uid", -1, "The numerical value of the user's ID")
	flag.StringVar(&flagComment, "comment", "", "Any text string. It is generally a short description of the login, and is currently used as the field for the user's full name.")
	flag.IntVar(&flagGid, "gid", -1, "The group name or number of the user's initial login group")
	flag.StringVar(&flagGroups, "groups", "", "A list of supplementary groups which the user is also a member of")
	flag.StringVar(&flagShell, "shell", "", "The name of the user's login shell")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Printf("incorrectly called\n")
		os.Exit(1)
	}

	username := flag.Args()[0]

	var uidGid int
	if flagUid == -1 || flagGid == -1 {
		var err error
		uidGid, err = getNextUidAndGid()
		if err != nil {
			fmt.Printf("error getting next uid/gid: %v\n", err)
			os.Exit(1)
		}
	}
	if flagUid == -1 {
		flagUid = uidGid
	}
	if flagGid == -1 {
		flagGid = uidGid
	}
	if flagHomeDir == "" {
		flagHomeDir = "/home/" + username
	}
	if flagShell == "" {
		flagShell = "/bin/bash"
	}
	if flagPassword == "" {
		flagPassword = "*"
	}

	passwdLine := fmt.Sprintf("%s:x:%d:%d:%s:%s:%s\n", username, flagUid, flagGid, flagComment, flagHomeDir, flagShell)

	passwdFile, err := os.OpenFile(path.Join(flagRoot, "/etc/passwd"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("couldn't open passwd file: %v\n", err)
		os.Exit(1)
	}
	defer passwdFile.Close()
	_, err = passwdFile.Write([]byte(passwdLine))
	if err != nil {
		fmt.Printf("couldn't write to passwd file: %v\n", err)
		os.Exit(1)
	}

	groupLine := fmt.Sprintf("%s:x:%d:\n", username, flagGid)

	groupFile, err := os.OpenFile(path.Join(flagRoot, "/etc/group"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("couldn't open group file: %v\n", err)
		os.Exit(1)
	}
	defer groupFile.Close()
	_, err = groupFile.Write([]byte(groupLine))
	if err != nil {
		fmt.Printf("couldn't write to group file: %v\n", err)
		os.Exit(1)
	}

	shadowLine := fmt.Sprintf("%s:%s:17331:0:99999:7:::\n", username, flagPassword)

	shadowFile, err := os.OpenFile(path.Join(flagRoot, "/etc/shadow"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("couldn't open shadow file: %v\n", err)
		os.Exit(1)
	}
	defer shadowFile.Close()
	_, err = shadowFile.Write([]byte(shadowLine))
	if err != nil {
		fmt.Printf("couldn't write to shadow file: %v\n", err)
		os.Exit(1)
	}

	gshadowLine := fmt.Sprintf("%s:!::\n", username)

	gshadowFile, err := os.OpenFile(path.Join(flagRoot, "etc/gshadow"), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("couldn't open gshadow file: %v\n", err)
		os.Exit(1)
	}
	defer gshadowFile.Close()
	_, err = gshadowFile.Write([]byte(gshadowLine))
	if err != nil {
		fmt.Printf("couldn't write to gshadow file: %v\n", err)
		os.Exit(1)
	}

	if !flagNoCreateHome {
		err = os.MkdirAll(path.Join(flagRoot, flagHomeDir), 0755)
		if err != nil {
			fmt.Printf("couldn't create home directory: %v\n", err)
			os.Exit(1)
		}
		err = os.Chown(path.Join(flagRoot, flagHomeDir), flagUid, flagGid)
		if err != nil {
			fmt.Printf("couldn't chown home directory: %v\n", err)
			os.Exit(1)
		}
	}
}

// getNextUidAndGid finds the next available uid/gid pair starting at 500. It
// returns an int that is both an available uid and an available gid.
func getNextUidAndGid() (int, error) {
	idMap := make(map[int]struct{})

	passwdContents, err := ioutil.ReadFile(path.Join(flagRoot, "/etc/passwd"))
	if err != nil {
		return -1, err
	}
	passwdLines := strings.Split(string(passwdContents), "\n")
	for i, l := range passwdLines {
		if i == len(passwdLines)-1 {
			// the last line is empty
			break
		}
		// Will panic due to out of bounds if /etc/passwd is malformed
		tokens := strings.Split(l, ":")
		uid, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			return -1, err
		}
		gid, err := strconv.ParseInt(tokens[3], 10, 64)
		if err != nil {
			return -1, err
		}
		idMap[int(uid)] = struct{}{}
		idMap[int(gid)] = struct{}{}
	}

	groupContents, err := ioutil.ReadFile(path.Join(flagRoot, "/etc/group"))
	if err != nil {
		return -1, err
	}
	groupLines := strings.Split(string(groupContents), "\n")
	for i, l := range groupLines {
		if i == len(groupLines)-1 {
			// the last line is empty
			break
		}
		// Will panic due to out of bounds if /etc/group is malformed
		tokens := strings.Split(l, ":")
		gid, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			return -1, err
		}
		idMap[int(gid)] = struct{}{}
	}
	for i := 1000; i < 65534; i++ {
		_, ok := idMap[i]
		if !ok {
			return i, nil
		}
	}
	return -1, fmt.Errorf("out of uid/gids")
}

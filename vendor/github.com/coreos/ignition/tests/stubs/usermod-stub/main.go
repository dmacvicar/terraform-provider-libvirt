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
	flagRoot        string
	flagHome        string
	flagMoveHome    bool
	flagNoUserGroup bool
	flagPassword    string
	flagUid         int
	flagComment     string
	flagGid         int
	flagGroups      string
	flagShell       string
)

func main() {
	flag.StringVar(&flagRoot, "root", "/", "Apply changes in the CHROOT_DIR directory and use the configuration files from the CHROOT_DIR directory")
	flag.StringVar(&flagHome, "home-dir", "", "The new user will be created using HOME_DIR as the value for the user's login directory")
	flag.BoolVar(&flagMoveHome, "move-home", false, "Move the content of the user's home directory to the new location")
	flag.BoolVar(&flagNoUserGroup, "no-user-group", false, "Do not create a group with the same name as the user")
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

	passwdContents, err := ioutil.ReadFile(path.Join(flagRoot, "/etc/passwd"))
	if err != nil {
		fmt.Printf("couldn't open /etc/passwd: %v\n", err)
		os.Exit(1)
	}
	passwdLines := strings.Split(string(passwdContents), "\n")
	for i, l := range passwdLines {
		if i == len(passwdLines)-1 {
			// The last line is empty
			break
		}
		tokens := strings.Split(l, ":")
		if len(tokens) != 7 {
			fmt.Printf("scanned incorrect number of items: %d\n", len(tokens))
			os.Exit(1)
		}
		currUser := tokens[0]
		currUid := tokens[2]
		currGid := tokens[3]
		currComment := tokens[4]
		currHome := tokens[5]
		currShell := tokens[6]

		if currUser != username {
			continue
		}

		if flagUid != -1 {
			currUid = strconv.Itoa(flagUid)
		}
		if flagGid != -1 {
			currGid = strconv.Itoa(flagGid)
		}
		if flagComment != "" {
			currComment = flagComment
		}
		if flagHome != "" {
			currHome = flagHome
		}
		if flagShell != "" {
			currShell = flagShell
		}

		newPasswdLine := fmt.Sprintf("%s:x:%s:%s:%s:%s:%s", currUser, currUid, currGid, currComment, currHome, currShell)

		passwdLines[i] = newPasswdLine
	}

	passwdFile, err := os.OpenFile(path.Join(flagRoot, "/etc/passwd"), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("couldn't open passwd file: %v\n", err)
		os.Exit(1)
	}
	defer passwdFile.Close()
	_, err = passwdFile.Write([]byte(strings.Join(passwdLines, "\n")))
	if err != nil {
		fmt.Printf("couldn't write to passwd file: %v\n", err)
		os.Exit(1)
	}
}

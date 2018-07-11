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
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("id called incorrectly\n")
		os.Exit(1)
	}
	fmt.Printf("id called for user %s\n", os.Args[1])

	// id accepts both usernames and UIDs, so attempt a lookup for both. If
	// either lookup doesn't return an error, exit cleanly.

	passwdContents, err := ioutil.ReadFile("/etc/passwd")
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
		if currUser == os.Args[1] {
			os.Exit(0)
		}
		if currUid == os.Args[1] {
			os.Exit(0)
		}
	}

	os.Exit(1)
}

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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"

	keys "github.com/coreos/update-ssh-keys/authorized_keys_d"
)

var (
	flagUser     = flag.String("u", "core", "Update the given user's authorized_keys file.")
	flagAdd      = flag.String("a", "", "Add the given keys, using the given name to identify them.")
	flagAddForce = flag.String("A", "", "Add the given keys, even if it was disabled with '-D'.")
	flagDelete   = flag.String("d", "", "Delete keys identified by the given name.")
	flagDisable  = flag.String("D", "", "Disable the given set from being added with '-a'.")
	flagList     = flag.Bool("l", false, "List the names and number of keys currently installed.")
	flagReplace  = flag.Bool("n", true, "When adding, don't replace an existing key with the given name.")
	flagHelp     = flag.Bool("h", false, "This ;-)")
)

func stderr(f string, a ...interface{}) {
	out := fmt.Sprintf(f, a...)
	fmt.Fprintln(os.Stderr, strings.TrimSuffix(out, "\n"))
}

func stdout(f string, a ...interface{}) {
	out := fmt.Sprintf(f, a...)
	fmt.Fprintln(os.Stdout, strings.TrimSuffix(out, "\n"))
}

func panicf(f string, a ...interface{}) {
	panic(fmt.Sprintf(f, a...))
}

// printKeys prints all the keys currently managed.
func printKeys(akd *keys.SSHAuthorizedKeysDir) error {
	stdout("All keys for %s", *flagUser)
	return akd.WalkKeys(func(k *keys.SSHAuthorizedKey) error {
		if !k.Disabled {
			cmd := exec.Command("ssh-keygen", "-l", "-f", k.Path)
			out, err := cmd.CombinedOutput()
			if err != nil {
				return err
			}
			stdout("%s: %s", k.Name, string(out))
		}
		return nil
	})
}

// addKeys adds keys to akd as name.
func addKeys(akd *keys.SSHAuthorizedKeysDir, name string, force bool) error {
	k := []byte{}
	files := []*os.File{}

	if flag.NArg() > 0 {
		// read from files
		for _, fp := range flag.Args() {
			f, err := os.Open(fp)
			if err != nil {
				return err
			}
			defer f.Close()
			files = append(files, f)
		}
	} else {
		// read from stdin
		files = append(files, os.Stdin)
	}

	for _, f := range files {
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		k = append(k, b...)
	}

	if !validKeys(k) {
		return fmt.Errorf("key(s) invalid.")
	}

	return akd.Add(name, k, *flagReplace, force)
}

// validKeys validates a byte slice contains valid ssh keys.
func validKeys(keys []byte) bool {
	// we need to write out a temp file for ssh-keygen to consume.
	tf, err := ioutil.TempFile("", "update-ssh-keys-")
	if err != nil {
		panicf("unable to create temporary file: %v", err)
	}
	defer func() {
		os.Remove(tf.Name())
		tf.Close()
	}()

	if _, err := tf.Write(keys); err != nil {
		panicf("unable to write temporary file: %v", err)
	}

	cmd := exec.Command("ssh-keygen", "-l", "-f", tf.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	stdout("%s", out)

	return true
}

func main() {
	flag.Parse()

	if *flagHelp {
		flag.Usage()
		os.Exit(1)
	}

	usr, err := user.Lookup(*flagUser)
	if err != nil {
		stderr("Failed to lookup user: %v", *flagUser)
		os.Exit(2)
	}

	akd, err := keys.Open(usr, true)
	if err != nil {
		stderr("Failed to open: %v", err)
		os.Exit(3)
	}
	defer akd.Close()

	switch {
	case *flagList:
		if err := printKeys(akd); err != nil {
			stderr("Failed to print keys: %v", err)
			os.Exit(4)
		}
	case *flagAdd != "":
		if err := addKeys(akd, *flagAdd, false); err != nil {
			stderr("failed to add keys %q: %v", *flagAdd, err)
			os.Exit(5)
		}
	case *flagAddForce != "":
		if err := addKeys(akd, *flagAddForce, true); err != nil {
			stderr("failed to add keys %q: %v", *flagAddForce, err)
			os.Exit(6)
		}
	case *flagDelete != "":
		if err := akd.Remove(*flagDelete); err != nil {
			stderr("failed to delete key %q: %v", *flagDelete, err)
			os.Exit(7)
		}
	case *flagDisable != "":
		if err := akd.Disable(*flagDisable); err != nil {
			stderr("failed to disable key %q: %v", *flagDisable, err)
			os.Exit(8)
		}
	}

	if err := akd.Sync(); err != nil {
		stderr("Failed sync: %v", err)
		os.Exit(9)
	}
	stdout("Updated %s", akd.KeysFilePath())

	os.Exit(0)
}

package libvirt

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

type GuestfsStatus int

const (
	Stopped GuestfsStatus = iota
	Added
	Started
	Mounted
)

var currentGuestfishStatus = Stopped

func (s GuestfsStatus) String() string {
	return []string{"Stopped", "Added", "Started", "Mounted"}[s]
}

/*
var supportedCommands = [][]string{
	{
		"guestfish", "--listen", "-a", "*",
	},
	{
		"guestfish", "--remote", "--", "run",
	},
	{
		"guestfish", "--remote", "--", "findfs-label", "boot",
	},
	{
		"guestfish", "--remote", "--", "mount", "*", "/",
	},
	{
		"guestfish", "--remote", "--", "mkdir-p", "/ignition",
	},
	{
		"guestfish", "--remote", "--", "upload", "*.ign", "/ignition/config.ign",
	},
	{
		"guestfish", "--remote", "--", "umount-all",
	},
	{
		"guestfish", "--remote", "--", "exit",
	},
} */

var supportedCommandTree = map[string]interface{}{
	"guestfish": map[string]interface{}{
		"--listen": map[string]interface{}{
			"-a": map[string]interface{}{
				".*": nil,
			},
		},
		"--remote": map[string]interface{}{
			"--": map[string]interface{}{
				"run": nil,
				"findfs-label": map[string]interface{}{
					"boot": nil,
				},
				"mount": map[string]interface{}{
					".*": map[string]interface{}{
						"/": nil,
					},
				},
				"mkdir-p": map[string]interface{}{
					"/ignition": nil,
				},
				"upload": map[string]interface{}{
					".*.ign": map[string]interface{}{
						"/ignition/config.ign": nil,
					},
				},
				"umount-all": nil,
				"exit":       nil,
			},
		},
	},
}

// fakeExecCommand is used to replace the real command in unit test
func fakeExecCommand(executable string, args ...string) *exec.Cmd {
	newArgs := []string{"-test.run=TestHelperCommand", "--", executable}
	newArgs = append(newArgs, args...)

	cmd := exec.Command(os.Args[0], newArgs...)
	cmd.Env = []string{"USE_HELPER_COMMAND=1"}

	cmdToValidate := []string{executable}
	cmdToValidate = append(cmdToValidate, args...)

	fmt.Printf("### Executing command: %s\n", strings.Join(cmd.Args, " "))

	if valid := validateCommandSyntax(cmdToValidate...); !valid {
		cmd.Env = append(cmd.Env, fmt.Sprintf("STDERR=%s", fmt.Sprintf("invalid command: %s", strings.Join(cmdToValidate, " "))))
		cmd.Env = append(cmd.Env, "EXIT_STATUS=1")
	} else {
		if valid, msg := validateCommandSemanticsAndGenerateOutput(cmdToValidate...); valid {
			cmd.Env = append(cmd.Env, fmt.Sprintf("STDOUT=%s", msg))
			cmd.Env = append(cmd.Env, "EXIT_STATUS=0")
		} else {
			cmd.Env = append(cmd.Env, fmt.Sprintf("STDERR=%s", msg))
			cmd.Env = append(cmd.Env, "EXIT_STATUS=2")
		}
	}
	return cmd
}

// validateCommand checks whether the command's syntax is valid or not
func validateCommandSyntax(args ...string) bool {
	if len(args) < 4 {
		return false
	}

	newArgs := args
	if newArgs[0] == "sudo" && newArgs[1] == "--preserve-env" {
		newArgs = newArgs[2:]
	}

	cmdTree := supportedCommandTree
	// check each argument one by one
	for index, arg := range newArgs {
		if v, ok := cmdTree[arg]; ok {
			if index+1 == len(newArgs) {
				return v == nil
			}
			cmdTree = v.(map[string]interface{})
		} else {
			matched := false
			for k, v := range cmdTree {
				if matched, _ = regexp.MatchString(k, arg); matched {
					if index+1 == len(newArgs) {
						return v == nil
					}
					cmdTree = v.(map[string]interface{})
					break
				}
			}
			if !matched {
				return false
			}
		}
	}

	return true
}

// validateCommandSemantic checks whether the command's semantics is valid or not.
// if return true, then the retunred string is the specific output message for the command;
// if return false, then the returned string is an error message.
func validateCommandSemanticsAndGenerateOutput(args ...string) (bool, string) {
	newArgs := args
	if newArgs[0] == "sudo" && newArgs[1] == "--preserve-env" {
		newArgs = newArgs[2:]
	}

	errMsg := fmt.Sprintf("Can't execute command '%v' when the guestfish status is %s", strings.Join(args, " "), currentGuestfishStatus)

	if newArgs[1] == "--listen" {
		if Stopped == currentGuestfishStatus {
			currentGuestfishStatus = Added
			return true, "GUESTFISH_PID=4513; export GUESTFISH_PID"
		}
		return false, errMsg
	}

	if newArgs[3] == "run" {
		if Added == currentGuestfishStatus {
			currentGuestfishStatus = Started
			return true, ""
		}
		return false, errMsg
	}

	if newArgs[3] == "findfs-label" {
		if Started == currentGuestfishStatus {
			return true, "/dev/sda1"
		}
		return false, errMsg
	}

	if newArgs[3] == "mount" {
		if Started == currentGuestfishStatus {
			currentGuestfishStatus = Mounted
			return true, ""
		}
		return false, errMsg
	}

	if newArgs[3] == "mkdir-p" {
		if Mounted == currentGuestfishStatus {
			return true, ""
		}
		return false, errMsg
	}

	if newArgs[3] == "upload" {
		if Mounted == currentGuestfishStatus {
			return true, ""
		}
		return false, errMsg
	}

	if newArgs[3] == "umount-all" {
		if Mounted == currentGuestfishStatus {
			currentGuestfishStatus = Started
			return true, ""
		}
		return false, errMsg
	}

	if newArgs[3] == "exit" {
		if Stopped < currentGuestfishStatus {
			return true, ""
		}

		return false, errMsg
	}
	return true, ""
}

func TestHelperCommand(t *testing.T) {
	if os.Getenv("USE_HELPER_COMMAND") != "1" {
		return
	}

	stdOutMsg := os.Getenv("STDOUT")
	if len(stdOutMsg) > 0 {
		fmt.Fprintf(os.Stdout, stdOutMsg)
	}

	stdErrMsg := os.Getenv("STDERR")
	if len(stdErrMsg) > 0 {
		fmt.Fprintf(os.Stderr, stdErrMsg)
	}

	exitCode, _ := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	os.Exit(exitCode)
}

func TestGuestfishExecutionLocal(t *testing.T) {

	execCommand = fakeExecCommand
	defer func() {
		execCommand = exec.Command
	}()

	err := guestfishExecutionLocal("/root/path", "my.ign")

	if err != nil {
		t.Errorf("failed to call guestfishExecutionLocal, error: %v", err)
	}
}

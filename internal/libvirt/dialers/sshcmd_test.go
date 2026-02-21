package dialers

import (
	"net/url"
	"testing"
)

func TestSSHCmdBuildSSHArgsDoesNotForceControlPath(t *testing.T) {
	t.Parallel()

	uri, err := url.Parse("qemu+sshcmd://user@example.com/system")
	if err != nil {
		t.Fatalf("failed to parse uri: %v", err)
	}

	dialer := NewSSHCmd(uri)
	args := dialer.buildSSHArgs()

	for i := range args {
		if args[i] == "ControlPath=none" {
			t.Fatalf("unexpected forced ControlPath option in args: %v", args)
		}
	}
}

func TestSSHCmdBuildSSHArgsIncludesExpectedCoreFlags(t *testing.T) {
	t.Parallel()

	uri, err := url.Parse("qemu+sshcmd://user@example.com/system")
	if err != nil {
		t.Fatalf("failed to parse uri: %v", err)
	}

	dialer := NewSSHCmd(uri)
	args := dialer.buildSSHArgs()

	assertContainsArg(t, args, "-T")
	assertContainsPair(t, args, "-e", "none")
}

func assertContainsArg(t *testing.T, args []string, arg string) {
	t.Helper()

	for i := range args {
		if args[i] == arg {
			return
		}
	}

	t.Fatalf("expected arg %q in args: %v", arg, args)
}

func assertContainsPair(t *testing.T, args []string, key string, value string) {
	t.Helper()

	for i := 0; i < len(args)-1; i++ {
		if args[i] == key && args[i+1] == value {
			return
		}
	}

	t.Fatalf("expected pair %q %q in args: %v", key, value, args)
}

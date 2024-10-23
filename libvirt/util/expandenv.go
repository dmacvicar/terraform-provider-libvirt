package util

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	userHomeDir = os.UserHomeDir
	expandEnv   = os.ExpandEnv
)

// ExpandEnvExt expands environment variables and resolves ~ to the home directory
// this is a drop-in replacement for os.ExpandEnv but is additionally '~' aware.
func ExpandEnvExt(path string) string {
	path = filepath.Clean(expandEnv(path))
	tilde := filepath.FromSlash("~/")

	// note to maintainers: tilde without a following slash character is simply
	// interpreted as part of the filename (e.g. ~foo/bar != ~/foo/bar). However,
	// when running on windows, the filepath will be represented by backslashes ('\'),
	// therefore we need to convert "~/" to the platform specific format to test for
	// it, otherwise on windows systems the prefix test will always fail.
	if strings.HasPrefix(path, tilde) {
		home, err := userHomeDir()
		if err != nil {
			return path // return path as-is if unable to resolve home directory
		}
		// Replace ~ with home directory
		path = filepath.Join(home, strings.TrimPrefix(path, tilde))
	}
	return path
}

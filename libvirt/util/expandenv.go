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
// this is a drop-in replacement for os.ExpandEnv but is additionally '~' aware
func ExpandEnvExt(path string) string {
	path = expandEnv(path)
	if strings.HasPrefix(path, "~/") {
		home, err := userHomeDir()
		if err != nil {
			return path // return path as-is if unable to resolve home directory
		}
		// Replace ~ with home directory
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path
}

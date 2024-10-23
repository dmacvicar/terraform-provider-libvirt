package util

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandEnvExt(t *testing.T) {
	userHomeDir = func() (string, error) {
		return "/home/mock", nil
	}
	expandEnv = func(s string) string {
		return strings.Replace(s, "${HOME}", "/home/mock", 1)
	}


	assert.Equal(t, filepath.FromSlash("foo/bar/baz"), ExpandEnvExt("foo/bar/baz"))
	assert.Equal(t, filepath.FromSlash("/home/mock/foo/bar/baz"), ExpandEnvExt("~/foo/bar/baz"))
	assert.Equal(t, filepath.FromSlash("/home/mock/foo/bar/baz"), ExpandEnvExt("${HOME}/foo/bar/baz"))
	assert.Equal(t, filepath.FromSlash("~foo/bar/baz"), ExpandEnvExt("~foo/bar/baz"))

	userHomeDir = func() (string, error) {
		return "", fmt.Errorf("some failure")
	}

	// failure to get home expansion should leave string unchanged
	assert.Equal(t, filepath.FromSlash("~/foo/bar/baz"), ExpandEnvExt("~/foo/bar/baz"))
}

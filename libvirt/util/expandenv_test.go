package util

import (
	"testing"
	"strings"

	"github.com/stretchr/testify/assert"
)

func TestExpandEnvExt(t *testing.T) {
	userHomeDir = func() (string, error) {
		return "/home/mock", nil
	}
	expandEnv = func(s string) string {
		return strings.Replace(s, "${HOME}", "/home/mock", 1)
	}


	assert.Equal(t, "foo/bar/baz", ExpandEnvExt("foo/bar/baz"))
	assert.Equal(t, "/home/mock/foo/bar/baz", ExpandEnvExt("~/foo/bar/baz"))
	assert.Equal(t, "/home/mock/foo/bar/baz", ExpandEnvExt("${HOME}/foo/bar/baz"))
	assert.Equal(t, "~foo/bar/baz", ExpandEnvExt("~foo/bar/baz"))
}

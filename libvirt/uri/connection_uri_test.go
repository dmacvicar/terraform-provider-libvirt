package uri

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestURI(t *testing.T) {
	fixtures := []struct {
		URI        string
		Driver     string
		Transport  string
		RemoteName string
	}{
		{"xxx://servername/", "xxx", "tls", "xxx:///"},
		{"xxx+tls://servername/", "xxx", "tls", "xxx:///"},
		{"xxx+tls:///", "xxx", "tls", "xxx:///"},
		{"xxx+tcp://servername/", "xxx", "tcp", "xxx:///"},
		{"xxx+tcp:///", "xxx", "tcp", "xxx:///"},
		{"xxx+unix:///", "xxx", "unix", "xxx:///"},
		{"xxx+tls://servername/?foo=bar&name=dong:///ding", "xxx", "tls", "dong:///ding"},
		{"xxx+ssh://username@hostname:2222/path?foo=bar&bar=foo", "xxx", "ssh", "xxx:///path"},
	}

	for _, fixture := range fixtures {
		u, err := Parse(fixture.URI)
		assert.NoError(t, err)
		assert.Equal(t, fixture.Transport, u.transport())
		assert.Equal(t, fixture.Driver, u.driver())
		assert.Equal(t, fixture.RemoteName, u.RemoteName())
	}
}

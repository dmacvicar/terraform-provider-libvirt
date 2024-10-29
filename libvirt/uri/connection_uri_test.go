package uri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURI(t *testing.T) {
	fixtures := []struct {
		URI        string
		Driver     string
		Transport  string
		RemoteName string
		HostName   string
		Port       string
	}{
		{"xxx://servername/", "xxx", "tls", "xxx:///", "servername", ""},
		{"xxx+tls://servername/", "xxx", "tls", "xxx:///", "servername", ""},
		{"xxx+tls:///", "xxx", "tls", "xxx:///", "", ""},
		{"xxx+tcp://servername/", "xxx", "tcp", "xxx:///", "servername", ""},
		{"xxx+tcp:///", "xxx", "tcp", "xxx:///", "", ""},
		{"xxx+unix:///", "xxx", "unix", "xxx:///", "", ""},
		{"xxx+tls://servername/?foo=bar&name=dong:///ding", "xxx", "tls", "dong:///ding", "servername", ""},
		{"xxx+ssh://username@hostname:2222/path?foo=bar&bar=foo", "xxx", "ssh", "xxx:///path", "hostname", "2222"},
	}

	for _, fixture := range fixtures {
		u, err := Parse(fixture.URI)
		assert.NoError(t, err)
		assert.Equal(t, fixture.Transport, u.transport())
		assert.Equal(t, fixture.Driver, u.driver())
		assert.Equal(t, fixture.RemoteName, u.RemoteName())
		assert.Equal(t, fixture.HostName, u.Host)
		assert.Equal(t, fixture.Port, u.Port())
	}
}

package uri

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyJumpStringToParsedTarget(t *testing.T) {
	in := []string{
		"host.enterprise.com",
		"user@host.enterprise.com",
	}
	expectedOut := []parsedTarget{
		{
			hostName: "host.enterprise.com",
		},
		{
			hostName: "host.enterprise.com",
			user:     "user",
		},
	}

	out := []parsedTarget{}
	for _, proxyJumpStr := range in {
		out = append(out, proxyJumpStringToParsedTarget(proxyJumpStr))
	}
	assert.Equal(t, expectedOut, out)
}

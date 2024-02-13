package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBoolYesNo(t *testing.T) {
	assert.Equal(t, "yes", FormatBoolYesNo(true))
	assert.Equal(t, "no", FormatBoolYesNo(false))
}

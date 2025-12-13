package testutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertEqualWithMessage(t *testing.T, name string, expected, actual any) {
	assert.Equal(t, expected, actual, "Expected %s to be %d, but got %d", name, expected, actual)
}

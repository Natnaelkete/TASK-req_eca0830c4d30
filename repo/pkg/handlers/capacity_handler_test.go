package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapacityHandler(t *testing.T) {
	h := NewCapacityHandler(nil)
	assert.NotNil(t, h)
}

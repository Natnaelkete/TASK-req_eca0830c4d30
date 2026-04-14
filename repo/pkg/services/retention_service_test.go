package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRetentionService(t *testing.T) {
	svc := NewRetentionService(nil)
	assert.NotNil(t, svc)
}

func TestHotRetentionDays(t *testing.T) {
	assert.Equal(t, 90, HotRetentionDays, "hot retention threshold must be 90 days")
}

func TestColdRetentionYears(t *testing.T) {
	assert.Equal(t, 3, ColdRetentionYears, "cold retention threshold must be 3 years")
}

func TestContainsStr(t *testing.T) {
	assert.True(t, containsStr("partition already exists", "already exists"))
	assert.True(t, containsStr("Error: already exists in table", "already exists"))
	assert.False(t, containsStr("some other error", "already exists"))
	assert.False(t, containsStr("", "already exists"))
}

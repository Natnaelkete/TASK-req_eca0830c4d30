package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCapacityService(t *testing.T) {
	svc := NewCapacityService(nil)
	assert.NotNil(t, svc)
}

func TestDiskUsageThreshold(t *testing.T) {
	assert.Equal(t, 80.0, DiskUsageThreshold)
}

func TestCapacityCheckInterval(t *testing.T) {
	assert.Equal(t, 60*time.Second, CapacityCheckInterval)
}

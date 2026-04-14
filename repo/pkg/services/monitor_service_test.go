package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMonitorService(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()
	svc := NewMonitorService(nil, q)
	assert.NotNil(t, svc)
}

func TestMonitorDeviceInput_Fields(t *testing.T) {
	in := MonitorDeviceInput{DeviceID: 5}
	assert.Equal(t, uint(5), in.DeviceID)
}

func TestThresholdCheckInput_Fields(t *testing.T) {
	in := ThresholdCheckInput{
		DeviceID:   1,
		MetricType: "temperature",
		Threshold:  30.0,
		Level:      "critical",
	}
	assert.Equal(t, "temperature", in.MetricType)
	assert.Equal(t, 30.0, in.Threshold)
	assert.Equal(t, "critical", in.Level)
}

func TestAlertListParams_Defaults(t *testing.T) {
	p := AlertListParams{}
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, 0, p.PageSize)
	assert.Nil(t, p.Resolved)
}

func TestErrAlertNotFound(t *testing.T) {
	assert.EqualError(t, ErrAlertNotFound, "alert not found")
}

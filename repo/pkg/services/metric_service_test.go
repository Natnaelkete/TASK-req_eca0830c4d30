package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricService(t *testing.T) {
	svc := NewMetricService(nil)
	assert.NotNil(t, svc)
}

func TestCreateMetricInput_Fields(t *testing.T) {
	in := CreateMetricInput{
		DeviceID:   1,
		MetricType: "temperature",
		Value:      23.5,
		Unit:       "celsius",
		EventTime:  "2026-01-15T10:00:00Z",
	}
	assert.Equal(t, uint(1), in.DeviceID)
	assert.Equal(t, "temperature", in.MetricType)
	assert.Equal(t, 23.5, in.Value)
}

func TestBatchMetricInput_Fields(t *testing.T) {
	in := BatchMetricInput{
		Metrics: []CreateMetricInput{
			{DeviceID: 1, MetricType: "temp", Value: 20, EventTime: "2026-01-15T10:00:00Z"},
			{DeviceID: 1, MetricType: "humidity", Value: 60, EventTime: "2026-01-15T10:00:00Z"},
		},
	}
	assert.Len(t, in.Metrics, 2)
}

func TestMetricListParams_Defaults(t *testing.T) {
	p := MetricListParams{}
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, "", p.MetricType)
	assert.Equal(t, "", p.StartTime)
	assert.Equal(t, "", p.EndTime)
}

func TestErrMetricNotFound(t *testing.T) {
	assert.EqualError(t, ErrMetricNotFound, "metric not found")
}

package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMonitoringDataService(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()
	svc := NewMonitoringDataService(nil, q)
	assert.NotNil(t, svc)
}

func TestMonitoringDataInput_Fields(t *testing.T) {
	in := MonitoringDataInput{
		SourceID:   "sensor-001",
		DeviceID:   1,
		PlotID:     2,
		MetricCode: "temperature",
		Value:      23.5,
		Unit:       "celsius",
		EventTime:  "2026-01-15T10:00:00Z",
		Tags:       `{"location":"field-A"}`,
	}
	assert.Equal(t, "sensor-001", in.SourceID)
	assert.Equal(t, uint(1), in.DeviceID)
	assert.Equal(t, uint(2), in.PlotID)
	assert.Equal(t, "temperature", in.MetricCode)
	assert.Equal(t, 23.5, in.Value)
	assert.Equal(t, "celsius", in.Unit)
}

func TestBatchMonitoringDataInput_Validation(t *testing.T) {
	in := BatchMonitoringDataInput{
		Data: []MonitoringDataInput{
			{SourceID: "s1", DeviceID: 1, PlotID: 1, MetricCode: "temp", Value: 20, EventTime: "2026-01-15T10:00:00Z"},
			{SourceID: "s2", DeviceID: 1, PlotID: 1, MetricCode: "humidity", Value: 60, EventTime: "2026-01-15T10:00:00Z"},
		},
	}
	assert.Len(t, in.Data, 2)
}

func TestAggregationParams_Fields(t *testing.T) {
	p := AggregationParams{
		PlotID:     1,
		DeviceID:   2,
		MetricCode: "temperature",
		StartTime:  "2026-01-01T00:00:00Z",
		EndTime:    "2026-01-31T23:59:59Z",
		Function:   "avg",
		GroupBy:    "device_id",
		UserID:     10,
		Role:       "researcher",
	}
	assert.Equal(t, "avg", p.Function)
	assert.Equal(t, "device_id", p.GroupBy)
	assert.Equal(t, uint(10), p.UserID)
	assert.Equal(t, "researcher", p.Role)
}

func TestCurveParams_Defaults(t *testing.T) {
	p := CurveParams{}
	assert.Equal(t, 0, p.Minutes) // Default handled in service
	assert.Equal(t, "", p.MetricCode)
}

func TestCurveParams_IsolationFields(t *testing.T) {
	p := CurveParams{
		PlotID:     1,
		MetricCode: "temperature",
		UserID:     42,
		Role:       "researcher",
	}
	assert.Equal(t, uint(42), p.UserID)
	assert.Equal(t, "researcher", p.Role)
}

func TestTrendParams_Fields(t *testing.T) {
	p := TrendParams{
		MetricCode: "temperature",
		StartTime:  "2026-01-01T00:00:00Z",
		EndTime:    "2026-12-31T23:59:59Z",
		Interval:   "monthly",
		Function:   "avg",
		UserID:     10,
		Role:       "researcher",
	}
	assert.Equal(t, "monthly", p.Interval)
	assert.Equal(t, "avg", p.Function)
	assert.Equal(t, uint(10), p.UserID)
	assert.Equal(t, "researcher", p.Role)
}

func TestMonitoringDataListParams_Defaults(t *testing.T) {
	p := MonitoringDataListParams{}
	assert.Equal(t, 0, p.Page)
	assert.Equal(t, "", p.MetricCode)
}

func TestExportParams_Fields(t *testing.T) {
	p := ExportParams{
		PlotID:     1,
		DeviceID:   2,
		MetricCode: "humidity",
		StartTime:  "2026-01-01T00:00:00Z",
		EndTime:    "2026-01-31T23:59:59Z",
		UserID:     42,
		Role:       "researcher",
	}
	assert.Equal(t, uint(1), p.PlotID)
	assert.Equal(t, "humidity", p.MetricCode)
	assert.Equal(t, uint(42), p.UserID)
	assert.Equal(t, "researcher", p.Role)
}

func TestErrMonitoringDataNotFound(t *testing.T) {
	assert.EqualError(t, ErrMonitoringDataNotFound, "monitoring data not found")
}

func TestErrMonitoringDataForbidden(t *testing.T) {
	assert.EqualError(t, ErrMonitoringDataForbidden, "not authorized to access this monitoring data")
}

func TestMonitoringDataListParams_IsolationFields(t *testing.T) {
	p := MonitoringDataListParams{
		Page:     1,
		PageSize: 20,
		UserID:   42,
		Role:     "researcher",
	}
	assert.Equal(t, uint(42), p.UserID)
	assert.Equal(t, "researcher", p.Role)
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		msg    string
		expect bool
	}{
		{"Duplicate entry '123' for key 'idx_monitoring_idempotent'", true},
		{"Error 1062: Duplicate entry", true},
		{"UNIQUE constraint failed: monitoring_data.source_id", true},
		{"some other error", false},
	}
	// assert.AnError won't match any duplicate patterns
	assert.False(t, isDuplicateKeyError(assert.AnError))

	for _, tt := range tests {
		err := &testError{msg: tt.msg}
		assert.Equal(t, tt.expect, isDuplicateKeyError(err), "msg=%s", tt.msg)
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestAggSQLExpr(t *testing.T) {
	tests := []struct {
		fn     string
		expect string
	}{
		{"sum", "SUM(value)"},
		{"min", "MIN(value)"},
		{"max", "MAX(value)"},
		{"count", "COUNT(*)"},
		{"avg", "AVG(value)"},
		{"unknown", "AVG(value)"}, // default
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, aggSQLExpr(tt.fn))
	}
}

func TestSubmitBatchIngest_EnqueuesJob(t *testing.T) {
	// Use 0 workers so the job stays in the queue (no nil-db panic from worker)
	q := NewQueueService(100, 0)
	svc := NewMonitoringDataService(nil, q)

	inputs := []MonitoringDataInput{
		{SourceID: "s1", DeviceID: 1, PlotID: 1, MetricCode: "temp", Value: 20, EventTime: "2026-01-15T10:00:00Z"},
	}

	job, err := svc.SubmitBatchIngest(inputs)
	require.NoError(t, err)
	assert.NotEmpty(t, job.ID)
	assert.Equal(t, JobPending, job.Status)
	assert.Equal(t, "batch_ingest", job.Type)

	// Verify payload has serialized data
	dataStr, ok := job.Payload["data"].(string)
	require.True(t, ok)

	var parsed []MonitoringDataInput
	err = json.Unmarshal([]byte(dataStr), &parsed)
	require.NoError(t, err)
	assert.Len(t, parsed, 1)
	assert.Equal(t, "s1", parsed[0].SourceID)
}

func TestHandleBatchIngest_BadPayload(t *testing.T) {
	q := NewQueueService(10, 1)
	defer q.Shutdown()
	svc := NewMonitoringDataService(nil, q)

	// Test with missing data field
	job := &Job{Payload: map[string]interface{}{}}
	_, err := svc.handleBatchIngest(context.Background(), job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing data")
}

func TestHandleBatchIngest_InvalidPayload(t *testing.T) {
	q := NewQueueService(100, 0)
	svc := NewMonitoringDataService(nil, q)

	// Test invalid JSON in payload
	job := &Job{Payload: map[string]interface{}{"data": "not-valid-json"}}
	_, err := svc.handleBatchIngest(context.Background(), job)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestTrendResult_Structure(t *testing.T) {
	result := TrendResult{
		Current: []TrendPoint{
			{Period: "2026-01", Value: 23.5, Count: 100},
			{Period: "2026-02", Value: 24.1, Count: 120},
		},
		Previous: []TrendPoint{
			{Period: "2025-01", Value: 22.0, Count: 90},
		},
	}
	assert.Len(t, result.Current, 2)
	assert.Len(t, result.Previous, 1)
	assert.Equal(t, "2026-01", result.Current[0].Period)
}

func TestCurvePoint_Structure(t *testing.T) {
	now := time.Now()
	point := CurvePoint{
		EventTime: now,
		Value:     42.5,
	}
	assert.Equal(t, now, point.EventTime)
	assert.Equal(t, 42.5, point.Value)
}

func TestAggregationResult_Structure(t *testing.T) {
	result := AggregationResult{
		GroupKey: "temperature",
		Value:    25.3,
		Count:    50,
	}
	assert.Equal(t, "temperature", result.GroupKey)
	assert.Equal(t, 25.3, result.Value)
	assert.Equal(t, int64(50), result.Count)
}

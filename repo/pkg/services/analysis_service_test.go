package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAnalysisService(t *testing.T) {
	svc := NewAnalysisService(nil)
	assert.NotNil(t, svc)
}

// --- TrendAnalysisParams ---

func TestTrendAnalysisParams_Fields(t *testing.T) {
	p := TrendAnalysisParams{
		MetricCode: "temperature",
		StartTime:  "2026-01-01T00:00:00Z",
		EndTime:    "2026-12-31T23:59:59Z",
		Interval:   "monthly",
		Function:   "avg",
		PlotID:     1,
		DeviceID:   5,
		DrillBy:    "plot_id",
	}
	assert.Equal(t, "temperature", p.MetricCode)
	assert.Equal(t, "monthly", p.Interval)
	assert.Equal(t, "avg", p.Function)
	assert.Equal(t, "plot_id", p.DrillBy)
}

func TestTrendBucket_Structure(t *testing.T) {
	b := TrendBucket{
		Period:   "2026-01",
		Value:    23.5,
		Count:    100,
		GroupKey: "3",
	}
	assert.Equal(t, "2026-01", b.Period)
	assert.Equal(t, 23.5, b.Value)
	assert.Equal(t, int64(100), b.Count)
	assert.Equal(t, "3", b.GroupKey)
}

func TestTrendAnalysisResult_Structure(t *testing.T) {
	result := TrendAnalysisResult{
		MetricCode: "humidity",
		Interval:   "daily",
		Function:   "avg",
		DrillBy:    "device_id",
		Buckets: []TrendBucket{
			{Period: "2026-01-01", Value: 60.0, Count: 50},
			{Period: "2026-01-02", Value: 62.5, Count: 48},
		},
	}
	assert.Equal(t, "humidity", result.MetricCode)
	assert.Len(t, result.Buckets, 2)
}

// --- FunnelAnalysisParams ---

func TestFunnelAnalysisParams_Fields(t *testing.T) {
	p := FunnelAnalysisParams{
		Stages:    []string{"soil_prep", "planting", "growth", "harvest"},
		StartTime: "2026-01-01T00:00:00Z",
		EndTime:   "2026-12-31T23:59:59Z",
		PlotID:    2,
		DrillBy:   "plot_id",
	}
	assert.Len(t, p.Stages, 4)
	assert.Equal(t, "soil_prep", p.Stages[0])
	assert.Equal(t, "harvest", p.Stages[3])
}

func TestFunnelStage_Structure(t *testing.T) {
	stage := FunnelStage{
		Stage:          "planting",
		Count:          80,
		ConversionRate: 80.0,
		GroupKey:       "1",
	}
	assert.Equal(t, "planting", stage.Stage)
	assert.Equal(t, int64(80), stage.Count)
	assert.Equal(t, 80.0, stage.ConversionRate)
}

func TestFunnelAnalysisResult_Structure(t *testing.T) {
	result := FunnelAnalysisResult{
		Stages:  []string{"a", "b", "c"},
		DrillBy: "",
		Steps: []FunnelStage{
			{Stage: "a", Count: 100, ConversionRate: 100},
			{Stage: "b", Count: 75, ConversionRate: 75},
			{Stage: "c", Count: 50, ConversionRate: 66.67},
		},
	}
	assert.Len(t, result.Steps, 3)
	assert.Equal(t, int64(100), result.Steps[0].Count)
	assert.Equal(t, 100.0, result.Steps[0].ConversionRate)
}

// --- RetentionAnalysisParams ---

func TestRetentionAnalysisParams_Fields(t *testing.T) {
	p := RetentionAnalysisParams{
		MetricCode:     "temperature",
		StartTime:      "2026-01-01T00:00:00Z",
		EndTime:        "2026-06-30T23:59:59Z",
		CohortInterval: "monthly",
		PlotID:         1,
		DrillBy:        "device_id",
	}
	assert.Equal(t, "temperature", p.MetricCode)
	assert.Equal(t, "monthly", p.CohortInterval)
}

func TestRetentionCohort_Structure(t *testing.T) {
	cohort := RetentionCohort{
		CohortPeriod: "2026-01",
		CohortSize:   10,
		Retention:    []float64{100, 90, 80, 70, 60, 50},
	}
	assert.Equal(t, "2026-01", cohort.CohortPeriod)
	assert.Equal(t, int64(10), cohort.CohortSize)
	assert.Len(t, cohort.Retention, 6)
	assert.Equal(t, 100.0, cohort.Retention[0])
}

func TestRetentionAnalysisResult_Structure(t *testing.T) {
	result := RetentionAnalysisResult{
		MetricCode:     "humidity",
		CohortInterval: "weekly",
		DrillBy:        "",
		Cohorts: []RetentionCohort{
			{CohortPeriod: "2026-01", CohortSize: 10, Retention: []float64{100, 80}},
			{CohortPeriod: "2026-02", CohortSize: 8, Retention: []float64{100}},
		},
	}
	assert.Len(t, result.Cohorts, 2)
	assert.Equal(t, int64(10), result.Cohorts[0].CohortSize)
}

// --- Helper functions ---

func TestNormalizeInterval(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"daily", "daily"},
		{"Daily", "daily"},
		{"weekly", "weekly"},
		{"MONTHLY", "monthly"},
		{"", "daily"},
		{"unknown", "daily"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, normalizeInterval(tt.input), "input=%s", tt.input)
	}
}

func TestNormalizeFunction(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"avg", "avg"},
		{"SUM", "sum"},
		{"min", "min"},
		{"MAX", "max"},
		{"count", "count"},
		{"", "avg"},
		{"unknown", "avg"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expect, normalizeFunction(tt.input), "input=%s", tt.input)
	}
}

func TestIntervalToDateFormat(t *testing.T) {
	assert.Equal(t, "%Y-%m-%d", intervalToDateFormat("daily"))
	assert.Equal(t, "%Y-%u", intervalToDateFormat("weekly"))
	assert.Equal(t, "%Y-%m", intervalToDateFormat("monthly"))
	assert.Equal(t, "%Y-%m-%d", intervalToDateFormat("unknown"))
}

func TestFunctionToSQL(t *testing.T) {
	assert.Equal(t, "AVG(value)", functionToSQL("avg"))
	assert.Equal(t, "SUM(value)", functionToSQL("sum"))
	assert.Equal(t, "MIN(value)", functionToSQL("min"))
	assert.Equal(t, "MAX(value)", functionToSQL("max"))
	assert.Equal(t, "COUNT(*)", functionToSQL("count"))
	assert.Equal(t, "AVG(value)", functionToSQL("unknown"))
}

func TestFindStepCount(t *testing.T) {
	steps := []FunnelStage{
		{Stage: "a", Count: 100, GroupKey: "g1"},
		{Stage: "a", Count: 80, GroupKey: "g2"},
		{Stage: "b", Count: 70, GroupKey: "g1"},
		{Stage: "b", Count: 50, GroupKey: "g2"},
	}

	// stageIdx=0 for "g1" should return 100
	assert.Equal(t, int64(100), findStepCount(steps, 0, "g1"))
	// stageIdx=0 for "g2" should return 80
	assert.Equal(t, int64(80), findStepCount(steps, 0, "g2"))
	// stageIdx=1 for "g1" should return 70
	assert.Equal(t, int64(70), findStepCount(steps, 1, "g1"))
	// stageIdx=1 for "g2" should return 50
	assert.Equal(t, int64(50), findStepCount(steps, 1, "g2"))
	// Non-existent group
	assert.Equal(t, int64(0), findStepCount(steps, 0, "g3"))
	// Out-of-range stage index
	assert.Equal(t, int64(0), findStepCount(steps, 5, "g1"))
}

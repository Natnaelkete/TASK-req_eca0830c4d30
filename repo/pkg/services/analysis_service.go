package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

// AnalysisService provides trend, funnel, and retention analysis over monitoring data
// with multi-dimensional drill-down capability.
type AnalysisService struct {
	db *gorm.DB
}

func NewAnalysisService(db *gorm.DB) *AnalysisService {
	return &AnalysisService{db: db}
}

// ============================================================
// Trend Analysis
// ============================================================

// TrendAnalysisParams defines a trend query with drill-down.
type TrendAnalysisParams struct {
	MetricCode string `json:"metric_code" binding:"required"`
	StartTime  string `json:"start_time"  binding:"required"` // RFC3339
	EndTime    string `json:"end_time"    binding:"required"` // RFC3339
	Interval   string `json:"interval"`                       // daily, weekly, monthly (default daily)
	Function   string `json:"function"`                       // avg, sum, min, max, count (default avg)
	PlotID     uint   `json:"plot_id"`                        // drill-down: filter by plot
	DeviceID   uint   `json:"device_id"`                      // drill-down: filter by device
	DrillBy    string `json:"drill_by"`                       // drill-down grouping: plot_id, device_id, or empty for aggregate
}

// TrendBucket is a single time-bucket in a trend result.
type TrendBucket struct {
	Period   string  `json:"period"`
	Value    float64 `json:"value"`
	Count    int64   `json:"count"`
	GroupKey string  `json:"group_key,omitempty"` // populated when drill_by is set
}

// TrendAnalysisResult holds the trend output.
type TrendAnalysisResult struct {
	MetricCode string        `json:"metric_code"`
	Interval   string        `json:"interval"`
	Function   string        `json:"function"`
	DrillBy    string        `json:"drill_by,omitempty"`
	Buckets    []TrendBucket `json:"buckets"`
}

func (s *AnalysisService) TrendAnalysis(ctx context.Context, p TrendAnalysisParams) (*TrendAnalysisResult, error) {
	interval := normalizeInterval(p.Interval)
	fn := normalizeFunction(p.Function)

	startTime, err := time.Parse(time.RFC3339, p.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, p.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time: %w", err)
	}

	dateFmt := intervalToDateFormat(interval)
	aggExpr := functionToSQL(fn)

	// Build SELECT
	selectParts := []string{
		fmt.Sprintf("DATE_FORMAT(event_time, '%s') AS period", dateFmt),
		fmt.Sprintf("%s AS value", aggExpr),
		"COUNT(*) AS count",
	}

	drillBy := strings.ToLower(p.DrillBy)
	if drillBy == "plot_id" || drillBy == "device_id" {
		selectParts = append(selectParts, fmt.Sprintf("CAST(%s AS CHAR) AS group_key", drillBy))
	}

	q := s.db.WithContext(ctx).Model(&models.MonitoringData{}).
		Select(strings.Join(selectParts, ", ")).
		Where("metric_code = ? AND event_time >= ? AND event_time <= ?", p.MetricCode, startTime, endTime)

	if p.PlotID > 0 {
		q = q.Where("plot_id = ?", p.PlotID)
	}
	if p.DeviceID > 0 {
		q = q.Where("device_id = ?", p.DeviceID)
	}

	// GROUP BY
	groupParts := []string{"period"}
	if drillBy == "plot_id" || drillBy == "device_id" {
		groupParts = append(groupParts, "group_key")
	}
	q = q.Group(strings.Join(groupParts, ", ")).Order("period ASC")

	var buckets []TrendBucket
	if err := q.Find(&buckets).Error; err != nil {
		return nil, fmt.Errorf("trend analysis query: %w", err)
	}

	return &TrendAnalysisResult{
		MetricCode: p.MetricCode,
		Interval:   interval,
		Function:   fn,
		DrillBy:    drillBy,
		Buckets:    buckets,
	}, nil
}

// ============================================================
// Funnel Analysis
// ============================================================

// FunnelAnalysisParams defines a funnel query.
// Stages are ordered metric_codes representing sequential steps
// (e.g. ["soil_prep", "planting", "growth", "harvest"]).
type FunnelAnalysisParams struct {
	Stages    []string `json:"stages"     binding:"required,min=2"` // ordered metric_codes
	StartTime string   `json:"start_time" binding:"required"`
	EndTime   string   `json:"end_time"   binding:"required"`
	PlotID    uint     `json:"plot_id"`
	DeviceID  uint     `json:"device_id"`
	DrillBy   string   `json:"drill_by"` // plot_id or device_id
}

// FunnelStage is one step in the funnel.
type FunnelStage struct {
	Stage          string  `json:"stage"`
	Count          int64   `json:"count"`
	ConversionRate float64 `json:"conversion_rate"` // percentage from previous stage (100% for first)
	GroupKey       string  `json:"group_key,omitempty"`
}

// FunnelAnalysisResult holds the funnel output.
type FunnelAnalysisResult struct {
	Stages  []string      `json:"stages"`
	DrillBy string        `json:"drill_by,omitempty"`
	Steps   []FunnelStage `json:"steps"`
}

func (s *AnalysisService) FunnelAnalysis(ctx context.Context, p FunnelAnalysisParams) (*FunnelAnalysisResult, error) {
	startTime, err := time.Parse(time.RFC3339, p.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, p.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time: %w", err)
	}

	drillBy := strings.ToLower(p.DrillBy)

	// For each stage, count distinct devices (or plots) that recorded that metric
	var steps []FunnelStage

	for i, stage := range p.Stages {
		q := s.db.WithContext(ctx).Model(&models.MonitoringData{}).
			Where("metric_code = ? AND event_time >= ? AND event_time <= ?", stage, startTime, endTime)

		if p.PlotID > 0 {
			q = q.Where("plot_id = ?", p.PlotID)
		}
		if p.DeviceID > 0 {
			q = q.Where("device_id = ?", p.DeviceID)
		}

		if drillBy == "plot_id" || drillBy == "device_id" {
			// Drill-down: count per group
			type groupCount struct {
				GroupKey string `json:"group_key"`
				Count    int64  `json:"count"`
			}
			var groups []groupCount
			err := q.Select(fmt.Sprintf("CAST(%s AS CHAR) AS group_key, COUNT(DISTINCT device_id) AS count", drillBy)).
				Group("group_key").
				Find(&groups).Error
			if err != nil {
				return nil, fmt.Errorf("funnel stage %q drill query: %w", stage, err)
			}
			for _, g := range groups {
				rate := 100.0
				if i > 0 {
					prev := findStepCount(steps, i-1, g.GroupKey)
					if prev > 0 {
						rate = float64(g.Count) / float64(prev) * 100
					} else {
						rate = 0
					}
				}
				steps = append(steps, FunnelStage{
					Stage:          stage,
					Count:          g.Count,
					ConversionRate: rate,
					GroupKey:       g.GroupKey,
				})
			}
		} else {
			// Aggregate: single count per stage
			var count int64
			if err := q.Select("COUNT(DISTINCT device_id)").Scan(&count).Error; err != nil {
				return nil, fmt.Errorf("funnel stage %q count: %w", stage, err)
			}
			rate := 100.0
			if i > 0 && len(steps) > 0 {
				prev := steps[i-1].Count
				if prev > 0 {
					rate = float64(count) / float64(prev) * 100
				} else {
					rate = 0
				}
			}
			steps = append(steps, FunnelStage{
				Stage:          stage,
				Count:          count,
				ConversionRate: rate,
			})
		}
	}

	return &FunnelAnalysisResult{
		Stages:  p.Stages,
		DrillBy: drillBy,
		Steps:   steps,
	}, nil
}

// findStepCount finds the count for a previous stage index for a given group key.
func findStepCount(steps []FunnelStage, stageIdx int, groupKey string) int64 {
	targetStageCount := 0
	for _, s := range steps {
		if s.GroupKey == groupKey {
			if targetStageCount == stageIdx {
				return s.Count
			}
			targetStageCount++
		}
	}
	return 0
}

// ============================================================
// Retention Analysis
// ============================================================

// RetentionAnalysisParams defines a retention cohort query.
type RetentionAnalysisParams struct {
	MetricCode     string `json:"metric_code"     binding:"required"`
	StartTime      string `json:"start_time"      binding:"required"`
	EndTime        string `json:"end_time"        binding:"required"`
	CohortInterval string `json:"cohort_interval"` // daily, weekly, monthly (default weekly)
	PlotID         uint   `json:"plot_id"`
	DeviceID       uint   `json:"device_id"`
	DrillBy        string `json:"drill_by"` // plot_id or device_id
}

// RetentionCohort is a single cohort row.
type RetentionCohort struct {
	CohortPeriod string    `json:"cohort_period"`
	CohortSize   int64     `json:"cohort_size"`           // devices active in the cohort period
	Retention    []float64 `json:"retention"`              // retention rates for subsequent periods (%)
	GroupKey     string    `json:"group_key,omitempty"`
}

// RetentionAnalysisResult holds the retention matrix.
type RetentionAnalysisResult struct {
	MetricCode     string            `json:"metric_code"`
	CohortInterval string            `json:"cohort_interval"`
	DrillBy        string            `json:"drill_by,omitempty"`
	Cohorts        []RetentionCohort `json:"cohorts"`
}

func (s *AnalysisService) RetentionAnalysis(ctx context.Context, p RetentionAnalysisParams) (*RetentionAnalysisResult, error) {
	startTime, err := time.Parse(time.RFC3339, p.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, p.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time: %w", err)
	}

	cohortInterval := normalizeInterval(p.CohortInterval)
	if cohortInterval == "" {
		cohortInterval = "weekly"
	}
	dateFmt := intervalToDateFormat(cohortInterval)
	drillBy := strings.ToLower(p.DrillBy)

	// Step 1: Get all distinct periods in the range
	type periodRow struct {
		Period string `json:"period"`
	}
	periodQ := s.db.WithContext(ctx).Model(&models.MonitoringData{}).
		Select(fmt.Sprintf("DISTINCT DATE_FORMAT(event_time, '%s') AS period", dateFmt)).
		Where("metric_code = ? AND event_time >= ? AND event_time <= ?", p.MetricCode, startTime, endTime)
	if p.PlotID > 0 {
		periodQ = periodQ.Where("plot_id = ?", p.PlotID)
	}
	if p.DeviceID > 0 {
		periodQ = periodQ.Where("device_id = ?", p.DeviceID)
	}
	periodQ = periodQ.Order("period ASC")

	var periods []periodRow
	if err := periodQ.Find(&periods).Error; err != nil {
		return nil, fmt.Errorf("retention periods query: %w", err)
	}

	if len(periods) == 0 {
		return &RetentionAnalysisResult{
			MetricCode:     p.MetricCode,
			CohortInterval: cohortInterval,
			DrillBy:        drillBy,
			Cohorts:        []RetentionCohort{},
		}, nil
	}

	// Step 2: For each cohort period, find devices active in that period,
	// then check which periods they remain active in.
	var cohorts []RetentionCohort

	for cohortIdx, cp := range periods {
		// Devices active in cohort period
		deviceQ := s.db.WithContext(ctx).Model(&models.MonitoringData{}).
			Select("DISTINCT device_id").
			Where("metric_code = ? AND DATE_FORMAT(event_time, ?) = ?", p.MetricCode, dateFmt, cp.Period)
		if p.PlotID > 0 {
			deviceQ = deviceQ.Where("plot_id = ?", p.PlotID)
		}
		if p.DeviceID > 0 {
			deviceQ = deviceQ.Where("device_id = ?", p.DeviceID)
		}

		type deviceRow struct {
			DeviceID uint
		}
		var cohortDevices []deviceRow
		if err := deviceQ.Find(&cohortDevices).Error; err != nil {
			return nil, fmt.Errorf("retention cohort devices query: %w", err)
		}

		cohortSize := int64(len(cohortDevices))
		if cohortSize == 0 {
			cohorts = append(cohorts, RetentionCohort{
				CohortPeriod: cp.Period,
				CohortSize:   0,
				Retention:    []float64{},
			})
			continue
		}

		deviceIDs := make([]uint, len(cohortDevices))
		for i, d := range cohortDevices {
			deviceIDs[i] = d.DeviceID
		}

		// For subsequent periods, check how many of these devices are still active
		retention := make([]float64, 0, len(periods)-cohortIdx)
		for j := cohortIdx; j < len(periods); j++ {
			var activeCount int64
			retQ := s.db.WithContext(ctx).Model(&models.MonitoringData{}).
				Where("metric_code = ? AND DATE_FORMAT(event_time, ?) = ? AND device_id IN ?",
					p.MetricCode, dateFmt, periods[j].Period, deviceIDs)
			if p.PlotID > 0 {
				retQ = retQ.Where("plot_id = ?", p.PlotID)
			}

			if err := retQ.Select("COUNT(DISTINCT device_id)").Scan(&activeCount).Error; err != nil {
				return nil, fmt.Errorf("retention count query: %w", err)
			}

			rate := float64(activeCount) / float64(cohortSize) * 100
			retention = append(retention, rate)
		}

		cohorts = append(cohorts, RetentionCohort{
			CohortPeriod: cp.Period,
			CohortSize:   cohortSize,
			Retention:    retention,
		})
	}

	return &RetentionAnalysisResult{
		MetricCode:     p.MetricCode,
		CohortInterval: cohortInterval,
		DrillBy:        drillBy,
		Cohorts:        cohorts,
	}, nil
}

// ============================================================
// Helpers
// ============================================================

func normalizeInterval(interval string) string {
	switch strings.ToLower(interval) {
	case "daily", "weekly", "monthly":
		return strings.ToLower(interval)
	default:
		return "daily"
	}
}

func normalizeFunction(fn string) string {
	switch strings.ToLower(fn) {
	case "avg", "sum", "min", "max", "count":
		return strings.ToLower(fn)
	default:
		return "avg"
	}
}

func intervalToDateFormat(interval string) string {
	switch interval {
	case "weekly":
		return "%Y-%u"
	case "monthly":
		return "%Y-%m"
	default:
		return "%Y-%m-%d"
	}
}

func functionToSQL(fn string) string {
	switch fn {
	case "sum":
		return "SUM(value)"
	case "min":
		return "MIN(value)"
	case "max":
		return "MAX(value)"
	case "count":
		return "COUNT(*)"
	default:
		return "AVG(value)"
	}
}

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var ErrMonitoringDataNotFound = errors.New("monitoring data not found")

// MonitoringDataService handles batch ingestion, aggregation, curves, and trends.
type MonitoringDataService struct {
	db    *gorm.DB
	queue *QueueService
}

func NewMonitoringDataService(db *gorm.DB, queue *QueueService) *MonitoringDataService {
	svc := &MonitoringDataService{db: db, queue: queue}
	queue.RegisterHandler("batch_ingest", svc.handleBatchIngest)
	return svc
}

// --- Batch Ingestion via Async Queue ---

type MonitoringDataInput struct {
	SourceID   string  `json:"source_id"   binding:"required"`
	DeviceID   uint    `json:"device_id"   binding:"required"`
	PlotID     uint    `json:"plot_id"     binding:"required"`
	MetricCode string  `json:"metric_code" binding:"required,max=100"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"        binding:"max=50"`
	EventTime  string  `json:"event_time"  binding:"required"` // RFC3339
	Tags       string  `json:"tags"`                           // JSON string
}

type BatchMonitoringDataInput struct {
	Data []MonitoringDataInput `json:"data" binding:"required,min=1,dive"`
}

// SubmitBatchIngest enqueues a batch of monitoring data for async processing.
func (s *MonitoringDataService) SubmitBatchIngest(data []MonitoringDataInput) (*Job, error) {
	// Serialize inputs into payload
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal batch data: %w", err)
	}
	payload := map[string]interface{}{
		"data": string(raw),
	}
	return s.queue.Submit("batch_ingest", payload)
}

// handleBatchIngest is the queue handler for batch_ingest jobs.
func (s *MonitoringDataService) handleBatchIngest(ctx context.Context, job *Job) (string, error) {
	rawData, ok := job.Payload["data"].(string)
	if !ok {
		return "", fmt.Errorf("missing data in payload")
	}

	var inputs []MonitoringDataInput
	if err := json.Unmarshal([]byte(rawData), &inputs); err != nil {
		return "", fmt.Errorf("unmarshal batch data: %w", err)
	}

	inserted := 0
	skipped := 0

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return "", fmt.Errorf("begin tx: %w", tx.Error)
	}

	for _, in := range inputs {
		eventTime, err := time.Parse(time.RFC3339, in.EventTime)
		if err != nil {
			tx.Rollback()
			return "", fmt.Errorf("invalid event_time %q: %w", in.EventTime, err)
		}

		record := models.MonitoringData{
			SourceID:   in.SourceID,
			DeviceID:   in.DeviceID,
			PlotID:     in.PlotID,
			MetricCode: in.MetricCode,
			Value:      in.Value,
			Unit:       in.Unit,
			EventTime:  eventTime,
			Tags:       in.Tags,
		}

		result := tx.Create(&record)
		if result.Error != nil {
			// Check for duplicate key (idempotent skip)
			if isDuplicateKeyError(result.Error) {
				skipped++
				continue
			}
			tx.Rollback()
			return "", fmt.Errorf("insert monitoring data: %w", result.Error)
		}
		inserted++
	}

	if err := tx.Commit().Error; err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}

	resultJSON, _ := json.Marshal(map[string]interface{}{
		"inserted": inserted,
		"skipped":  skipped,
		"total":    len(inputs),
	})
	return string(resultJSON), nil
}

func isDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "Duplicate entry") ||
		strings.Contains(err.Error(), "UNIQUE constraint failed") ||
		strings.Contains(err.Error(), "1062")
}

// --- Multi-Dimensional Aggregated Queries ---

type AggregationParams struct {
	PlotID     uint   `json:"plot_id"`
	DeviceID   uint   `json:"device_id"`
	MetricCode string `json:"metric_code"`
	StartTime  string `json:"start_time"` // RFC3339
	EndTime    string `json:"end_time"`   // RFC3339
	Tags       string `json:"tags"`       // JSON key-value to match
	Function   string `json:"function"`   // count, min, max, avg, sum
	GroupBy    string `json:"group_by"`   // metric_code, device_id, plot_id, or empty
}

type AggregationResult struct {
	GroupKey string  `json:"group_key,omitempty"`
	Value   float64 `json:"value"`
	Count   int64   `json:"count,omitempty"`
}

func (s *MonitoringDataService) Aggregate(ctx context.Context, p AggregationParams) ([]AggregationResult, error) {
	fn := strings.ToLower(p.Function)
	if fn == "" {
		fn = "count"
	}

	var selectExpr string
	switch fn {
	case "count":
		selectExpr = "COUNT(*) as value"
	case "min":
		selectExpr = "MIN(value) as value"
	case "max":
		selectExpr = "MAX(value) as value"
	case "avg":
		selectExpr = "AVG(value) as value"
	case "sum":
		selectExpr = "SUM(value) as value"
	default:
		return nil, fmt.Errorf("unsupported aggregation function: %s", fn)
	}

	q := s.db.WithContext(ctx).Model(&models.MonitoringData{})
	q = applyMonitoringFilters(q, p.PlotID, p.DeviceID, p.MetricCode, p.StartTime, p.EndTime, p.Tags)

	groupBy := strings.ToLower(p.GroupBy)
	var groupCol string
	switch groupBy {
	case "metric_code":
		groupCol = "metric_code"
	case "device_id":
		groupCol = "device_id"
	case "plot_id":
		groupCol = "plot_id"
	case "":
		// No grouping — single result
	default:
		return nil, fmt.Errorf("unsupported group_by: %s", p.GroupBy)
	}

	if groupCol != "" {
		selectExpr = fmt.Sprintf("%s as group_key, %s", groupCol, selectExpr)
		q = q.Select(selectExpr).Group(groupCol)
	} else {
		q = q.Select(selectExpr)
	}

	var results []AggregationResult
	if err := q.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("aggregate query: %w", err)
	}
	return results, nil
}

// --- Real-Time Curves ---

type CurveParams struct {
	PlotID     uint   `json:"plot_id"`
	DeviceID   uint   `json:"device_id"`
	MetricCode string `json:"metric_code" binding:"required"`
	Minutes    int    `json:"minutes"` // last N minutes, default 60
}

type CurvePoint struct {
	EventTime time.Time `json:"event_time"`
	Value     float64   `json:"value"`
}

func (s *MonitoringDataService) RealtimeCurve(ctx context.Context, p CurveParams) ([]CurvePoint, error) {
	if p.Minutes <= 0 {
		p.Minutes = 60
	}

	since := time.Now().Add(-time.Duration(p.Minutes) * time.Minute)

	q := s.db.WithContext(ctx).Model(&models.MonitoringData{}).
		Select("event_time, value").
		Where("event_time >= ?", since)

	if p.DeviceID > 0 {
		q = q.Where("device_id = ?", p.DeviceID)
	}
	if p.PlotID > 0 {
		q = q.Where("plot_id = ?", p.PlotID)
	}
	if p.MetricCode != "" {
		q = q.Where("metric_code = ?", p.MetricCode)
	}

	q = q.Order("event_time ASC").Limit(1000)

	var points []CurvePoint
	if err := q.Find(&points).Error; err != nil {
		return nil, fmt.Errorf("realtime curve query: %w", err)
	}
	return points, nil
}

// --- Trend Statistics (Daily/Weekly/Monthly with YoY and MoM) ---

type TrendParams struct {
	PlotID     uint   `json:"plot_id"`
	DeviceID   uint   `json:"device_id"`
	MetricCode string `json:"metric_code" binding:"required"`
	StartTime  string `json:"start_time" binding:"required"` // RFC3339
	EndTime    string `json:"end_time"   binding:"required"` // RFC3339
	Interval   string `json:"interval"`                      // daily, weekly, monthly
	Function   string `json:"function"`                      // avg, sum, min, max, count
}

type TrendPoint struct {
	Period string  `json:"period"` // formatted date bucket
	Value  float64 `json:"value"`
	Count  int64   `json:"count"`
}

type TrendResult struct {
	Current  []TrendPoint `json:"current"`
	Previous []TrendPoint `json:"previous,omitempty"` // YoY or MoM comparison
}

func (s *MonitoringDataService) TrendStatistics(ctx context.Context, p TrendParams) (*TrendResult, error) {
	interval := strings.ToLower(p.Interval)
	if interval == "" {
		interval = "daily"
	}
	fn := strings.ToLower(p.Function)
	if fn == "" {
		fn = "avg"
	}

	startTime, err := time.Parse(time.RFC3339, p.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, p.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time: %w", err)
	}

	// Get current period trends
	current, err := s.queryTrend(ctx, p, startTime, endTime, interval, fn)
	if err != nil {
		return nil, fmt.Errorf("current trend: %w", err)
	}

	// Calculate comparison period for YoY or MoM
	var prevStart, prevEnd time.Time
	switch interval {
	case "monthly":
		// YoY: same period one year ago
		prevStart = startTime.AddDate(-1, 0, 0)
		prevEnd = endTime.AddDate(-1, 0, 0)
	case "weekly", "daily":
		// MoM: same duration one month ago
		prevStart = startTime.AddDate(0, -1, 0)
		prevEnd = endTime.AddDate(0, -1, 0)
	default:
		return &TrendResult{Current: current}, nil
	}

	previous, err := s.queryTrend(ctx, p, prevStart, prevEnd, interval, fn)
	if err != nil {
		return nil, fmt.Errorf("previous trend: %w", err)
	}

	return &TrendResult{
		Current:  current,
		Previous: previous,
	}, nil
}

func (s *MonitoringDataService) queryTrend(ctx context.Context, p TrendParams, start, end time.Time, interval, fn string) ([]TrendPoint, error) {
	var dateFormat string
	switch interval {
	case "daily":
		dateFormat = "%Y-%m-%d"
	case "weekly":
		dateFormat = "%Y-%u" // year-week
	case "monthly":
		dateFormat = "%Y-%m"
	default:
		dateFormat = "%Y-%m-%d"
	}

	aggExpr := aggSQLExpr(fn)

	q := s.db.WithContext(ctx).Model(&models.MonitoringData{}).
		Select(fmt.Sprintf("DATE_FORMAT(event_time, '%s') as period, %s as value, COUNT(*) as count", dateFormat, aggExpr)).
		Where("event_time >= ? AND event_time <= ?", start, end)

	if p.DeviceID > 0 {
		q = q.Where("device_id = ?", p.DeviceID)
	}
	if p.PlotID > 0 {
		q = q.Where("plot_id = ?", p.PlotID)
	}
	if p.MetricCode != "" {
		q = q.Where("metric_code = ?", p.MetricCode)
	}

	q = q.Group("period").Order("period ASC")

	var points []TrendPoint
	if err := q.Find(&points).Error; err != nil {
		return nil, err
	}
	return points, nil
}

func aggSQLExpr(fn string) string {
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

// --- Query / List ---

type MonitoringDataListParams struct {
	Page       int
	PageSize   int
	PlotID     uint
	DeviceID   uint
	MetricCode string
	StartTime  string
	EndTime    string
	Tags       string
}

type PaginatedMonitoringData struct {
	Data       []models.MonitoringData `json:"data"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

func (s *MonitoringDataService) List(ctx context.Context, p MonitoringDataListParams) (*PaginatedMonitoringData, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.MonitoringData{})
	q = applyMonitoringFilters(q, p.PlotID, p.DeviceID, p.MetricCode, p.StartTime, p.EndTime, p.Tags)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count monitoring data: %w", err)
	}

	var data []models.MonitoringData
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("event_time DESC").Offset(offset).Limit(p.PageSize).Find(&data).Error; err != nil {
		return nil, fmt.Errorf("list monitoring data: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedMonitoringData{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *MonitoringDataService) GetByID(ctx context.Context, id uint) (*models.MonitoringData, error) {
	var record models.MonitoringData
	if err := s.db.WithContext(ctx).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMonitoringDataNotFound
		}
		return nil, fmt.Errorf("get monitoring data: %w", err)
	}
	return &record, nil
}

// --- Export ---

type ExportParams struct {
	PlotID     uint
	DeviceID   uint
	MetricCode string
	StartTime  string
	EndTime    string
	Tags       string
}

func (s *MonitoringDataService) ExportData(ctx context.Context, p ExportParams) ([]models.MonitoringData, error) {
	q := s.db.WithContext(ctx).Model(&models.MonitoringData{})
	q = applyMonitoringFilters(q, p.PlotID, p.DeviceID, p.MetricCode, p.StartTime, p.EndTime, p.Tags)
	q = q.Order("event_time ASC").Limit(10000)

	var data []models.MonitoringData
	if err := q.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("export monitoring data: %w", err)
	}
	return data, nil
}

// --- Helpers ---

func applyMonitoringFilters(q *gorm.DB, plotID, deviceID uint, metricCode, startTime, endTime, tags string) *gorm.DB {
	if plotID > 0 {
		q = q.Where("plot_id = ?", plotID)
	}
	if deviceID > 0 {
		q = q.Where("device_id = ?", deviceID)
	}
	if metricCode != "" {
		q = q.Where("metric_code = ?", metricCode)
	}
	if startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			q = q.Where("event_time >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			q = q.Where("event_time <= ?", t)
		}
	}
	if tags != "" {
		// JSON contains-based tag filtering
		q = q.Where("tags LIKE ?", "%"+tags+"%")
	}
	return q
}

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var ErrAlertNotFound = errors.New("alert not found")

// MonitorService handles device monitoring, threshold checks, and alert creation.
type MonitorService struct {
	db    *gorm.DB
	queue *QueueService
}

// NewMonitorService creates a MonitorService and registers its queue handlers.
func NewMonitorService(db *gorm.DB, queue *QueueService) *MonitorService {
	svc := &MonitorService{db: db, queue: queue}
	queue.RegisterHandler("monitor_device", svc.handleMonitorDevice)
	queue.RegisterHandler("threshold_check", svc.handleThresholdCheck)
	return svc
}

// MonitorDeviceInput is the payload for a device monitoring job.
type MonitorDeviceInput struct {
	DeviceID uint `json:"device_id" binding:"required"`
}

// ThresholdCheckInput defines a threshold check job.
type ThresholdCheckInput struct {
	DeviceID   uint    `json:"device_id"   binding:"required"`
	MetricType string  `json:"metric_type" binding:"required"`
	Threshold  float64 `json:"threshold"   binding:"required"`
	Level      string  `json:"level"       binding:"omitempty,oneof=warning critical"`
}

// SubmitMonitorDevice enqueues a device monitoring job.
func (s *MonitorService) SubmitMonitorDevice(deviceID uint) (*Job, error) {
	payload := map[string]interface{}{"device_id": float64(deviceID)}
	return s.queue.Submit("monitor_device", payload)
}

// SubmitThresholdCheck enqueues a threshold check job.
func (s *MonitorService) SubmitThresholdCheck(in ThresholdCheckInput) (*Job, error) {
	payload := map[string]interface{}{
		"device_id":   float64(in.DeviceID),
		"metric_type": in.MetricType,
		"threshold":   in.Threshold,
		"level":       in.Level,
	}
	return s.queue.Submit("threshold_check", payload)
}

// ListAlerts returns paginated alerts with optional filters.
type AlertListParams struct {
	Page     int
	PageSize int
	DeviceID uint
	Level    string
	Resolved *bool
}

type PaginatedAlerts struct {
	Data       []models.Alert `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

func (s *MonitorService) ListAlerts(ctx context.Context, p AlertListParams) (*PaginatedAlerts, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Alert{})
	if p.DeviceID > 0 {
		q = q.Where("device_id = ?", p.DeviceID)
	}
	if p.Level != "" {
		q = q.Where("level = ?", p.Level)
	}
	if p.Resolved != nil {
		q = q.Where("resolved = ?", *p.Resolved)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count alerts: %w", err)
	}

	var alerts []models.Alert
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("id DESC").Offset(offset).Limit(p.PageSize).Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedAlerts{
		Data:       alerts,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ResolveAlert marks an alert as resolved.
func (s *MonitorService) ResolveAlert(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Model(&models.Alert{}).Where("id = ?", id).Update("resolved", true)
	if res.Error != nil {
		return fmt.Errorf("resolve alert: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrAlertNotFound
	}
	return nil
}

// --- Queue handlers (run inside workers) ---

func (s *MonitorService) handleMonitorDevice(ctx context.Context, job *Job) (string, error) {
	deviceID := uint(job.Payload["device_id"].(float64))

	var device models.Device
	if err := s.db.WithContext(ctx).First(&device, deviceID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("device %d not found", deviceID)
		}
		return "", fmt.Errorf("query device: %w", err)
	}

	// Check latest metric: if no reading in the last hour, create a warning
	var latestMetric models.Metric
	err := s.db.WithContext(ctx).
		Where("device_id = ?", deviceID).
		Order("event_time DESC").
		First(&latestMetric).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		alert := models.Alert{
			DeviceID:   deviceID,
			MetricType: "none",
			Value:      0,
			Threshold:  0,
			Level:      "warning",
			Message:    fmt.Sprintf("device %d (%s) has no metrics recorded", deviceID, device.Name),
		}
		s.db.WithContext(ctx).Create(&alert)
		return fmt.Sprintf("alert created: no metrics for device %d", deviceID), nil
	} else if err != nil {
		return "", fmt.Errorf("query latest metric: %w", err)
	}

	oneHourAgo := time.Now().Add(-1 * time.Hour)
	if latestMetric.EventTime.Before(oneHourAgo) {
		alert := models.Alert{
			DeviceID:   deviceID,
			MetricType: latestMetric.MetricType,
			Value:      0,
			Threshold:  0,
			Level:      "warning",
			Message:    fmt.Sprintf("device %d (%s) last reading is stale (>1h ago)", deviceID, device.Name),
		}
		s.db.WithContext(ctx).Create(&alert)
		return fmt.Sprintf("alert created: stale data for device %d", deviceID), nil
	}

	return fmt.Sprintf("device %d is healthy, last reading at %s", deviceID, latestMetric.EventTime.Format(time.RFC3339)), nil
}

func (s *MonitorService) handleThresholdCheck(ctx context.Context, job *Job) (string, error) {
	deviceID := uint(job.Payload["device_id"].(float64))
	metricType := job.Payload["metric_type"].(string)
	threshold := job.Payload["threshold"].(float64)
	level := "warning"
	if l, ok := job.Payload["level"].(string); ok && l != "" {
		level = l
	}

	// Find recent metrics that exceed the threshold
	var breaching []models.Metric
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	err := s.db.WithContext(ctx).
		Where("device_id = ? AND metric_type = ? AND value > ? AND event_time > ?",
			deviceID, metricType, threshold, oneHourAgo).
		Order("event_time DESC").
		Limit(10).
		Find(&breaching).Error
	if err != nil {
		return "", fmt.Errorf("query breaching metrics: %w", err)
	}

	if len(breaching) == 0 {
		return fmt.Sprintf("no threshold breaches for device %d, metric %s (threshold %.2f)", deviceID, metricType, threshold), nil
	}

	// Create alert for the worst breach
	worst := breaching[0]
	alert := models.Alert{
		DeviceID:   deviceID,
		MetricType: metricType,
		Value:      worst.Value,
		Threshold:  threshold,
		Level:      level,
		Message:    fmt.Sprintf("%d readings exceeded %.2f in the last hour (worst: %.2f)", len(breaching), threshold, worst.Value),
	}
	if err := s.db.WithContext(ctx).Create(&alert).Error; err != nil {
		return "", fmt.Errorf("create alert: %w", err)
	}

	result, _ := json.Marshal(map[string]interface{}{
		"alert_id":       alert.ID,
		"breach_count":   len(breaching),
		"worst_value":    worst.Value,
		"threshold":      threshold,
	})
	return string(result), nil
}

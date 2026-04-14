package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

var ErrMetricNotFound = errors.New("metric not found")

type MetricService struct {
	db *gorm.DB
}

func NewMetricService(db *gorm.DB) *MetricService {
	return &MetricService{db: db}
}

type CreateMetricInput struct {
	DeviceID   uint    `json:"device_id"   binding:"required"`
	MetricType string  `json:"metric_type" binding:"required,max=100"`
	Value      float64 `json:"value"       binding:"required"`
	Unit       string  `json:"unit"        binding:"max=50"`
	EventTime  string  `json:"event_time"  binding:"required"`
}

type BatchMetricInput struct {
	Metrics []CreateMetricInput `json:"metrics" binding:"required,min=1,dive"`
}

func (s *MetricService) Create(ctx context.Context, in CreateMetricInput) (*models.Metric, error) {
	eventTime, err := time.Parse(time.RFC3339, in.EventTime)
	if err != nil {
		return nil, fmt.Errorf("invalid event_time format (use RFC3339): %w", err)
	}

	metric := models.Metric{
		DeviceID:   in.DeviceID,
		MetricType: in.MetricType,
		Value:      in.Value,
		Unit:       in.Unit,
		EventTime:  eventTime,
	}
	if err := s.db.WithContext(ctx).Create(&metric).Error; err != nil {
		return nil, fmt.Errorf("create metric: %w", err)
	}
	return &metric, nil
}

func (s *MetricService) BatchCreate(ctx context.Context, inputs []CreateMetricInput) ([]models.Metric, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin tx: %w", tx.Error)
	}

	metrics := make([]models.Metric, 0, len(inputs))
	for _, in := range inputs {
		eventTime, err := time.Parse(time.RFC3339, in.EventTime)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("invalid event_time format (use RFC3339): %w", err)
		}
		m := models.Metric{
			DeviceID:   in.DeviceID,
			MetricType: in.MetricType,
			Value:      in.Value,
			Unit:       in.Unit,
			EventTime:  eventTime,
		}
		if err := tx.Create(&m).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("batch create metric: %w", err)
		}
		metrics = append(metrics, m)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return metrics, nil
}

type MetricListParams struct {
	Page       int
	PageSize   int
	DeviceID   uint
	MetricType string
	StartTime  string
	EndTime    string
}

type PaginatedMetrics struct {
	Data       []models.Metric `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalPages int             `json:"total_pages"`
}

func (s *MetricService) List(ctx context.Context, p MetricListParams) (*PaginatedMetrics, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&models.Metric{})
	if p.DeviceID > 0 {
		q = q.Where("device_id = ?", p.DeviceID)
	}
	if p.MetricType != "" {
		q = q.Where("metric_type = ?", p.MetricType)
	}
	if p.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, p.StartTime); err == nil {
			q = q.Where("event_time >= ?", t)
		}
	}
	if p.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, p.EndTime); err == nil {
			q = q.Where("event_time <= ?", t)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count metrics: %w", err)
	}

	var metrics []models.Metric
	offset := (p.Page - 1) * p.PageSize
	if err := q.Order("event_time DESC").Offset(offset).Limit(p.PageSize).Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("list metrics: %w", err)
	}

	totalPages := int(total) / p.PageSize
	if int(total)%p.PageSize != 0 {
		totalPages++
	}

	return &PaginatedMetrics{
		Data:       metrics,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *MetricService) GetByID(ctx context.Context, id uint) (*models.Metric, error) {
	var metric models.Metric
	if err := s.db.WithContext(ctx).First(&metric, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMetricNotFound
		}
		return nil, fmt.Errorf("get metric: %w", err)
	}
	return &metric, nil
}

func (s *MetricService) Delete(ctx context.Context, id uint) error {
	res := s.db.WithContext(ctx).Delete(&models.Metric{}, id)
	if res.Error != nil {
		return fmt.Errorf("delete metric: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrMetricNotFound
	}
	return nil
}

package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

const (
	DiskUsageThreshold = 80.0 // percentage
	CapacityCheckInterval = 60 * time.Second
)

// CapacityService monitors system capacity and creates notifications when thresholds are exceeded.
type CapacityService struct {
	db *gorm.DB
}

func NewCapacityService(db *gorm.DB) *CapacityService {
	return &CapacityService{db: db}
}

// CheckDiskUsage checks disk usage and inserts a notification if over threshold.
// Returns the usage percentage and any error.
func (s *CapacityService) CheckDiskUsage(ctx context.Context) (float64, error) {
	usage, err := getDiskUsagePercent()
	if err != nil {
		return 0, fmt.Errorf("get disk usage: %w", err)
	}

	if usage > DiskUsageThreshold {
		notification := models.SystemNotification{
			Type:    "capacity",
			Message: fmt.Sprintf("Disk usage at %.1f%% exceeds threshold of %.0f%%", usage, DiskUsageThreshold),
			Level:   "warning",
		}
		if usage > 90 {
			notification.Level = "critical"
		}
		if err := s.db.WithContext(ctx).Create(&notification).Error; err != nil {
			return usage, fmt.Errorf("create notification: %w", err)
		}
		log.Printf("capacity alert: disk usage %.1f%%", usage)
	}

	return usage, nil
}

// ListNotifications returns recent system notifications.
func (s *CapacityService) ListNotifications(ctx context.Context, page, pageSize int) ([]models.SystemNotification, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	if err := s.db.WithContext(ctx).Model(&models.SystemNotification{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count notifications: %w", err)
	}

	var notifications []models.SystemNotification
	offset := (page - 1) * pageSize
	if err := s.db.WithContext(ctx).Order("id DESC").Offset(offset).Limit(pageSize).Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("list notifications: %w", err)
	}

	return notifications, total, nil
}

// StartCapacityMonitor starts a background goroutine checking capacity every 60 seconds.
func (s *CapacityService) StartCapacityMonitor(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(CapacityCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("capacity monitor stopped")
				return
			case <-ticker.C:
				if _, err := s.CheckDiskUsage(ctx); err != nil {
					log.Printf("capacity check error: %v", err)
				}
			}
		}
	}()
}

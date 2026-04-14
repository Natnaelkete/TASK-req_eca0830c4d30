package models

import "time"

// MonitoringData stores a single time-series sensor reading.
// Idempotent key: (source_id, event_time, metric_code).
// Core index: (device_id, plot_id, metric_code, event_time).
type MonitoringData struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SourceID   string    `gorm:"size:255;not null;uniqueIndex:idx_monitoring_idempotent" json:"source_id"`
	DeviceID   uint      `gorm:"not null;index:idx_monitoring_core" json:"device_id"`
	PlotID     uint      `gorm:"not null;index:idx_monitoring_core" json:"plot_id"`
	MetricCode string    `gorm:"size:100;not null;uniqueIndex:idx_monitoring_idempotent;index:idx_monitoring_core" json:"metric_code"`
	Value      float64   `gorm:"not null" json:"value"`
	Unit       string    `gorm:"size:50" json:"unit"`
	EventTime  time.Time `gorm:"not null;uniqueIndex:idx_monitoring_idempotent;index:idx_monitoring_core" json:"event_time"`
	Tags       string    `gorm:"type:text" json:"tags"` // JSON-encoded key-value pairs for tag filtering
	CreatedAt  time.Time `json:"created_at"`
}

func (MonitoringData) TableName() string { return "monitoring_data" }

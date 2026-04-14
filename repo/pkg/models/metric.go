package models

import "time"

// Metric stores a single sensor reading with its event timestamp.
type Metric struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceID   uint      `gorm:"not null;index" json:"device_id"`
	MetricType string    `gorm:"size:100;not null;index" json:"metric_type"`
	Value      float64   `gorm:"not null" json:"value"`
	Unit       string    `gorm:"size:50" json:"unit"`
	EventTime  time.Time `gorm:"not null;index" json:"event_time"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Metric) TableName() string { return "metrics" }

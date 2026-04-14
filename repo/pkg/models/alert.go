package models

import "time"

// Alert records a monitoring event triggered by a threshold breach or device issue.
type Alert struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	DeviceID   uint      `gorm:"not null;index" json:"device_id"`
	MetricType string    `gorm:"size:100;not null" json:"metric_type"`
	Value      float64   `gorm:"not null" json:"value"`
	Threshold  float64   `gorm:"not null" json:"threshold"`
	Level      string    `gorm:"size:50;not null;default:'warning';index" json:"level"` // warning, critical
	Message    string    `gorm:"size:500" json:"message"`
	Resolved   bool      `gorm:"not null;default:false" json:"resolved"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Alert) TableName() string { return "alerts" }

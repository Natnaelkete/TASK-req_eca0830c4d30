package models

import "time"

// DashboardConfig stores a user's custom dashboard configuration.
type DashboardConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index:idx_dashboard_user" json:"user_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Config    string    `gorm:"type:text;not null" json:"config"` // JSON: plots, devices, metrics, time window, tags, aggregation, chart type
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (DashboardConfig) TableName() string { return "dashboard_configs" }

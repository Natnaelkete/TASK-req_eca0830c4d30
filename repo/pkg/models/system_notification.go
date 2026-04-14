package models

import "time"

// SystemNotification records internal system alerts (e.g., disk usage > 80%).
type SystemNotification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"size:100;not null;index" json:"type"` // capacity, system, etc.
	Message   string    `gorm:"type:text;not null" json:"message"`
	Level     string    `gorm:"size:50;not null;default:'warning'" json:"level"` // info, warning, critical
	CreatedAt time.Time `json:"created_at"`
}

func (SystemNotification) TableName() string { return "system_notifications" }

package models

import "time"

// Device represents a sensor/actuator device deployed on a plot.
type Device struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Name         string     `gorm:"size:200;not null" json:"name"`
	Type         string     `gorm:"size:100;not null" json:"type"`
	SerialNumber string     `gorm:"uniqueIndex;size:100;not null" json:"serial_number"`
	PlotID       uint       `gorm:"not null;index" json:"plot_id"`
	Status       string     `gorm:"size:50;not null;default:'active'" json:"status"`
	InstalledAt  *time.Time `json:"installed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (Device) TableName() string { return "devices" }

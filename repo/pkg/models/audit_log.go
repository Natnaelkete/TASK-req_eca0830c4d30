package models

import "time"

// AuditLog records every API request for compliance and debugging.
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Action     string    `gorm:"size:100;not null" json:"action"`
	Resource   string    `gorm:"size:100;not null" json:"resource"`
	ResourceID uint      `json:"resource_id"`
	IPAddress  string    `gorm:"size:45" json:"ip_address"`
	CreatedAt  time.Time `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

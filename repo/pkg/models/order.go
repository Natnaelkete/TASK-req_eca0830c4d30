package models

import "time"

// Order represents a business order with an associated conversation thread.
type Order struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	ResearcherID uint      `gorm:"not null;index" json:"researcher_id"`
	Title        string    `gorm:"size:255;not null" json:"title"`
	Status       string    `gorm:"size:50;not null;default:'open'" json:"status"` // open, in_progress, closed
	AssignedTo   uint      `gorm:"index" json:"assigned_to"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

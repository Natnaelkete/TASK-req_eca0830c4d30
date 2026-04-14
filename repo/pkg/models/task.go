package models

import "time"

// Task represents an evaluation/research task assigned to a user.
type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"size:255;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Status      string     `gorm:"size:50;not null;default:'pending';index" json:"status"`
	AssignedTo  uint       `gorm:"index" json:"assigned_to"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Task) TableName() string { return "tasks" }

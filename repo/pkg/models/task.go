package models

import "time"

// Task represents an evaluation task with visibility windows and review flow.
// Overdue tasks are automatically marked as delayed after 7 days past DueEnd.
type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"size:255;not null" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	ObjectID    uint       `gorm:"not null;index" json:"object_id"`
	ObjectType  string     `gorm:"size:100;not null" json:"object_type"`
	CycleType   string     `gorm:"size:50" json:"cycle_type"`
	Status      string     `gorm:"size:50;not null;default:'pending';index" json:"status"`
	AssignedTo  uint       `gorm:"index" json:"assigned_to"`
	ReviewerID  *uint      `gorm:"index" json:"reviewer_id,omitempty"`
	DueStart    *time.Time `json:"due_start,omitempty"`
	DueEnd      *time.Time `json:"due_end,omitempty"`
	SubmittedAt *time.Time `json:"submitted_at,omitempty"`
	OverdueAt   *time.Time `json:"overdue_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Task) TableName() string { return "tasks" }

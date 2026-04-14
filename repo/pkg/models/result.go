package models

import "time"

// Result stores a research result or analysis output tied to a plot.
type Result struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PlotID    uint      `gorm:"not null;index" json:"plot_id"`
	TaskID    *uint     `gorm:"index" json:"task_id,omitempty"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Summary   string    `gorm:"type:text" json:"summary"`
	Data      string    `gorm:"type:text" json:"data"` // JSON-encoded result data
	CreatedBy uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Result) TableName() string { return "results" }

package models

import "time"

// IndicatorDefinition represents an analysis indicator (metric definition) that can be versioned.
type IndicatorDefinition struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:100;not null" json:"code"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Unit        string    `gorm:"size:50" json:"unit"`
	Formula     string    `gorm:"type:text" json:"formula"`
	Category    string    `gorm:"size:100;index" json:"category"`
	Status      string    `gorm:"size:50;not null;default:'active'" json:"status"` // active, deprecated
	CreatedBy   uint      `gorm:"not null;index" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (IndicatorDefinition) TableName() string { return "indicator_definitions" }

// IndicatorVersion tracks every change to an indicator definition for full audit-diff traceability.
type IndicatorVersion struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	IndicatorID  uint      `gorm:"not null;index" json:"indicator_id"`
	Version      int       `gorm:"not null" json:"version"`
	Name         string    `gorm:"size:255;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	Unit         string    `gorm:"size:50" json:"unit"`
	Formula      string    `gorm:"type:text" json:"formula"`
	Category     string    `gorm:"size:100" json:"category"`
	DiffSummary  string    `gorm:"type:text" json:"diff_summary"`  // human-readable change description
	ModifiedBy   uint      `gorm:"not null;index" json:"modified_by"`
	ModifiedAt   time.Time `gorm:"not null" json:"modified_at"`
	CreatedAt    time.Time `json:"created_at"`
}

func (IndicatorVersion) TableName() string { return "indicator_versions" }

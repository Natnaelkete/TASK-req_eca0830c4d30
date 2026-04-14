package models

import "time"

// Result stores a research result with strict status state machine.
// Status transitions: draft → submitted → returned → approved → archived.
type Result struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Type              string     `gorm:"size:50;not null" json:"type"`   // paper, project, patent
	PlotID            uint       `gorm:"not null;index" json:"plot_id"`
	TaskID            *uint      `gorm:"index" json:"task_id,omitempty"`
	Title             string     `gorm:"size:255;not null" json:"title"`
	Summary           string     `gorm:"type:text" json:"summary"`
	Fields            string     `gorm:"type:text" json:"fields"`  // JSON: actual field values
	Status            string     `gorm:"size:50;not null;default:'draft';index" json:"status"`
	SubmitterID       uint       `gorm:"not null;index" json:"submitter_id"`
	Notes             string     `gorm:"type:text" json:"notes"` // post-archive retrospective notes
	ArchivedAt        *time.Time `json:"archived_at,omitempty"`
	InvalidatedReason string     `gorm:"type:text" json:"invalidated_reason,omitempty"`
	InvalidatedBy     *uint      `json:"invalidated_by,omitempty"`
	InvalidatedAt     *time.Time `json:"invalidated_at,omitempty"`
	CreatedBy         uint       `gorm:"not null;index" json:"created_by"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (Result) TableName() string { return "results" }

// FieldRule defines configurable validation rules for result fields per type.
type FieldRule struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	ResultType string `gorm:"size:50;not null;index" json:"result_type"` // paper, project, patent
	FieldName  string `gorm:"size:100;not null" json:"field_name"`
	Required   bool   `gorm:"not null;default:false" json:"required"`
	MaxLength  int    `gorm:"default:0" json:"max_length"`            // 0 = no limit
	EnumValues string `gorm:"size:500" json:"enum_values,omitempty"`  // comma-separated allowed values
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (FieldRule) TableName() string { return "field_rules" }

// ResultStatusLog tracks every status transition for full traceability.
type ResultStatusLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ResultID   uint      `gorm:"not null;index" json:"result_id"`
	FromStatus string    `gorm:"size:50;not null" json:"from_status"`
	ToStatus   string    `gorm:"size:50;not null" json:"to_status"`
	ChangedBy  uint      `gorm:"not null" json:"changed_by"`
	Reason     string    `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (ResultStatusLog) TableName() string { return "result_status_logs" }

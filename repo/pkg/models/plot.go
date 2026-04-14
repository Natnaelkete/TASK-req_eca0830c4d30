package models

import "time"

// Plot represents an agricultural research plot.
type Plot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:200;not null" json:"name"`
	Location  string    `gorm:"size:255" json:"location"`
	Area      float64   `json:"area"`
	SoilType  string    `gorm:"size:100" json:"soil_type"`
	CropType  string    `gorm:"size:100" json:"crop_type"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Plot) TableName() string { return "plots" }

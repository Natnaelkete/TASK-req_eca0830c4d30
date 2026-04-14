package models

import "time"

// Message represents a chat message between users.
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	SenderID  uint      `gorm:"not null;index" json:"sender_id"`
	ReceiverID uint     `gorm:"not null;index" json:"receiver_id"`
	PlotID    *uint     `gorm:"index" json:"plot_id,omitempty"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Read      bool      `gorm:"not null;default:false" json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

func (Message) TableName() string { return "messages" }

package models

import "time"

// Conversation is a message within an order-level conversation thread.
type Conversation struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	OrderID       uint       `gorm:"not null;index" json:"order_id"`
	UserID        uint       `gorm:"not null;index" json:"user_id"`
	Message       string     `gorm:"type:text;not null" json:"message"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	TransferredTo *uint      `json:"transferred_to,omitempty"` // user ID if this is a transfer record
	CreatedAt     time.Time  `json:"created_at"`
}

func (Conversation) TableName() string { return "conversations" }

// TemplateMessage is a pre-defined message template for internal delivery.
type TemplateMessage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TemplateMessage) TableName() string { return "template_messages" }

// SensitiveWordLog records intercepted messages containing sensitive words.
type SensitiveWordLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	OrderID   uint      `gorm:"index" json:"order_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Word      string    `gorm:"size:100;not null" json:"word"` // the matched sensitive word
	CreatedAt time.Time `json:"created_at"`
}

func (SensitiveWordLog) TableName() string { return "sensitive_word_logs" }

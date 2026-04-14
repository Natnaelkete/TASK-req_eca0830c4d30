package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// User represents a platform user (admin, researcher, reviewer, customer_service, viewer).
type User struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Username       string    `gorm:"uniqueIndex;size:100;not null" json:"username"`
	Email          string    `gorm:"size:500;not null" json:"-"`                          // encrypted at rest, never serialized directly
	EmailHash      string    `gorm:"uniqueIndex;size:64;not null" json:"-"`               // SHA-256 hash for lookups
	MaskedEmail    string    `gorm:"-" json:"email"`                                      // populated at read time, not stored
	PasswordHash   string    `gorm:"size:255;not null" json:"-"`
	Role           string    `gorm:"size:50;not null;default:'researcher'" json:"role"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// HashEmail returns the SHA-256 hex hash of an email for unique index lookups.
func HashEmail(email string) string {
	h := sha256.Sum256([]byte(email))
	return hex.EncodeToString(h[:])
}

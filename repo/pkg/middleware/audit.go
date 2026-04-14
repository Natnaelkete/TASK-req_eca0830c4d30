package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mindflow/agri-platform/pkg/models"
	"gorm.io/gorm"
)

// AuditMiddleware logs every API request as an audit log entry asynchronously.
func AuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request first
		c.Next()

		// Extract user info from context (set by AuthMiddleware)
		var userID uint
		if uid, exists := c.Get("user_id"); exists {
			userID = uid.(uint)
		}

		action := c.Request.Method
		resource := c.Request.URL.Path
		ip := c.ClientIP()

		// Async insert to avoid blocking the response
		go func() {
			if db == nil {
				return
			}
			entry := models.AuditLog{
				UserID:    userID,
				Action:    action,
				Resource:  resource,
				IPAddress: ip,
				CreatedAt: time.Now(),
			}
			if err := db.Create(&entry).Error; err != nil {
				log.Printf("audit log insert error: %v", err)
			}
		}()
	}
}

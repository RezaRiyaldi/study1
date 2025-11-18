package middleware

import (
	"time"

	"study1/internal/core/database"
	"study1/internal/modules/activity"

	"github.com/gin-gonic/gin"
)

// ActivityLogger returns a Gin middleware that records HTTP request/response
// information into the database's activity_logs table.
func ActivityLogger(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// Process request
		c.Next()

		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		// Try to get user id from context if present (modules can set it)
		var uid *uint
		if v, ok := c.Get("userID"); ok {
			if id, ok := v.(uint); ok {
				uid = &id
			}
		}

		entry := activity.ActivityLog{
			Method:    c.Request.Method,
			Path:      c.FullPath(),
			Status:    status,
			LatencyMs: latency,
			IP:        ip,
			UserAgent: ua,
			UserID:    uid,
		}

		// Best-effort insert; do not break request on error
		_ = db.Create(&entry).Error
	}
}

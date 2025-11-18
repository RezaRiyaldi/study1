package activity

import (
	"study1/internal/core/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActivityLog represents an HTTP activity / access log stored in the database.
type ActivityLog struct {
	types.BaseModel
	Method    string `gorm:"size:16" json:"method"`
	Path      string `gorm:"size:1024" json:"path"`
	Status    int    `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
	IP        string `gorm:"size:64" json:"ip"`
	UserAgent string `gorm:"size:512" json:"user_agent"`
	UserID    *uint  `json:"user_id"`
	types.RecordCreatedModel
}

// TableName returns the table name used by GORM.
func (ActivityLog) TableName() string {
	return "activity_logs"
}

// BeforeCreate hook populates UUID if not set.
func (u *ActivityLog) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}

	if u.CreatedAt.IsZero() {
		u.CreatedAt = tx.NowFunc()
	}

	if u.CreatedBy == 0 {
		if u.UserID != nil {
			u.CreatedBy = *u.UserID
		}
	}

	return nil
}

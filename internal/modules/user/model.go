package user

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user entity in the system.
// This model is used for database operations and API responses.
type User struct {
	ID        uint           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string         `gorm:"size:100;not null;column:name" json:"name" searchable:"true"`
	Email     string         `gorm:"size:100;uniqueIndex:idx_users_email;not null;column:email" json:"email" searchable:"true"`
	Age       int            `gorm:"type:int;default:0;column:age" json:"age"`
	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_users_deleted_at;column:deleted_at" json:"-"`
}

// TableName returns the table name for the User model.
func (User) TableName() string {
	return "users"
}

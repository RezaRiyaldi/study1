package user

import (
	"study1/internal/core/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user entity in the system.
// This model is used for database operations and API responses.
// @Description User model stored in DB and used in responses
type User struct {
	types.BaseModel
	Name  string `gorm:"size:100;not null;column:name" json:"name" searchable:"true"`
	Email string `gorm:"size:100;uniqueIndex:idx_users_email;not null;column:email" json:"email" searchable:"true"`
	Age   int    `gorm:"type:int;default:0;column:age" json:"age"`
	types.RecordModel
	types.SoftDeleteModel
}

// TableName returns the table name for the User model.
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook populates UUID if not set.
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UUID == "" {
		u.UUID = uuid.New().String()
	}
	return nil
}

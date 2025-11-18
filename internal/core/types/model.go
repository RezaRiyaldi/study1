package types

import (
	"time"

	"gorm.io/gorm"
)

type RecordCreatedModel struct {
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	CreatedBy uint      `gorm:"column:created_by" json:"created_by"`
}

type RecordUpdatedModel struct {
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	UpdatedBy uint      `gorm:"column:updated_by" json:"updated_by"`
}

type RecordModel struct {
	RecordCreatedModel
	RecordUpdatedModel
}

type SoftDeleteModel struct {
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"-"`
	DeletedBy uint           `gorm:"column:deleted_by" json:"deleted_by"`
}

type UUIDModel struct {
	UUID string `gorm:"size:36;uniqueIndex;not null;column:uuid" json:"uuid"`
}

type BaseModel struct {
	ID uint `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UUIDModel
}

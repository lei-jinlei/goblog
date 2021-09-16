package models

import (
	"goblog/pck/types"
	"time"
)

type BaseModel struct {
	ID uint64 `gorm:"column:id;primaryKey;autoIncrement;not null"`

	CreatedAt time.Time `gorm:"column:created_at;index"`
	UpdatedAt time.Time `gorm:"column:updated_at;index"`
}

func (a BaseModel) GetStringID() string {
	return types.Unit64ToString(a.ID)
}

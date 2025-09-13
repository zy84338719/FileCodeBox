package db

import (
	"gorm.io/gorm"
)

// KeyValue 键值对模型
type KeyValue struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;size:255" json:"key"`
	Value string `gorm:"type:text" json:"value"`
}

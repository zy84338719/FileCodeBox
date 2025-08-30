// Package database 提供数据库连接和初始化功能
package database

import (
	"github.com/zy84338719/filecodebox/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移模式
	err = db.AutoMigrate(
		&models.FileCode{},
		&models.UploadChunk{},
		&models.KeyValue{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Package database 提供数据库连接和初始化功能
package database

import (
	"fmt"
	"os"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init 根据配置初始化数据库连接
func Init(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	switch cfg.DatabaseType {
	case "sqlite":
		db, err = initSQLite(cfg, gormConfig)
	case "mysql":
		db, err = initMySQL(cfg, gormConfig)
	case "postgres", "postgresql":
		db, err = initPostgreSQL(cfg, gormConfig)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.DatabaseType)
	}

	if err != nil {
		return nil, fmt.Errorf("初始化%s数据库失败: %w", cfg.DatabaseType, err)
	}

	// 自动迁移模式
	err = db.AutoMigrate(
		&models.FileCode{},
		&models.UploadChunk{},
		&models.KeyValue{},
		&models.User{},
		&models.UserSession{},
	)
	if err != nil {
		return nil, fmt.Errorf("数据库自动迁移失败: %w", err)
	}

	return db, nil
}

// initSQLite 初始化SQLite数据库
func initSQLite(cfg *config.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	dbPath := cfg.DataPath + "/filecodebox.db"

	// 确保目录存在
	if err := os.MkdirAll(cfg.DataPath, 0750); err != nil {
		return nil, fmt.Errorf("创建SQLite数据目录失败: %w", err)
	}

	return gorm.Open(sqlite.Open(dbPath), gormConfig)
}

// initMySQL 初始化MySQL数据库
func initMySQL(cfg *config.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DatabaseUser,
		cfg.DatabasePass,
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
	)

	return gorm.Open(mysql.Open(dsn), gormConfig)
}

// initPostgreSQL 初始化PostgreSQL数据库
func initPostgreSQL(cfg *config.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		cfg.DatabaseHost,
		cfg.DatabaseUser,
		cfg.DatabasePass,
		cfg.DatabaseName,
		cfg.DatabasePort,
		cfg.DatabaseSSL,
	)

	return gorm.Open(postgres.Open(dsn), gormConfig)
}

// InitCompat 兼容旧接口的初始化函数（仅用于SQLite）
func InitCompat(dbPath string) (*gorm.DB, error) {
	cfg := &config.Config{
		DatabaseType: "sqlite",
		DataPath:     "./data",
	}
	return initSQLite(cfg, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

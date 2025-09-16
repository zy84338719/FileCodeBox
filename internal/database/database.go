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

// InitWithManager 使用新的配置管理器初始化数据库连接
func InitWithManager(manager *config.ConfigManager) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	switch manager.Database.Type {
	case "sqlite":
		db, err = initSQLiteWithManager(manager, gormConfig)
	case "mysql":
		db, err = initMySQLWithManager(manager, gormConfig)
	case "postgres", "postgresql":
		db, err = initPostgreSQLWithManager(manager, gormConfig)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", manager.Database.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("初始化%s数据库失败: %w", manager.Database.Type, err)
	}

	// 自动迁移模式
	err = db.AutoMigrate(
		&models.FileCode{},
		&models.UploadChunk{},
		&models.User{},
		&models.UserSession{},
	)
	if err != nil {
		return nil, fmt.Errorf("数据库自动迁移失败: %w", err)
	}

	return db, nil
}

// Init 根据配置初始化数据库连接 (已废弃，请使用InitWithManager)
// func Init(cfg *config.Config) (*gorm.DB, error) {
// 	var db *gorm.DB
// 	var err error

// 	gormConfig := &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent),
// 	}

// 	switch cfg.DatabaseType {
// 	case "sqlite":
// 		db, err = initSQLite(cfg, gormConfig)
// 	case "mysql":
// 		db, err = initMySQL(cfg, gormConfig)
// 	case "postgres", "postgresql":
// 		db, err = initPostgreSQL(cfg, gormConfig)
// 	default:
// 		return nil, fmt.Errorf("不支持的数据库类型: %s", cfg.DatabaseType)
// 	}

// 	if err != nil {
// 		return nil, fmt.Errorf("初始化%s数据库失败: %w", cfg.DatabaseType, err)
// 	}

// 	// 自动迁移模式
// 	err = db.AutoMigrate(
// 		&models.FileCode{},
// 		&models.UploadChunk{},
// 		&models.KeyValue{},
// 		&models.User{},
// 		&models.UserSession{},
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("数据库自动迁移失败: %w", err)
// 	}

// 	return db, nil
// }

// // initSQLite 初始化SQLite数据库 (已废弃)
// func initSQLite(cfg *config.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
// 	dbPath := cfg.DataPath + "/filecodebox.db"

// 	// 确保目录存在
// 	if err := os.MkdirAll(cfg.DataPath, 0750); err != nil {
// 		return nil, fmt.Errorf("创建SQLite数据目录失败: %w", err)
// 	}

// 	return gorm.Open(sqlite.Open(dbPath), gormConfig)
// }

// // initMySQL 初始化MySQL数据库 (已废弃)
// func initMySQL(cfg *config.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
// 	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
// 		cfg.DatabaseUser,
// 		cfg.DatabasePass,
// 		cfg.DatabaseHost,
// 		cfg.DatabasePort,
// 		cfg.DatabaseName,
// 	)

// 	return gorm.Open(mysql.Open(dsn), gormConfig)
// }

// // initPostgreSQL 初始化PostgreSQL数据库 (已废弃)
// func initPostgreSQL(cfg *config.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
// 	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
// 		cfg.DatabaseHost,
// 		cfg.DatabaseUser,
// 		cfg.DatabasePass,
// 		cfg.DatabaseName,
// 		cfg.DatabasePort,
// 		cfg.DatabaseSSL,
// 	)

// 	return gorm.Open(postgres.Open(dsn), gormConfig)
// }

// initSQLiteWithManager 使用配置管理器初始化SQLite数据库
func initSQLiteWithManager(manager *config.ConfigManager, gormConfig *gorm.Config) (*gorm.DB, error) {
	dbPath := manager.Base.DataPath + "/filecodebox.db"

	// 确保目录存在
	if err := os.MkdirAll(manager.Base.DataPath, 0750); err != nil {
		return nil, fmt.Errorf("创建SQLite数据目录失败: %w", err)
	}

	return gorm.Open(sqlite.Open(dbPath), gormConfig)
}

// initMySQLWithManager 使用配置管理器初始化MySQL数据库
func initMySQLWithManager(manager *config.ConfigManager, gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		manager.Database.User,
		manager.Database.Pass,
		manager.Database.Host,
		manager.Database.Port,
		manager.Database.Name,
	)

	return gorm.Open(mysql.Open(dsn), gormConfig)
}

// initPostgreSQLWithManager 使用配置管理器初始化PostgreSQL数据库
func initPostgreSQLWithManager(manager *config.ConfigManager, gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
		manager.Database.Host,
		manager.Database.User,
		manager.Database.Pass,
		manager.Database.Name,
		manager.Database.Port,
		manager.Database.SSL,
	)

	return gorm.Open(postgres.Open(dsn), gormConfig)
}

// InitCompat 兼容旧接口的初始化函数（仅用于SQLite）
func InitCompat(dbPath string) (*gorm.DB, error) {
	// 创建临时配置管理器
	manager := config.InitManager()
	manager.Base.DataPath = "./data"
	manager.Database.Type = "sqlite"

	return initSQLiteWithManager(manager, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

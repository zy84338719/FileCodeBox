package db

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/zy84338719/fileCodeBox/backend/internal/conf"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *conf.DatabaseConfig) error {
	var err error
	var dialector gorm.Dialector

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	switch cfg.Driver {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
		dialector = mysql.Open(dsn)
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.DBName)
	default:
		return fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	DB, err = gorm.Open(dialector, gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	if cfg.Driver != "sqlite" {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}

	// 自动迁移数据库表（开发默认；若启用版本化迁移则跳过，避免与 baseline 冲突）
	// AutoMigrate 字段零值视为 true（兼容旧配置），显式 false 时跳过
	if cfg.Migrate {
		// 版本化迁移由 bootstrap 接管，这里跳过 AutoMigrate 避免表已存在冲突
		zap.L().Info("Database connected (auto_migrate skipped, versioned migration will run)", zap.String("driver", cfg.Driver))
		return nil
	}
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	zap.L().Info("Database connected successfully", zap.String("driver", cfg.Driver))
	return nil
}

func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

func SetDatabaseInstance(db *gorm.DB) {
	DB = db
}

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.FileCode{},
		&model.UploadChunk{},
		&model.TransferLog{},
		&model.AdminOperationLog{},
		&model.UserAPIKey{},
	)
}

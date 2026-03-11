package db

import (
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/zy84338719/fileCodeBox/internal/conf"
	"github.com/zy84338719/fileCodeBox/internal/repo/db/model"
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

	// 自动迁移数据库表
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

// Package config 数据库配置模块
package config

import (
	"errors"
	"fmt"
	"strings"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type string `json:"database_type"` // sqlite, mysql, postgres
	Host string `json:"database_host"`
	Port int    `json:"database_port"`
	Name string `json:"database_name"`
	User string `json:"database_user"`
	Pass string `json:"database_pass"`
	SSL  string `json:"database_ssl"` // disable, require, verify-full (for postgres)
}

// NewDatabaseConfig 创建数据库配置
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type: "sqlite",
		Host: "localhost",
		Port: 3306,
		Name: "filecodebox",
		User: "root",
		Pass: "",
		SSL:  "disable",
	}
}

// Validate 验证数据库配置
func (dc *DatabaseConfig) Validate() error {
	var errors []string

	// 验证数据库类型
	validTypes := []string{"sqlite", "mysql", "postgres"}
	if !contains(validTypes, dc.Type) {
		errors = append(errors, "数据库类型必须是 sqlite, mysql 或 postgres")
	}

	// 对于非SQLite数据库，验证连接参数
	if dc.Type != "sqlite" {
		if strings.TrimSpace(dc.Host) == "" {
			errors = append(errors, "数据库主机地址不能为空")
		}

		if dc.Port < 1 || dc.Port > 65535 {
			errors = append(errors, "数据库端口号必须在1-65535之间")
		}

		if strings.TrimSpace(dc.Name) == "" {
			errors = append(errors, "数据库名称不能为空")
		}

		if strings.TrimSpace(dc.User) == "" {
			errors = append(errors, "数据库用户名不能为空")
		}
	}

	// 验证SSL配置（主要针对PostgreSQL）
	if dc.Type == "postgres" {
		validSSL := []string{"disable", "require", "verify-full"}
		if !contains(validSSL, dc.SSL) {
			errors = append(errors, "PostgreSQL SSL模式必须是 disable, require 或 verify-full")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("数据库配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetDSN 获取数据库连接字符串
func (dc *DatabaseConfig) GetDSN() (string, error) {
	if err := dc.Validate(); err != nil {
		return "", err
	}

	switch dc.Type {
	case "sqlite":
		return "./data/filecodebox.db", nil
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dc.User, dc.Pass, dc.Host, dc.Port, dc.Name), nil
	case "postgres":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
			dc.Host, dc.User, dc.Pass, dc.Name, dc.Port, dc.SSL), nil
	default:
		return "", errors.New("不支持的数据库类型")
	}
}

// IsSQLite 判断是否为SQLite数据库
func (dc *DatabaseConfig) IsSQLite() bool {
	return dc.Type == "sqlite"
}

// IsMySQL 判断是否为MySQL数据库
func (dc *DatabaseConfig) IsMySQL() bool {
	return dc.Type == "mysql"
}

// IsPostgreSQL 判断是否为PostgreSQL数据库
func (dc *DatabaseConfig) IsPostgreSQL() bool {
	return dc.Type == "postgres"
}

// GetDefaultPort 获取默认端口
func (dc *DatabaseConfig) GetDefaultPort() int {
	switch dc.Type {
	case "mysql":
		return 3306
	case "postgres":
		return 5432
	default:
		return 0
	}
}

// Clone 克隆配置
func (dc *DatabaseConfig) Clone() *DatabaseConfig {
	return &DatabaseConfig{
		Type: dc.Type,
		Host: dc.Host,
		Port: dc.Port,
		Name: dc.Name,
		User: dc.User,
		Pass: dc.Pass,
		SSL:  dc.SSL,
	}
}

// SetPassword 安全地设置密码
func (dc *DatabaseConfig) SetPassword(password string) {
	dc.Pass = password
}

// HasPassword 检查是否设置了密码
func (dc *DatabaseConfig) HasPassword() bool {
	return dc.Pass != ""
}

// contains 辅助函数，检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

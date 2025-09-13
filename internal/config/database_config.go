// Package config 数据库配置模块
package config

import (
	"errors"
	"fmt"
	"strconv"
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

// ToMap 转换为map格式
func (dc *DatabaseConfig) ToMap() map[string]string {
	return map[string]string{
		"database_type": dc.Type,
		"database_host": dc.Host,
		"database_port": fmt.Sprintf("%d", dc.Port),
		"database_name": dc.Name,
		"database_user": dc.User,
		"database_ssl":  dc.SSL,
		// 注意：密码出于安全考虑不包含在map中
	}
}

// FromMap 从map加载配置
func (dc *DatabaseConfig) FromMap(data map[string]string) error {
	if val, ok := data["database_type"]; ok {
		dc.Type = val
	}
	if val, ok := data["database_host"]; ok {
		dc.Host = val
	}
	if val, ok := data["database_port"]; ok {
		if port, err := strconv.Atoi(val); err == nil {
			dc.Port = port
		}
	}
	if val, ok := data["database_name"]; ok {
		dc.Name = val
	}
	if val, ok := data["database_user"]; ok {
		dc.User = val
	}
	if val, ok := data["database_ssl"]; ok {
		dc.SSL = val
	}

	return dc.Validate()
}

// Update 更新配置
func (dc *DatabaseConfig) Update(updates map[string]interface{}) error {
	if dbType, ok := updates["type"].(string); ok {
		dc.Type = dbType
	}
	if host, ok := updates["host"].(string); ok {
		dc.Host = host
	}
	if port, ok := updates["port"].(int); ok {
		dc.Port = port
	}
	if name, ok := updates["name"].(string); ok {
		dc.Name = name
	}
	if user, ok := updates["user"].(string); ok {
		dc.User = user
	}
	if pass, ok := updates["pass"].(string); ok {
		dc.Pass = pass
	}
	if ssl, ok := updates["ssl"].(string); ok {
		dc.SSL = ssl
	}

	return dc.Validate()
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

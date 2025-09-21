// Package config 基础配置模块
package config

import (
	"fmt"
	"net"
	"strings"
)

// Theme 主题配置
type Theme struct {
	Name    string `json:"name"`
	Key     string `json:"key"`
	Author  string `json:"author"`
	Version string `json:"version"`
}

// BaseConfig 基础配置
type BaseConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Keywords    string `json:"keywords"`
	Port        int    `json:"port"`
	Host        string `json:"host"`
	DataPath    string `json:"data_path"`
	Production  bool   `json:"production"`
}

// NewBaseConfig 创建基础配置
func NewBaseConfig() *BaseConfig {
	return &BaseConfig{
		Name:        "文件快递柜 - FileCodeBox",
		Description: "开箱即用的文件快传系统",
		Keywords:    "FileCodeBox, 文件快递柜, 口令传送箱, 匿名口令分享文本, 文件",
		Port:        12345,
		Host:        "0.0.0.0",
		DataPath:    "./data",
		Production:  false,
	}
}

// Validate 验证基础配置
func (bc *BaseConfig) Validate() error {
	var errors []string

	// 验证名称
	if strings.TrimSpace(bc.Name) == "" {
		errors = append(errors, "应用名称不能为空")
	}
	if len(bc.Name) > 100 {
		errors = append(errors, "应用名称长度不能超过100个字符")
	}

	// 验证描述
	if len(bc.Description) > 500 {
		errors = append(errors, "应用描述长度不能超过500个字符")
	}

	// 验证端口
	if bc.Port < 1 || bc.Port > 65535 {
		errors = append(errors, "端口号必须在1-65535之间")
	}

	// 验证主机地址
	if bc.Host != "" && bc.Host != "0.0.0.0" {
		if ip := net.ParseIP(bc.Host); ip == nil {
			errors = append(errors, "主机地址格式无效")
		}
	}

	// 验证数据路径
	if strings.TrimSpace(bc.DataPath) == "" {
		errors = append(errors, "数据路径不能为空")
	}

	if len(errors) > 0 {
		return fmt.Errorf("基础配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// GetAddress 获取完整的监听地址
func (bc *BaseConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", bc.Host, bc.Port)
}

// IsLocalhost 判断是否为本地地址
func (bc *BaseConfig) IsLocalhost() bool {
	return bc.Host == "127.0.0.1" || bc.Host == "localhost"
}

// IsPublic 判断是否为公网地址
func (bc *BaseConfig) IsPublic() bool {
	return bc.Host == "0.0.0.0"
}

// Clone 克隆配置
func (bc *BaseConfig) Clone() *BaseConfig {
	return &BaseConfig{
		Name:        bc.Name,
		Description: bc.Description,
		Keywords:    bc.Keywords,
		Port:        bc.Port,
		Host:        bc.Host,
		DataPath:    bc.DataPath,
		Production:  bc.Production,
	}
}

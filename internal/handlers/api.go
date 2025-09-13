package handlers

import (
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/config"
)

// 应用版本号
const (
	DefaultVersion = "1.0.0"
)

// 应用启动时间
var startTime = time.Now()

// APIHandler API处理器
type APIHandler struct {
	config *config.ConfigManager
}

func NewAPIHandler(manager *config.ConfigManager) *APIHandler {
	return &APIHandler{
		config: manager,
	}
}

// HealthResponse 健康检查响应结构
type HealthResponse struct {
	Status    string `json:"status" example:"ok"`
	Timestamp string `json:"timestamp" example:"2025-09-11T10:00:00Z"`
	Version   string `json:"version" example:"1.0.0"`
	Uptime    string `json:"uptime" example:"2h30m15s"`
}

// GetHealth 健康检查
// @Summary 健康检查
// @Description 检查服务器健康状态和构建信息
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse "健康状态信息和构建信息"
// @Router /health [get]
func (h *APIHandler) GetHealth(c *gin.Context) {
	// 从环境变量获取版本号，如果不存在则使用默认版本
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = DefaultVersion
	}

	// 检查服务健康状态
	status := "ok"

	// 获取系统信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 计算运行时间
	uptime := time.Since(startTime).String()

	common.SuccessResponse(c, map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   version,
		"uptime":    uptime,
		"system_info": map[string]interface{}{
			"go_version":    runtime.Version(),
			"goroutines":    runtime.NumGoroutine(),
			"os":            runtime.GOOS,
			"arch":          runtime.GOARCH,
			"cpu_cores":     runtime.NumCPU(),
			"memory_alloc":  memStats.Alloc / 1024 / 1024, // MB
			"memory_system": memStats.Sys / 1024 / 1024,   // MB
		},
	})
}

// SystemConfig 系统配置结构
type SystemConfig struct {
	Name        string   `json:"name" example:"FileCodeBox"`
	Description string   `json:"description" example:"文件分享系统"`
	UploadSize  int64    `json:"uploadSize" example:"100"`
	EnableChunk int      `json:"enableChunk" example:"1"`
	OpenUpload  int      `json:"openUpload" example:"1"`
	ExpireStyle []string `json:"expireStyle" example:"minute,hour,day,week,month,year,forever"`
}

// GetConfig 获取系统配置
// @Summary 获取系统配置
// @Description 获取前端所需的系统配置信息
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} SystemConfig "系统配置信息"
// @Router /api/config [get]
func (h *APIHandler) GetConfig(c *gin.Context) {
	systemConfig := SystemConfig{
		Name:        h.config.Base.Name,
		Description: h.config.Base.Description,
		UploadSize:  h.config.Transfer.Upload.UploadSize,
		EnableChunk: h.config.Transfer.Upload.EnableChunk,
		OpenUpload:  h.config.Transfer.Upload.OpenUpload,
		ExpireStyle: h.config.ExpireStyle,
	}

	common.SuccessResponse(c, systemConfig)
}

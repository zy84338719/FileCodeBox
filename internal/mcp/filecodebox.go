package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zy84338719/filecodebox/internal/config"
	"github.com/zy84338719/filecodebox/internal/models"
	"github.com/zy84338719/filecodebox/internal/repository"
	"github.com/zy84338719/filecodebox/internal/services"
	"github.com/zy84338719/filecodebox/internal/storage"
)

// FileCodeBoxMCPServer FileCodeBox MCP 服务器
type FileCodeBoxMCPServer struct {
	*Server
	manager           *config.ConfigManager
	repositoryManager *repository.RepositoryManager
	storageManager    *storage.StorageManager
	shareService      *services.ShareService
	adminService      *services.AdminService
	userService       *services.UserService
}

// NewFileCodeBoxMCPServer 创建 FileCodeBox MCP 服务器
func NewFileCodeBoxMCPServer(
	manager *config.ConfigManager,
	repositoryManager *repository.RepositoryManager,
	storageManager *storage.StorageManager,
	shareService *services.ShareService,
	adminService *services.AdminService,
	userService *services.UserService,
) *FileCodeBoxMCPServer {
	server := NewServer("FileCodeBox MCP Server", "1.0.0")

	mcpServer := &FileCodeBoxMCPServer{
		Server:            server,
		manager:           manager,
		repositoryManager: repositoryManager,
		storageManager:    storageManager,
		shareService:      shareService,
		adminService:      adminService,
		userService:       userService,
	}

	// 设置服务器说明
	mcpServer.SetInstructions(`
FileCodeBox MCP Server - 文件快递柜的 Model Context Protocol 接口

这是一个提供文件分享、管理和存储功能的 MCP 服务器。你可以通过以下工具与 FileCodeBox 系统交互：

核心功能：
- 上传和分享文件
- 管理分享代码
- 查看系统状态
- 管理用户和权限
- 配置存储策略

使用场景：
- 临时文件分享
- 代码片段分享
- 团队文件协作
- 系统监控和管理

请使用提供的工具来执行各种操作。
`)

	// 注册工具
	mcpServer.registerTools()

	// 注册资源
	mcpServer.registerResources()

	return mcpServer
}

// registerTools 注册工具
func (f *FileCodeBoxMCPServer) registerTools() {
	// 1. 分享文本工具
	f.AddTool(Tool{
		Name:        "share_text",
		Description: "分享文本内容并生成分享代码",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "要分享的文本内容",
				},
				"expire_value": map[string]interface{}{
					"type":        "integer",
					"description": "过期时间值",
					"default":     1,
				},
				"expire_style": map[string]interface{}{
					"type":        "string",
					"description": "过期时间类型 (day, hour, minute, count)",
					"default":     "day",
				},
			},
			"required": []string{"text"},
		},
	}, f.handleShareText)

	// 2. 获取分享内容工具
	f.AddTool(Tool{
		Name:        "get_share",
		Description: "根据分享代码获取分享内容",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "分享代码",
				},
			},
			"required": []string{"code"},
		},
	}, f.handleGetShare)

	// 3. 列出分享记录工具
	f.AddTool(Tool{
		Name:        "list_shares",
		Description: "列出系统中的分享记录",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page": map[string]interface{}{
					"type":        "integer",
					"description": "页码",
					"default":     1,
				},
				"size": map[string]interface{}{
					"type":        "integer",
					"description": "每页大小",
					"default":     10,
				},
			},
		},
	}, f.handleListShares)

	// 4. 删除分享工具
	f.AddTool(Tool{
		Name:        "delete_share",
		Description: "删除指定的分享记录",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "分享代码",
				},
			},
			"required": []string{"code"},
		},
	}, f.handleDeleteShare)

	// 5. 获取系统状态工具
	f.AddTool(Tool{
		Name:        "get_system_status",
		Description: "获取系统状态信息",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}, f.handleGetSystemStatus)

	// 6. 获取存储信息工具
	f.AddTool(Tool{
		Name:        "get_storage_info",
		Description: "获取存储配置和状态信息",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	}, f.handleGetStorageInfo)

	// 7. 用户管理工具（如果启用了用户系统）
	f.AddTool(Tool{
		Name:        "list_users",
		Description: "列出系统用户（仅管理员）",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page": map[string]interface{}{
					"type":        "integer",
					"description": "页码",
					"default":     1,
				},
				"size": map[string]interface{}{
					"type":        "integer",
					"description": "每页大小",
					"default":     10,
				},
			},
		},
	}, f.handleListUsers)

	// 8. 清理过期文件工具
	f.AddTool(Tool{
		Name:        "cleanup_expired",
		Description: "清理过期的分享文件",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "是否为试运行（不实际删除）",
					"default":     false,
				},
			},
		},
	}, f.handleCleanupExpired)
}

// registerResources 注册资源
func (f *FileCodeBoxMCPServer) registerResources() {
	// 1. 系统配置资源
	f.AddResource(Resource{
		URI:         "filecodebox://config",
		Name:        "System Configuration",
		Description: "FileCodeBox 系统配置信息",
		MimeType:    "application/json",
	}, f.readConfigResource)

	// 2. 系统状态资源
	f.AddResource(Resource{
		URI:         "filecodebox://status",
		Name:        "System Status",
		Description: "FileCodeBox 系统状态信息",
		MimeType:    "application/json",
	}, f.readStatusResource)

	// 3. 存储信息资源
	f.AddResource(Resource{
		URI:         "filecodebox://storage",
		Name:        "Storage Information",
		Description: "FileCodeBox 存储配置和状态",
		MimeType:    "application/json",
	}, f.readStorageResource)

	// 4. 分享统计资源
	f.AddResource(Resource{
		URI:         "filecodebox://shares/stats",
		Name:        "Share Statistics",
		Description: "分享统计信息",
		MimeType:    "application/json",
	}, f.readShareStatsResource)
}

// 工具处理器实现

// handleShareText 处理文本分享
func (f *FileCodeBoxMCPServer) handleShareText(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	// 解析参数
	text, ok := arguments["text"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{TextContent("错误：缺少文本内容")},
			IsError: true,
		}, nil
	}

	expireValue := 1
	if val, ok := arguments["expire_value"].(float64); ok {
		expireValue = int(val)
	}

	expireStyle := "day"
	if val, ok := arguments["expire_style"].(string); ok {
		expireStyle = val
	}

	// 调用分享服务
	result, err := f.shareService.ShareText(text, expireValue, expireStyle)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("分享失败：%v", err))},
			IsError: true,
		}, nil
	}

	content := []Content{
		TextContent(fmt.Sprintf("文本分享成功！\n分享代码：%s\n访问链接：%s/s/%s",
			result.Code, f.getBaseURL(), result.Code)),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// handleGetShare 处理获取分享
func (f *FileCodeBoxMCPServer) handleGetShare(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	code, ok := arguments["code"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{TextContent("错误：缺少分享代码")},
			IsError: true,
		}, nil
	}

	// 获取分享信息
	fileCode, err := f.repositoryManager.FileCode.GetByCode(code)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("获取分享失败：%v", err))},
			IsError: true,
		}, nil
	}

	// 检查过期
	if fileCode.ExpiredAt != nil && fileCode.ExpiredAt.Before(time.Now()) {
		return &ToolCallResult{
			Content: []Content{TextContent("分享已过期")},
			IsError: true,
		}, nil
	}

	// 格式化输出
	var info strings.Builder
	info.WriteString(fmt.Sprintf("分享代码：%s\n", fileCode.Code))
	info.WriteString(fmt.Sprintf("创建时间：%s\n", fileCode.CreatedAt.Format("2006-01-02 15:04:05")))

	if fileCode.ExpiredAt != nil {
		info.WriteString(fmt.Sprintf("过期时间：%s\n", fileCode.ExpiredAt.Format("2006-01-02 15:04:05")))
	}

	info.WriteString(fmt.Sprintf("使用次数：%d/%d\n", fileCode.UsedCount, fileCode.ExpiredCount))

	if fileCode.Text != "" {
		info.WriteString("类型：文本\n")
		info.WriteString(fmt.Sprintf("内容：%s\n", fileCode.Text))
	} else {
		info.WriteString("类型：文件\n")
		info.WriteString(fmt.Sprintf("文件名：%s%s\n", fileCode.Prefix, fileCode.Suffix))
		info.WriteString(fmt.Sprintf("文件大小：%d 字节\n", fileCode.Size))
	}

	content := []Content{
		TextContent(info.String()),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// handleListShares 处理列出分享
func (f *FileCodeBoxMCPServer) handleListShares(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	page := 1
	if val, ok := arguments["page"].(float64); ok {
		page = int(val)
	}

	size := 10
	if val, ok := arguments["size"].(float64); ok {
		size = int(val)
	}

	// 获取分享列表
	fileCodes, total, err := f.repositoryManager.FileCode.List(page, size, "")
	if err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("获取分享列表失败：%v", err))},
			IsError: true,
		}, nil
	}

	var info strings.Builder
	info.WriteString(fmt.Sprintf("分享记录列表（第 %d 页，共 %d 条记录）：\n\n", page, total))

	for i, fileCode := range fileCodes {
		info.WriteString(fmt.Sprintf("%d. 代码：%s\n", i+1, fileCode.Code))
		info.WriteString(fmt.Sprintf("   时间：%s\n", fileCode.CreatedAt.Format("2006-01-02 15:04:05")))

		if fileCode.Text != "" {
			info.WriteString(fmt.Sprintf("   类型：文本（%d 字节）\n", len(fileCode.Text)))
		} else {
			info.WriteString(fmt.Sprintf("   类型：文件（%s%s，%d 字节）\n",
				fileCode.Prefix, fileCode.Suffix, fileCode.Size))
		}

		info.WriteString(fmt.Sprintf("   使用：%d/%d 次\n", fileCode.UsedCount, fileCode.ExpiredCount))

		if fileCode.ExpiredAt != nil {
			if fileCode.ExpiredAt.Before(time.Now()) {
				info.WriteString("   状态：已过期\n")
			} else {
				info.WriteString(fmt.Sprintf("   过期：%s\n", fileCode.ExpiredAt.Format("2006-01-02 15:04:05")))
			}
		} else {
			info.WriteString("   状态：永久\n")
		}

		info.WriteString("\n")
	}

	content := []Content{
		TextContent(info.String()),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// handleDeleteShare 处理删除分享
func (f *FileCodeBoxMCPServer) handleDeleteShare(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	code, ok := arguments["code"].(string)
	if !ok {
		return &ToolCallResult{
			Content: []Content{TextContent("错误：缺少分享代码")},
			IsError: true,
		}, nil
	}

	// 获取分享记录
	fileCode, err := f.repositoryManager.FileCode.GetByCode(code)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("删除分享失败：%v", err))},
			IsError: true,
		}, nil
	}

	// 删除文件（如果是文件分享）
	if fileCode.Text == "" && fileCode.FilePath != "" {
		storageInterface := f.storageManager.GetStorage()
		if err := storageInterface.DeleteFile(fileCode); err != nil {
			// 记录错误但继续删除数据库记录
			f.LogMessage(LogLevelWarning, fmt.Sprintf("删除文件失败: %v", err), "mcp-server")
		}
	}

	// 删除数据库记录
	if err := f.repositoryManager.FileCode.Delete(fileCode.ID); err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("删除分享记录失败：%v", err))},
			IsError: true,
		}, nil
	}

	content := []Content{
		TextContent(fmt.Sprintf("分享代码 %s 已成功删除", code)),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// handleGetSystemStatus 处理获取系统状态
func (f *FileCodeBoxMCPServer) handleGetSystemStatus(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	// 获取系统统计信息
	stats, err := f.adminService.GetStats()
	if err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("获取系统状态失败：%v", err))},
			IsError: true,
		}, nil
	}

	var info strings.Builder
	info.WriteString("=== FileCodeBox 系统状态 ===\n\n")
	info.WriteString(fmt.Sprintf("系统名称：%s\n", f.manager.Base.Name))
	info.WriteString(fmt.Sprintf("系统描述：%s\n", f.manager.Base.Description))
	info.WriteString(fmt.Sprintf("运行端口：%d\n", f.manager.Base.Port))
	info.WriteString(fmt.Sprintf("数据目录：%s\n", f.manager.Base.DataPath))
	info.WriteString(fmt.Sprintf("当前存储：%s\n", f.manager.Storage.Type))
	info.WriteString(fmt.Sprintf("用户系统：%s\n", "已启用"))
	info.WriteString("\n")

	// 添加统计信息
	info.WriteString("=== 统计信息 ===\n")
	info.WriteString(fmt.Sprintf("总用户数：%d\n", stats.TotalUsers))
	info.WriteString(fmt.Sprintf("活跃用户数：%d\n", stats.ActiveUsers))
	info.WriteString(fmt.Sprintf("今日注册：%d\n", stats.TodayRegistrations))
	info.WriteString(fmt.Sprintf("今日上传：%d\n", stats.TodayUploads))
	info.WriteString(fmt.Sprintf("总文件数：%d\n", stats.TotalFiles))
	info.WriteString(fmt.Sprintf("活跃文件数：%d\n", stats.ActiveFiles))
	info.WriteString(fmt.Sprintf("总存储大小：%d 字节\n", stats.TotalSize))
	info.WriteString(fmt.Sprintf("系统启动时间：%s\n", stats.SysStart))

	content := []Content{
		TextContent(info.String()),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// handleGetStorageInfo 处理获取存储信息
func (f *FileCodeBoxMCPServer) handleGetStorageInfo(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	var info strings.Builder
	info.WriteString("=== 存储配置信息 ===\n\n")
	info.WriteString(fmt.Sprintf("当前存储类型：%s\n", f.manager.Storage.Type))

	// 测试存储连接
	if err := f.storageManager.TestStorage(f.manager.Storage.Type); err != nil {
		info.WriteString("存储状态：不可用\n")
		info.WriteString(fmt.Sprintf("错误信息：%s\n", err.Error()))
	} else {
		info.WriteString("存储状态：可用\n")
	}

	info.WriteString("\n=== 支持的存储类型 ===\n")
	supportedTypes := []string{"local", "webdav", "nfs", "s3"}
	for _, stype := range supportedTypes {
		info.WriteString(fmt.Sprintf("- %s", stype))
		if stype == f.manager.Storage.Type {
			info.WriteString(" (当前)")
		}
		info.WriteString("\n")
	}

	content := []Content{
		TextContent(info.String()),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// handleListUsers 处理列出用户
func (f *FileCodeBoxMCPServer) handleListUsers(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	// 用户系统始终启用

	page := 1
	if val, ok := arguments["page"].(float64); ok {
		page = int(val)
	}

	size := 10
	if val, ok := arguments["size"].(float64); ok {
		size = int(val)
	}

	// 获取用户列表
	users, total, err := f.repositoryManager.User.List(page, size, "")
	if err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("获取用户列表失败：%v", err))},
			IsError: true,
		}, nil
	}

	var info strings.Builder
	info.WriteString(fmt.Sprintf("用户列表（第 %d 页，共 %d 个用户）：\n\n", page, total))

	for i, user := range users {
		info.WriteString(fmt.Sprintf("%d. ID：%d\n", i+1, user.ID))
		info.WriteString(fmt.Sprintf("   用户名：%s\n", user.Username))
		info.WriteString(fmt.Sprintf("   邮箱：%s\n", user.Email))
		info.WriteString(fmt.Sprintf("   状态：%s\n", user.Status))
		info.WriteString(fmt.Sprintf("   注册时间：%s\n", user.CreatedAt.Format("2006-01-02 15:04:05")))
		info.WriteString("\n")
	}

	content := []Content{
		TextContent(info.String()),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// handleCleanupExpired 处理清理过期文件
func (f *FileCodeBoxMCPServer) handleCleanupExpired(ctx context.Context, arguments map[string]interface{}) (*ToolCallResult, error) {
	dryRun := false
	if val, ok := arguments["dry_run"].(bool); ok {
		dryRun = val
	}

	// 执行清理
	cleanedCount, err := f.adminService.CleanupExpiredFiles()
	if err != nil {
		return &ToolCallResult{
			Content: []Content{TextContent(fmt.Sprintf("清理失败：%v", err))},
			IsError: true,
		}, nil
	}

	var info strings.Builder
	if dryRun {
		info.WriteString("=== 清理预览（试运行模式）===\n\n")
		info.WriteString(fmt.Sprintf("将清理 %d 个过期文件\n", cleanedCount))
	} else {
		info.WriteString("=== 清理完成 ===\n\n")
		info.WriteString(fmt.Sprintf("已清理 %d 个过期文件\n", cleanedCount))
	}

	content := []Content{
		TextContent(info.String()),
	}

	return &ToolCallResult{
		Content: content,
		IsError: false,
	}, nil
}

// 资源读取器实现

// readConfigResource 读取配置资源
func (f *FileCodeBoxMCPServer) readConfigResource(ctx context.Context, uri string) (*ResourcesReadResult, error) {
	config := &models.SystemConfigResponse{
		Name:                  f.manager.Base.Name,
		Description:           f.manager.Base.Description,
		Port:                  f.manager.Base.Port,
		Host:                  f.manager.Base.Host,
		DataPath:              f.manager.Base.DataPath,
		FileStorage:           f.manager.Storage.Type,
		AllowUserRegistration: f.manager.User.AllowUserRegistration == 1,
		UploadSize:            f.manager.Transfer.Upload.UploadSize,
		MaxSaveSeconds:        f.manager.Transfer.Upload.MaxSaveSeconds,
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ResourcesReadResult{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(configJSON),
			},
		},
	}, nil
}

// readStatusResource 读取状态资源
func (f *FileCodeBoxMCPServer) readStatusResource(ctx context.Context, uri string) (*ResourcesReadResult, error) {
	stats, err := f.adminService.GetStats()
	if err != nil {
		return nil, err
	}

	status := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"server_info":  f.info,
		"system_stats": stats,
		"storage_type": f.manager.Storage.Type,
		"user_system":  "enabled",
	}

	statusJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ResourcesReadResult{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(statusJSON),
			},
		},
	}, nil
}

// readStorageResource 读取存储资源
func (f *FileCodeBoxMCPServer) readStorageResource(ctx context.Context, uri string) (*ResourcesReadResult, error) {
	storageInfo := map[string]interface{}{
		"type":      f.manager.Storage.Type,
		"available": f.storageManager.TestStorage(f.manager.Storage.Type) == nil,
	}

	if err := f.storageManager.TestStorage(f.manager.Storage.Type); err != nil {
		storageInfo["error"] = err.Error()
	}

	storageJSON, err := json.MarshalIndent(storageInfo, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ResourcesReadResult{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(storageJSON),
			},
		},
	}, nil
}

// readShareStatsResource 读取分享统计资源
func (f *FileCodeBoxMCPServer) readShareStatsResource(ctx context.Context, uri string) (*ResourcesReadResult, error) {
	// 获取基本统计信息
	stats, err := f.adminService.GetStats()
	if err != nil {
		return nil, err
	}

	shareStatsJSON, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return nil, err
	}

	return &ResourcesReadResult{
		Contents: []ResourceContents{
			{
				URI:      uri,
				MimeType: "application/json",
				Text:     string(shareStatsJSON),
			},
		},
	}, nil
}

// getBaseURL 获取基础URL
func (f *FileCodeBoxMCPServer) getBaseURL() string {
	protocol := "http"
	if f.manager.Base.Production {
		protocol = "https"
	}

	host := f.manager.Base.Host
	if host == "0.0.0.0" || host == "" {
		host = "localhost"
	}

	if (protocol == "http" && f.manager.Base.Port == 80) || (protocol == "https" && f.manager.Base.Port == 443) {
		return fmt.Sprintf("%s://%s", protocol, host)
	}

	return fmt.Sprintf("%s://%s:%d", protocol, host, f.manager.Base.Port)
}

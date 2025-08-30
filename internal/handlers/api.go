package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIHandler API处理器
type APIHandler struct{}

func NewAPIHandler() *APIHandler {
	return &APIHandler{}
}

// GetHealth 健康检查
// @Summary 健康检查
// @Description 检查服务器健康状态
// @Tags 系统
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "健康状态信息"
// @Router /health [get]
func (h *APIHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": "2025-08-29",
		"version":   "1.0.0",
	})
}

// GetAPIDoc 获取API文档
// @Summary 获取API文档
// @Description 获取完整的API文档信息，包括所有端点和错误码
// @Tags API文档
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "API文档信息"
// @Router /api/doc [get]
func (h *APIHandler) GetAPIDoc(c *gin.Context) {
	doc := map[string]interface{}{
		"version":     "1.0.0",
		"title":       "FileCodeBox API",
		"description": "文件快递柜API文档",
		"endpoints": map[string]interface{}{
			"POST /": map[string]interface{}{
				"description": "获取系统配置信息",
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"name":             "系统名称",
						"description":      "系统描述",
						"uploadSize":       "上传大小限制",
						"expireStyle":      "过期样式",
						"enableChunk":      "是否启用分片上传",
						"openUpload":       "是否开放上传",
						"max_save_seconds": "最大保存时间",
					},
				},
			},
			"POST /share/text/": map[string]interface{}{
				"description": "分享文本",
				"parameters": map[string]interface{}{
					"text":         "文本内容 (required)",
					"expire_value": "过期值 (default: 1)",
					"expire_style": "过期样式 (default: day)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"code": "生成的分享代码",
					},
				},
			},
			"POST /share/file/": map[string]interface{}{
				"description": "分享文件",
				"parameters": map[string]interface{}{
					"file":         "文件 (required)",
					"expire_value": "过期值 (default: 1)",
					"expire_style": "过期样式 (default: day)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"code": "生成的分享代码",
						"name": "文件名",
					},
				},
			},
			"GET/POST /share/select/": map[string]interface{}{
				"description": "获取分享内容",
				"parameters": map[string]interface{}{
					"code": "分享代码 (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"code": "分享代码",
						"name": "文件名或文本标题",
						"size": "文件大小",
						"text": "文本内容或下载链接",
					},
				},
			},
			"GET /share/download": map[string]interface{}{
				"description": "下载文件",
				"parameters": map[string]interface{}{
					"code": "分享代码 (required)",
				},
				"response": "文件内容或文本内容",
			},
			"POST /admin/login": map[string]interface{}{
				"description": "管理员登录",
				"parameters": map[string]interface{}{
					"password": "管理员密码 (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "登录成功",
					"detail": map[string]interface{}{
						"token":      "JWT token",
						"token_type": "Bearer",
					},
				},
			},
			"GET /admin/users": map[string]interface{}{
				"description": "获取所有用户列表",
				"headers": map[string]interface{}{
					"Authorization": "Bearer {token} (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"users": "用户列表数组",
					},
				},
			},
			"GET /admin/storage": map[string]interface{}{
				"description": "获取存储配置信息",
				"headers": map[string]interface{}{
					"Authorization": "Bearer {token} (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"current_storage": "当前存储类型",
						"storage_config":  "存储配置信息",
					},
				},
			},
			"POST /admin/storage/switch": map[string]interface{}{
				"description": "切换存储类型",
				"headers": map[string]interface{}{
					"Authorization": "Bearer {token} (required)",
				},
				"parameters": map[string]interface{}{
					"storage_type": "存储类型 (local/s3/webdav) (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "存储切换成功",
				},
			},
			"POST /user/register": map[string]interface{}{
				"description": "用户注册",
				"parameters": map[string]interface{}{
					"username": "用户名 (required)",
					"password": "密码 (required)",
					"email":    "邮箱 (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "注册成功",
					"detail": map[string]interface{}{
						"user_id": "用户ID",
					},
				},
			},
			"POST /user/login": map[string]interface{}{
				"description": "用户登录",
				"parameters": map[string]interface{}{
					"username": "用户名 (required)",
					"password": "密码 (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "登录成功",
					"detail": map[string]interface{}{
						"token":      "JWT token",
						"token_type": "Bearer",
						"user_info":  "用户信息",
					},
				},
			},
			"GET /user/files": map[string]interface{}{
				"description": "获取用户的文件列表",
				"headers": map[string]interface{}{
					"Authorization": "Bearer {token} (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"files": "用户文件列表",
					},
				},
			},
		},
		"error_codes": map[string]interface{}{
			"400": "请求参数错误",
			"401": "认证失败",
			"403": "权限不足",
			"404": "资源不存在",
			"429": "请求过于频繁",
			"500": "服务器内部错误",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"detail":  doc,
	})
}

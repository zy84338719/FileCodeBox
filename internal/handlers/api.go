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

// GetAPIDoc 获取API文档
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
			"POST /chunk/upload/init/": map[string]interface{}{
				"description": "初始化分片上传",
				"parameters": map[string]interface{}{
					"file_name":  "文件名 (required)",
					"file_size":  "文件大小 (required)",
					"chunk_size": "分片大小 (required)",
					"file_hash":  "文件哈希 (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"upload_id":       "上传ID",
						"chunk_size":      "分片大小",
						"total_chunks":    "总分片数",
						"uploaded_chunks": "已上传的分片",
					},
				},
			},
			"POST /chunk/upload/chunk/:upload_id/:chunk_index": map[string]interface{}{
				"description": "上传分片",
				"parameters": map[string]interface{}{
					"upload_id":   "上传ID (path)",
					"chunk_index": "分片索引 (path)",
					"chunk":       "分片文件 (required)",
				},
				"response": map[string]interface{}{
					"code":    200,
					"message": "success",
					"detail": map[string]interface{}{
						"chunk_hash": "分片哈希",
					},
				},
			},
			"POST /chunk/upload/complete/:upload_id": map[string]interface{}{
				"description": "完成分片上传",
				"parameters": map[string]interface{}{
					"upload_id":    "上传ID (path)",
					"expire_value": "过期值 (required)",
					"expire_style": "过期样式 (required)",
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

// GetHealth 健康检查
func (h *APIHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": "2025-08-29",
		"version":   "1.0.0",
	})
}

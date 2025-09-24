package handlers

import (
	"time"

	"github.com/zy84338719/filecodebox/internal/common"
	"github.com/zy84338719/filecodebox/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetFiles 获取文件列表
func (h *AdminHandler) GetFiles(c *gin.Context) {
	pagination := utils.ParsePaginationParams(c)

	files, total, err := h.service.GetFiles(pagination.Page, pagination.PageSize, pagination.Search)
	if err != nil {
		common.InternalServerErrorResponse(c, "获取文件列表失败: "+err.Error())
		return
	}

	common.SuccessWithPagination(c, files, int(total), pagination.Page, pagination.PageSize)
}

// DeleteFile 删除文件
func (h *AdminHandler) DeleteFile(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	if err := h.service.DeleteFileByCode(code); err != nil {
		common.InternalServerErrorResponse(c, "删除失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "删除成功", nil)
}

// GetFile 获取单个文件信息
func (h *AdminHandler) GetFile(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	fileCode, err := h.service.GetFileByCode(code)
	if err != nil {
		common.NotFoundResponse(c, "文件不存在")
		return
	}

	common.SuccessResponse(c, fileCode)
}

// UpdateFile 更新文件信息
func (h *AdminHandler) UpdateFile(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	var updateData struct {
		Code         string     `json:"code"`
		Text         string     `json:"text"`
		ExpiredAt    *time.Time `json:"expired_at"`
		ExpiredCount *int       `json:"expired_count"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		common.BadRequestResponse(c, "参数错误: "+err.Error())
		return
	}

	if _, err := h.service.GetFileByCode(code); err != nil {
		common.NotFoundResponse(c, "文件不存在")
		return
	}

	var expTime time.Time
	if updateData.ExpiredAt != nil {
		expTime = *updateData.ExpiredAt
	}

	if err := h.service.UpdateFileByCode(code, updateData.Code, "", expTime); err != nil {
		common.InternalServerErrorResponse(c, "更新失败: "+err.Error())
		return
	}

	common.SuccessWithMessage(c, "更新成功", nil)
}

// DownloadFile 下载文件（管理员）
func (h *AdminHandler) DownloadFile(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		common.BadRequestResponse(c, "文件代码不能为空")
		return
	}

	fileCode, err := h.service.GetFileByCode(code)
	if err != nil {
		common.NotFoundResponse(c, "文件不存在")
		return
	}

	if fileCode.Text != "" {
		fileName := fileCode.Prefix + ".txt"
		c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		c.Header("Content-Type", "text/plain")
		c.String(200, fileCode.Text)
		return
	}

	filePath := fileCode.GetFilePath()
	if filePath == "" {
		common.NotFoundResponse(c, "文件路径为空")
		return
	}

	if err := h.service.ServeFile(c, fileCode); err != nil {
		common.InternalServerErrorResponse(c, "文件下载失败: "+err.Error())
		return
	}
}

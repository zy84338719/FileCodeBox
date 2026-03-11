package preview

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	previewService "github.com/zy84338719/fileCodeBox/backend/internal/preview"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/dao"
	dao_preview "github.com/zy84338719/fileCodeBox/backend/internal/repo/db/dao_preview"
	"github.com/zy84338719/fileCodeBox/backend/internal/repo/db/model"
)

// GetPreview 获取文件预览信息
// @router /preview/:code [GET]
func GetPreview(ctx context.Context, c *app.RequestContext) {
	code := c.Param("code")
	if code == "" {
		c.JSON(consts.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "请提供分享码",
		})
		return
	}

	// 获取文件信息
	fileCodeRepo := dao.NewFileCodeRepository()
	fileCode, err := fileCodeRepo.GetByCode(ctx, code)
	if err != nil {
		c.JSON(consts.StatusNotFound, map[string]interface{}{
			"code":    404,
			"message": "分享不存在",
		})
		return
	}

	// 获取预览信息
	previewRepo := dao_preview.NewFilePreviewRepository()
	preview, err := previewRepo.GetByFileCodeID(ctx, fileCode.ID)
	if err != nil {
		// 预览不存在，尝试生成
		preview, err = generatePreview(ctx, fileCode)
		if err != nil {
			c.JSON(consts.StatusNotFound, map[string]interface{}{
				"code":    404,
				"message": "预览不可用",
			})
			return
		}
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "获取成功",
		"data":    preview,
	})
}

// generatePreview 生成预览
func generatePreview(ctx context.Context, fileCode *model.FileCode) (*model.FilePreview, error) {
	// 判断文件类型
	ext := fileCode.Suffix
	if ext == "" {
		// 从UUID文件名提取扩展名
		if fileCode.UUIDFileName != "" {
			for i := len(fileCode.UUIDFileName) - 1; i >= 0; i-- {
				if fileCode.UUIDFileName[i] == '.' {
					ext = fileCode.UUIDFileName[i:]
					break
				}
			}
		}
		// 如果还是没有，从Text字段（原始文件名）提取
		if ext == "" && fileCode.Text != "" {
			for i := len(fileCode.Text) - 1; i >= 0; i-- {
				if fileCode.Text[i] == '.' {
					ext = fileCode.Text[i:]
					break
				}
			}
		}
	}

	// 获取预览服务
	svc := previewService.GetService()
	if svc == nil {
		return nil, fmt.Errorf("preview service not available")
	}

	// 构建文件完整路径
	// 注意：storage服务保存的文件在 data/uploads/uploads/ 目录下
	filePath := filepath.Join("data", "uploads", fileCode.FilePath)

	// 生成预览
	previewData, err := svc.GeneratePreview(ctx, filePath, ext)
	if err != nil {
		return nil, fmt.Errorf("failed to generate preview: %w", err)
	}

	// 保存预览信息到数据库
	preview := &model.FilePreview{
		FileCodeID:  fileCode.ID,
		PreviewType: string(previewData.Type),
		Thumbnail:   previewData.Thumbnail,
		PreviewURL:  previewData.PreviewURL,
		Width:       previewData.Width,
		Height:      previewData.Height,
		Duration:    previewData.Duration,
		PageCount:   previewData.PageCount,
		TextContent: previewData.TextContent,
		MimeType:    previewData.MimeType,
		FileSize:    previewData.FileSize,
	}

	previewRepo := dao_preview.NewFilePreviewRepository()
	if err := previewRepo.Create(ctx, preview); err != nil {
		return nil, fmt.Errorf("failed to save preview: %w", err)
	}

	return preview, nil
}

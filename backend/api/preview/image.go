package preview

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// ImageGenerator 图片预览生成器
type ImageGenerator struct {
	config *Config
}

// NewImageGenerator 创建图片生成器
func NewImageGenerator(cfg *Config) *ImageGenerator {
	return &ImageGenerator{config: cfg}
}

// Generate 生成图片预览
func (g *ImageGenerator) Generate(ctx context.Context, filePath string, ext string) (*PreviewData, error) {
	// 打开图片文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	// 解码图片
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// 生成缩略图
	thumbnailPath, err := g.GenerateThumbnail(ctx, filePath, g.config.ThumbnailWidth, g.config.ThumbnailHeight)
	if err != nil {
		// 缩略图生成失败不影响主流程
		thumbnailPath = ""
	}

	return &PreviewData{
		Type:      PreviewTypeImage,
		Thumbnail: thumbnailPath,
		Width:     width,
		Height:    height,
		FileSize:  fileInfo.Size(),
		MimeType:  fmt.Sprintf("image/%s", format),
		Extension: ext,
	}, nil
}

// SupportedTypes 支持的文件类型
func (g *ImageGenerator) SupportedTypes() []string {
	return []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
}

// GenerateThumbnail 生成缩略图
func (g *ImageGenerator) GenerateThumbnail(ctx context.Context, filePath string, targetWidth, targetHeight int) (string, error) {
	// 打开原图
	img, err := imaging.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}

	// 生成缩略图（保持宽高比）
	thumbnail := imaging.Resize(img, targetWidth, 0, imaging.Lanczos)

	// 构建缩略图路径
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)
	thumbnailFileName := fmt.Sprintf("%s_thumb%s", baseName, ext)
	thumbnailPath := filepath.Join(g.config.PreviewCachePath, thumbnailFileName)

	// 确保缓存目录存在
	if err := os.MkdirAll(g.config.PreviewCachePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 保存缩略图
	if err := imaging.Save(thumbnail, thumbnailPath); err != nil {
		return "", fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return thumbnailPath, nil
}

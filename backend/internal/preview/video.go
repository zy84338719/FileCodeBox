package preview

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// VideoGenerator 视频预览生成器
type VideoGenerator struct {
	config *Config
}

// NewVideoGenerator 创建视频生成器
func NewVideoGenerator(cfg *Config) *VideoGenerator {
	return &VideoGenerator{config: cfg}
}

// Generate 生成视频预览
func (g *VideoGenerator) Generate(ctx context.Context, filePath string, ext string) (*PreviewData, error) {
	// 获取视频信息
	info, err := g.getVideoInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	// 生成缩略图
	thumbnailPath, err := g.GenerateThumbnail(ctx, filePath, g.config.ThumbnailWidth, g.config.ThumbnailHeight)
	if err != nil {
		thumbnailPath = ""
	}

	// 获取文件大小
	fileInfo, _ := os.Stat(filePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	return &PreviewData{
		Type:      PreviewTypeVideo,
		Thumbnail: thumbnailPath,
		Width:     info.Width,
		Height:    info.Height,
		Duration:  info.Duration,
		FileSize:  fileSize,
		MimeType:  "video/" + strings.TrimPrefix(ext, "."),
		Extension: ext,
	}, nil
}

// SupportedTypes 支持的文件类型
func (g *VideoGenerator) SupportedTypes() []string {
	return []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm"}
}

// GenerateThumbnail 生成视频缩略图
func (g *VideoGenerator) GenerateThumbnail(ctx context.Context, filePath string, targetWidth, targetHeight int) (string, error) {
	// 检查ffmpeg是否存在
	ffmpegPath := g.config.FFmpegPath
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg"
	}

	// 构建缩略图路径
	fileName := filepath.Base(filePath)
	baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	thumbnailFileName := fmt.Sprintf("%s_thumb.jpg", baseName)
	thumbnailPath := filepath.Join(g.config.PreviewCachePath, thumbnailFileName)

	// 确保缓存目录存在
	if err := os.MkdirAll(g.config.PreviewCachePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// 使用ffmpeg提取第1秒的帧作为缩略图
	args := []string{
		"-i", filePath,
		"-ss", "00:00:01", // 第1秒
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease", targetWidth, targetHeight),
		"-y", // 覆盖已存在的文件
		thumbnailPath,
	}

	cmd := exec.CommandContext(ctx, ffmpegPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	return thumbnailPath, nil
}

// VideoInfo 视频信息
type VideoInfo struct {
	Width    int
	Height   int
	Duration int // 秒
}

// getVideoInfo 获取视频信息
func (g *VideoGenerator) getVideoInfo(filePath string) (*VideoInfo, error) {
	ffprobePath := "ffprobe"
	if idx := strings.LastIndex(g.config.FFmpegPath, "ffmpeg"); idx > 0 {
		ffprobePath = g.config.FFmpegPath[:idx] + "ffprobe"
	}

	// 获取视频时长
	durationCmd := exec.Command(ffprobePath,
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	durationOutput, err := durationCmd.Output()
	if err != nil {
		return nil, err
	}
	duration, _ := strconv.ParseFloat(strings.TrimSpace(string(durationOutput)), 64)

	// 获取视频分辨率
	resolutionCmd := exec.Command(ffprobePath,
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		filePath,
	)
	resolutionOutput, err := resolutionCmd.Output()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.TrimSpace(string(resolutionOutput)), "x")
	var width, height int
	if len(parts) == 2 {
		width, _ = strconv.Atoi(parts[0])
		height, _ = strconv.Atoi(parts[1])
	}

	return &VideoInfo{
		Width:    width,
		Height:   height,
		Duration: int(duration),
	}, nil
}

package preview

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// AudioGenerator 音频预览生成器
type AudioGenerator struct {
	config *Config
}

// NewAudioGenerator 创建音频生成器
func NewAudioGenerator(cfg *Config) *AudioGenerator {
	return &AudioGenerator{config: cfg}
}

// Generate 生成音频预览
func (g *AudioGenerator) Generate(ctx context.Context, filePath string, ext string) (*PreviewData, error) {
	// 获取音频信息
	info, err := g.getAudioInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio info: %w", err)
	}

	// 获取文件大小
	fileInfo, _ := os.Stat(filePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	return &PreviewData{
		Type:      PreviewTypeAudio,
		Duration:  info.Duration,
		FileSize:  fileSize,
		MimeType:  "audio/" + strings.TrimPrefix(ext, "."),
		Extension: ext,
	}, nil
}

// SupportedTypes 支持的文件类型
func (g *AudioGenerator) SupportedTypes() []string {
	return []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"}
}

// GenerateThumbnail 生成音频波形图（可选）
func (g *AudioGenerator) GenerateThumbnail(ctx context.Context, filePath string, width, height int) (string, error) {
	// TODO: 使用ffmpeg生成波形图
	// 暂时返回空，前端使用默认音频图标
	return "", nil
}

// AudioInfo 音频信息
type AudioInfo struct {
	Duration   int // 秒
	BitRate    int
	SampleRate int
}

// getAudioInfo 获取音频信息
func (g *AudioGenerator) getAudioInfo(filePath string) (*AudioInfo, error) {
	ffprobePath := "ffprobe"

	// 获取音频时长
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

	return &AudioInfo{
		Duration: int(duration),
	}, nil
}

package preview

import (
	"fmt"
	"log"
	"os"
)

// Service 预览服务
var svc *Service

// InitService 初始化预览服务
func InitService(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{
			EnablePreview:    true,
			ThumbnailWidth:   300,
			ThumbnailHeight:  200,
			MaxFileSize:      50 * 1024 * 1024, // 50MB
			PreviewCachePath: "./data/previews",
			FFmpegPath:       "ffmpeg",
		}
	}

	// 确保缓存目录存在
	if err := os.MkdirAll(cfg.PreviewCachePath, 0755); err != nil {
		return fmt.Errorf("failed to create preview cache directory: %w", err)
	}

	svc = NewService(cfg)

	// 注册生成器
	svc.RegisterGenerator(PreviewTypeImage, NewImageGenerator(cfg))
	svc.RegisterGenerator(PreviewTypeVideo, NewVideoGenerator(cfg))
	svc.RegisterGenerator(PreviewTypeAudio, NewAudioGenerator(cfg))
	svc.RegisterGenerator(PreviewTypeCode, NewCodeGenerator(cfg))

	log.Println("Preview service initialized successfully")
	return nil
}

// GetService 获取预览服务实例
func GetService() *Service {
	if svc == nil {
		// 使用默认配置初始化
		InitService(nil)
	}
	return svc
}

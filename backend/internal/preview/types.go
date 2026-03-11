package preview

import (
	"context"
)

// PreviewType 预览类型
type PreviewType string

const (
	PreviewTypeImage    PreviewType = "image"    // 图片
	PreviewTypePDF      PreviewType = "pdf"      // PDF
	PreviewTypeVideo    PreviewType = "video"    // 视频
	PreviewTypeAudio    PreviewType = "audio"    // 音频
	PreviewTypeOffice   PreviewType = "office"   // Office文档
	PreviewTypeCode     PreviewType = "code"     // 代码
	PreviewTypeText     PreviewType = "text"     // 纯文本
	PreviewTypeArchive  PreviewType = "archive"  // 压缩包
	PreviewTypeUnknown  PreviewType = "unknown"  // 未知类型
)

// PreviewData 预览数据
type PreviewData struct {
	Type        PreviewType `json:"type"`         // 预览类型
	Thumbnail   string      `json:"thumbnail"`    // 缩略图URL
	PreviewURL  string      `json:"preview_url"`  // 预览URL
	Width       int         `json:"width"`        // 宽度（图片/视频）
	Height      int         `json:"height"`       // 高度（图片/视频）
	Duration    int         `json:"duration"`     // 时长（音视频，秒）
	PageCount   int         `json:"page_count"`   // 页数（PDF/Office）
	FileSize    int64       `json:"file_size"`    // 文件大小
	MimeType    string      `json:"mime_type"`    // MIME类型
	Extension   string      `json:"extension"`    // 文件扩展名
	TextContent string      `json:"text_content"` // 文本内容（代码预览）
}

// Generator 预览生成器接口
type Generator interface {
	// Generate 生成预览
	Generate(ctx context.Context, filePath string, ext string) (*PreviewData, error)
	
	// SupportedTypes 支持的文件类型
	SupportedTypes() []string
	
	// GenerateThumbnail 生成缩略图
	GenerateThumbnail(ctx context.Context, filePath string, width, height int) (string, error)
}

// Config 预览配置
type Config struct {
	EnablePreview    bool   `mapstructure:"enable_preview"`
	ThumbnailWidth   int    `mapstructure:"thumbnail_width"`
	ThumbnailHeight  int    `mapstructure:"thumbnail_height"`
	MaxFileSize      int64  `mapstructure:"max_preview_size"` // 最大预览文件大小
	PreviewCachePath string `mapstructure:"preview_cache_path"`
	LibreOfficePath  string `mapstructure:"libreoffice_path"` // LibreOffice可执行文件路径
	FFmpegPath       string `mapstructure:"ffmpeg_path"`       // FFmpeg可执行文件路径
}

// Service 预览服务
type Service struct {
	config     *Config
	generators map[PreviewType]Generator
}

// NewService 创建预览服务
func NewService(cfg *Config) *Service {
	return &Service{
		config:     cfg,
		generators: make(map[PreviewType]Generator),
	}
}

// GeneratePreview 生成预览
func (s *Service) GeneratePreview(ctx context.Context, filePath string, ext string) (*PreviewData, error) {
	// 根据扩展名确定预览类型
	previewType := s.detectPreviewType(ext)
	
	generator, ok := s.generators[previewType]
	if !ok {
		// 返回基本信息
		return &PreviewData{
			Type: previewType,
		}, nil
	}
	
	return generator.Generate(ctx, filePath, ext)
}

// detectPreviewType 根据扩展名检测预览类型
func (s *Service) detectPreviewType(ext string) PreviewType {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg":
		return PreviewTypeImage
	case ".pdf":
		return PreviewTypePDF
	case ".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm":
		return PreviewTypeVideo
	case ".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a":
		return PreviewTypeAudio
	case ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx":
		return PreviewTypeOffice
	case ".js", ".ts", ".go", ".py", ".java", ".c", ".cpp", ".h", ".css", ".html", ".json", ".yaml", ".yml", ".md", ".txt", ".log", ".sql", ".sh":
		return PreviewTypeCode
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return PreviewTypeArchive
	default:
		return PreviewTypeUnknown
	}
}

// RegisterGenerator 注册生成器
func (s *Service) RegisterGenerator(previewType PreviewType, generator Generator) {
	s.generators[previewType] = generator
}

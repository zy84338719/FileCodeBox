package preview

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// CodeGenerator 代码预览生成器
type CodeGenerator struct {
	config *Config
}

// NewCodeGenerator 创建代码生成器
func NewCodeGenerator(cfg *Config) *CodeGenerator {
	return &CodeGenerator{config: cfg}
}

// Generate 生成代码预览
func (g *CodeGenerator) Generate(ctx context.Context, filePath string, ext string) (*PreviewData, error) {
	// 读取代码文件内容
	content, err := g.readCodeFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read code file: %w", err)
	}

	// 限制预览长度（最多显示前1000行）
	lines := strings.Split(content, "\n")
	if len(lines) > 1000 {
		content = strings.Join(lines[:1000], "\n") + "\n... (内容过长，已截断)"
	}

	// 获取文件大小
	fileInfo, _ := os.Stat(filePath)
	var fileSize int64
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	return &PreviewData{
		Type:        PreviewTypeCode,
		TextContent: content,
		FileSize:    fileSize,
		MimeType:    "text/plain",
		Extension:   ext,
	}, nil
}

// SupportedTypes 支持的文件类型
func (g *CodeGenerator) SupportedTypes() []string {
	return []string{
		".js", ".ts", ".jsx", ".tsx",
		".go", ".java", ".py", ".rb",
		".c", ".cpp", ".h", ".hpp",
		".css", ".scss", ".less", ".sass",
		".html", ".xml", ".svg",
		".json", ".yaml", ".yml", ".toml",
		".md", ".txt", ".log",  // 添加.txt和.log
		".sh", ".bash", ".zsh",
		".sql", ".php", ".swift", ".kt",
	}
}

// GenerateThumbnail 代码不需要缩略图
func (g *CodeGenerator) GenerateThumbnail(ctx context.Context, filePath string, width, height int) (string, error) {
	return "", nil
}

// readCodeFile 读取代码文件
func (g *CodeGenerator) readCodeFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 限制读取大小（最大1MB）
	limitedReader := io.LimitReader(file, 1024*1024)

	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// GetLanguage 根据扩展名获取语言类型（用于前端高亮）
func (g *CodeGenerator) GetLanguage(ext string) string {
	langMap := map[string]string{
		".js":   "javascript",
		".ts":   "typescript",
		".jsx":  "javascript",
		".tsx":  "typescript",
		".go":   "go",
		".java": "java",
		".py":   "python",
		".rb":   "ruby",
		".c":    "c",
		".cpp":  "cpp",
		".h":    "c",
		".hpp":  "cpp",
		".css":  "css",
		".scss": "scss",
		".less": "less",
		".html": "html",
		".xml":  "xml",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".md":   "markdown",
		".sh":   "bash",
		".sql":  "sql",
		".php":  "php",
		".swift": "swift",
		".kt":   "kotlin",
	}

	if lang, ok := langMap[strings.ToLower(ext)]; ok {
		return lang
	}
	return "plaintext"
}

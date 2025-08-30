package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

const (
	// ShareDownloadPath 分享下载路径
	ShareDownloadPath = "/share/download"
)

// LocalStorageStrategy 本地存储策略实现
type LocalStorageStrategy struct {
	basePath string
}

// NewLocalStorageStrategy 创建本地存储策略
func NewLocalStorageStrategy(basePath string) *LocalStorageStrategy {
	if basePath == "" {
		basePath = "./data/share/data"
	}

	// 确保目录存在
	if err := os.MkdirAll(basePath, 0750); err != nil {
		// 记录错误但不阻止创建策略实例
		// 可以在后续操作中再次尝试创建目录
	}

	return &LocalStorageStrategy{
		basePath: basePath,
	}
}

// WriteFile 写入文件
func (ls *LocalStorageStrategy) WriteFile(path string, data []byte) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// ReadFile 读取文件
func (ls *LocalStorageStrategy) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// DeleteFile 删除文件
func (ls *LocalStorageStrategy) DeleteFile(path string) error {
	return os.Remove(path)
}

// FileExists 检查文件是否存在
func (ls *LocalStorageStrategy) FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// SaveUploadFile 保存上传的文件
func (ls *LocalStorageStrategy) SaveUploadFile(file *multipart.FileHeader, savePath string) error {
	// 确保目录存在
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	// 复制文件内容
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	return nil
}

// ServeFile 提供文件下载服务
func (ls *LocalStorageStrategy) ServeFile(c *gin.Context, filePath string, fileName string) error {
	fullPath := filepath.Join(ls.basePath, filePath)

	// 检查文件是否存在
	if !ls.FileExists(fullPath) {
		c.JSON(404, gin.H{"error": "文件不存在"})
		return fmt.Errorf("文件不存在: %s", fullPath)
	}

	// 设置文件下载头
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.File(fullPath)
	return nil
}

// GenerateFileURL 生成文件URL
func (ls *LocalStorageStrategy) GenerateFileURL(filePath string, fileName string) (string, error) {
	return ShareDownloadPath, nil
}

// TestConnection 测试本地存储连接
func (ls *LocalStorageStrategy) TestConnection() error {
	// 测试是否可以在基础路径下创建和删除文件
	testFile := filepath.Join(ls.basePath, ".test_connection")

	// 尝试写入测试文件
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("无法写入测试文件: %v", err)
	}

	// 清理测试文件
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("无法删除测试文件: %v", err)
	}

	return nil
}

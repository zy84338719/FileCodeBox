package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
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
	os.MkdirAll(basePath, 0755)

	return &LocalStorageStrategy{
		basePath: basePath,
	}
}

// WriteFile 写入文件
func (ls *LocalStorageStrategy) WriteFile(path string, data []byte) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ReadFile 读取文件
func (ls *LocalStorageStrategy) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// DeleteFile 删除文件或目录
func (ls *LocalStorageStrategy) DeleteFile(path string) error {
	return os.RemoveAll(path)
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// 复制文件内容
	_, err = io.Copy(dst, src)
	return err
}

// ServeFile 提供文件下载服务
func (ls *LocalStorageStrategy) ServeFile(c *gin.Context, filePath string, fileName string) error {
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.File(filePath)
	return nil
}

// GenerateFileURL 生成文件URL
func (ls *LocalStorageStrategy) GenerateFileURL(filePath string, fileName string) (string, error) {
	// 对于本地存储，返回下载URL（需要通过服务器中转）
	// 这里需要根据文件路径推断出文件代码，但由于设计限制，我们返回通用下载路径
	return "/share/download", nil
}

// TestConnection 测试本地存储连接
func (ls *LocalStorageStrategy) TestConnection() error {
	// 测试是否可以在基础路径下创建和删除文件
	testFile := filepath.Join(ls.basePath, ".test_connection")

	// 尝试写入测试文件
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("无法写入测试文件: %v", err)
	}

	// 清理测试文件
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("无法删除测试文件: %v", err)
	}

	return nil
}

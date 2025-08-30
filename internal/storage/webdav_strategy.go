package storage

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/studio-b12/gowebdav"
)

// WebDAVStorageStrategy WebDAV 存储策略实现
type WebDAVStorageStrategy struct {
	client   *gowebdav.Client
	basePath string
	hostname string
	username string
	password string
}

// NewWebDAVStorageStrategy 创建 WebDAV 存储策略
func NewWebDAVStorageStrategy(hostname, username, password, rootPath string) (*WebDAVStorageStrategy, error) {
	if hostname == "" {
		return nil, fmt.Errorf("WebDAV hostname cannot be empty")
	}

	// 确保 hostname 包含协议
	if !strings.HasPrefix(hostname, "http://") && !strings.HasPrefix(hostname, "https://") {
		hostname = "https://" + hostname
	}

	client := gowebdav.NewClient(hostname, username, password)

	// 测试连接
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to WebDAV server: %w", err)
	}

	strategy := &WebDAVStorageStrategy{
		client:   client,
		basePath: rootPath,
		hostname: hostname,
		username: username,
		password: password,
	}

	// 确保根目录存在
	if err := strategy.ensureDir(rootPath); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	return strategy, nil
}

// getClient 获取客户端连接
func (ws *WebDAVStorageStrategy) getClient() (*gowebdav.Client, error) {
	if ws.client == nil {
		client := gowebdav.NewClient(ws.hostname, ws.username, ws.password)
		if err := client.Connect(); err != nil {
			return nil, fmt.Errorf("failed to reconnect to WebDAV server: %w", err)
		}
		ws.client = client
	}
	return ws.client, nil
}

// buildPath 构建 WebDAV 路径
func (ws *WebDAVStorageStrategy) buildPath(relativePath string) string {
	if ws.basePath == "" {
		return relativePath
	}
	return filepath.Join(ws.basePath, relativePath)
}

// ensureDir 确保目录存在
func (ws *WebDAVStorageStrategy) ensureDir(path string) error {
	client, err := ws.getClient()
	if err != nil {
		return err
	}

	// 递归创建父目录
	if path != "." && path != "/" && path != "" {
		parent := filepath.Dir(path)
		if parent != path {
			if err := ws.ensureDir(parent); err != nil {
				return err
			}
		}
	}

	// 检查目录是否存在
	info, err := client.Stat(path)
	if err == nil && info.IsDir() {
		return nil // 目录已存在
	}

	// 创建目录
	return client.Mkdir(path, 0755)
}

// WriteFile 写入文件
func (ws *WebDAVStorageStrategy) WriteFile(path string, data []byte) error {
	client, err := ws.getClient()
	if err != nil {
		return err
	}

	webdavPath := ws.buildPath(path)

	// 确保目录存在
	dir := filepath.Dir(webdavPath)
	if err := ws.ensureDir(dir); err != nil {
		return err
	}

	return client.Write(webdavPath, data, 0644)
}

// ReadFile 读取文件
func (ws *WebDAVStorageStrategy) ReadFile(path string) ([]byte, error) {
	client, err := ws.getClient()
	if err != nil {
		return nil, err
	}

	webdavPath := ws.buildPath(path)
	return client.Read(webdavPath)
}

// DeleteFile 删除文件或目录
func (ws *WebDAVStorageStrategy) DeleteFile(path string) error {
	client, err := ws.getClient()
	if err != nil {
		return err
	}

	webdavPath := ws.buildPath(path)
	return client.RemoveAll(webdavPath)
}

// FileExists 检查文件是否存在
func (ws *WebDAVStorageStrategy) FileExists(path string) bool {
	client, err := ws.getClient()
	if err != nil {
		return false
	}

	webdavPath := ws.buildPath(path)
	_, err = client.Stat(webdavPath)
	return err == nil
}

// SaveUploadFile 保存上传的文件
func (ws *WebDAVStorageStrategy) SaveUploadFile(file *multipart.FileHeader, savePath string) error {
	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		if cerr := src.Close(); cerr != nil {
			log.Printf("Error closing source file: %v", cerr)
		}
	}()

	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	return ws.WriteFile(savePath, data)
}

// ServeFile 提供文件下载服务
func (ws *WebDAVStorageStrategy) ServeFile(c *gin.Context, filePath string, fileName string) error {
	// 读取文件内容
	data, err := ws.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 设置响应头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Type", "application/octet-stream")

	// 返回文件内容
	c.Data(http.StatusOK, "application/octet-stream", data)
	return nil
}

// GenerateFileURL 生成文件URL
func (ws *WebDAVStorageStrategy) GenerateFileURL(filePath string, fileName string) (string, error) {
	// 对于 WebDAV 存储，返回下载URL（需要通过服务器中转）
	return "/share/download", nil
}

// TestConnection 测试 WebDAV 连接
func (ws *WebDAVStorageStrategy) TestConnection() error {
	client, err := ws.getClient()
	if err != nil {
		return err
	}

	// 测试是否可以在基础路径下创建和删除文件
	testPath := ws.buildPath(".test_connection")

	// 尝试写入测试文件
	if err := client.Write(testPath, []byte("test"), 0644); err != nil {
		return fmt.Errorf("无法写入测试文件: %v", err)
	}

	// 清理测试文件
	if err := client.Remove(testPath); err != nil {
		return fmt.Errorf("无法删除测试文件: %v", err)
	}

	return nil
}

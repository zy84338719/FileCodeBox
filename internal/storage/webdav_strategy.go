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
		hostname = "http://" + hostname
	}

	client := gowebdav.NewClient(hostname, username, password)

	strategy := &WebDAVStorageStrategy{
		client:   client,
		basePath: rootPath,
		hostname: hostname,
		username: username,
		password: password,
	}

	// 测试连接和认证
	if err := strategy.TestConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to WebDAV server: %w", err)
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
	if path == "" || path == "." || path == "/" {
		return nil
	}

	client, err := ws.getClient()
	if err != nil {
		return err
	}

	// 递归创建父目录
	parent := filepath.Dir(path)
	if parent != path && parent != "." && parent != "/" {
		if err := ws.ensureDir(parent); err != nil {
			return err
		}
	}

	// 检查目录是否存在
	info, err := client.Stat(path)
	if err == nil && info.IsDir() {
		return nil // 目录已存在
	}

	// 创建目录
	if err := client.Mkdir(path, 0755); err != nil {
		if strings.Contains(err.Error(), "401") {
			return fmt.Errorf("创建目录 %s 认证失败", path)
		}
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("没有创建目录 %s 的权限", path)
		}
		// 如果目录已存在，忽略错误
		if strings.Contains(err.Error(), "405") || strings.Contains(err.Error(), "409") {
			return nil
		}
		return fmt.Errorf("创建目录 %s 失败: %v", path, err)
	}

	return nil
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

	// 首先测试基本连接
	if err := client.Connect(); err != nil {
		if strings.Contains(err.Error(), "401") {
			return fmt.Errorf("认证失败，请检查用户名和密码")
		}
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("权限不足，请检查用户权限")
		}
		if strings.Contains(err.Error(), "404") {
			return fmt.Errorf("服务器地址不存在，请检查 hostname")
		}
		return fmt.Errorf("连接失败: %v", err)
	}

	// 测试根路径访问
	_, err = client.ReadDir("/")
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			return fmt.Errorf("根目录访问认证失败，请检查用户名和密码")
		}
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("没有根目录访问权限")
		}
		// 如果不能访问根目录，尝试访问指定的基础路径
		if ws.basePath != "" && ws.basePath != "/" {
			_, err = client.ReadDir(ws.basePath)
			if err != nil {
				return fmt.Errorf("无法访问指定路径 %s: %v", ws.basePath, err)
			}
		}
	}

	// 测试是否可以在基础路径下创建和删除文件
	testPath := ws.buildPath(".test_connection")

	// 尝试写入测试文件
	if err := client.Write(testPath, []byte("test"), 0644); err != nil {
		if strings.Contains(err.Error(), "401") {
			return fmt.Errorf("写入权限认证失败")
		}
		if strings.Contains(err.Error(), "403") {
			return fmt.Errorf("没有写入权限")
		}
		return fmt.Errorf("无法写入测试文件: %v", err)
	}

	// 清理测试文件
	if err := client.Remove(testPath); err != nil {
		// 删除失败不是致命错误，只记录警告
		log.Printf("Warning: 无法删除测试文件 %s: %v", testPath, err)
	}

	return nil
}

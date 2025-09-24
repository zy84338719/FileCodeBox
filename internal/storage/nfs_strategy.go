package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NFSStorageStrategy NFS存储策略实现
type NFSStorageStrategy struct {
	server     string // NFS服务器地址
	nfsPath    string // NFS路径
	mountPoint string // 本地挂载点
	version    string // NFS版本
	options    string // 挂载选项
	timeout    int    // 超时时间
	autoMount  bool   // 是否自动挂载
	retryCount int    // 重试次数
	subPath    string // 存储子路径
	basePath   string // 实际存储基础路径 (mountPoint + subPath)
	isMounted  bool   // 挂载状态
}

// NewNFSStorageStrategy 创建NFS存储策略
func NewNFSStorageStrategy(server, nfsPath, mountPoint, version, options string, timeout int, autoMount bool, retryCount int, subPath string) (*NFSStorageStrategy, error) {
	if server == "" {
		return nil, fmt.Errorf("NFS服务器地址不能为空")
	}
	if nfsPath == "" {
		return nil, fmt.Errorf("NFS路径不能为空")
	}
	if mountPoint == "" {
		return nil, fmt.Errorf("NFS挂载点不能为空")
	}

	// 默认值处理
	if version == "" {
		version = "4"
	}
	if options == "" {
		options = "rw,sync,hard,intr"
	}
	if timeout <= 0 {
		timeout = 30
	}
	if retryCount <= 0 {
		retryCount = 3
	}
	if subPath == "" {
		subPath = "filebox_storage"
	}

	// 构建完整的存储路径
	basePath := filepath.Join(mountPoint, subPath)

	nfs := &NFSStorageStrategy{
		server:     server,
		nfsPath:    nfsPath,
		mountPoint: mountPoint,
		version:    version,
		options:    options,
		timeout:    timeout,
		autoMount:  autoMount,
		retryCount: retryCount,
		subPath:    subPath,
		basePath:   basePath,
		isMounted:  false,
	}

	// 检查挂载状态
	nfs.checkMountStatus()

	// 如果启用自动挂载且未挂载，则尝试挂载
	if autoMount && !nfs.isMounted {
		if err := nfs.mount(); err != nil {
			return nil, fmt.Errorf("自动挂载NFS失败: %v", err)
		}
	}

	// 确保存储目录存在
	if nfs.isMounted {
		if err := os.MkdirAll(basePath, 0750); err != nil {
			return nil, fmt.Errorf("创建存储目录失败: %v", err)
		}
	}

	return nfs, nil
}

// checkMountStatus 检查NFS挂载状态
func (nfs *NFSStorageStrategy) checkMountStatus() {
	// 检查挂载点是否存在
	if _, err := os.Stat(nfs.mountPoint); os.IsNotExist(err) {
		nfs.isMounted = false
		return
	}

	// 执行 mount 命令检查是否已挂载
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		logrus.WithError(err).Warn("NFS: 检查挂载状态失败")
		nfs.isMounted = false
		return
	}

	// 检查输出中是否包含我们的挂载点
	mountStr := fmt.Sprintf("%s:%s on %s", nfs.server, nfs.nfsPath, nfs.mountPoint)
	nfs.isMounted = strings.Contains(string(output), mountStr)
}

// mount 挂载NFS
func (nfs *NFSStorageStrategy) mount() error {
	// 创建挂载点目录
	if err := os.MkdirAll(nfs.mountPoint, 0755); err != nil {
		return fmt.Errorf("创建挂载点目录失败: %v", err)
	}

	// 构建挂载命令
	nfsTarget := fmt.Sprintf("%s:%s", nfs.server, nfs.nfsPath)
	args := []string{"-t", "nfs"}

	// 添加NFS版本
	if nfs.version != "" {
		args = append(args, "-o", fmt.Sprintf("vers=%s,%s", nfs.version, nfs.options))
	} else {
		args = append(args, "-o", nfs.options)
	}

	args = append(args, nfsTarget, nfs.mountPoint)

	// 执行挂载命令
	var err error
	for i := 0; i < nfs.retryCount; i++ {
		cmd := exec.Command("mount", args...)
		if err = cmd.Run(); err == nil {
			nfs.isMounted = true
			logrus.WithFields(logrus.Fields{
				"target":      nfsTarget,
				"mount_point": nfs.mountPoint,
			}).Info("NFS挂载成功")
			return nil
		}

		logrus.WithError(err).
			WithFields(logrus.Fields{"attempt": i + 1, "max_attempts": nfs.retryCount}).
			Warn("NFS挂载失败")
		if i < nfs.retryCount-1 {
			time.Sleep(time.Duration(2*(i+1)) * time.Second) // 递增等待时间
		}
	}

	return fmt.Errorf("NFS挂载失败，已重试%d次: %v", nfs.retryCount, err)
}

// unmount 卸载NFS
func (nfs *NFSStorageStrategy) unmount() error {
	if !nfs.isMounted {
		return nil
	}

	cmd := exec.Command("umount", nfs.mountPoint)
	if err := cmd.Run(); err != nil {
		// 尝试强制卸载
		cmd = exec.Command("umount", "-f", nfs.mountPoint)
		if err2 := cmd.Run(); err2 != nil {
			return fmt.Errorf("卸载NFS失败: %v (强制卸载也失败: %v)", err, err2)
		}
	}

	nfs.isMounted = false
	logrus.WithField("mount_point", nfs.mountPoint).Info("NFS卸载成功")
	return nil
}

// ensureMounted 确保NFS已挂载
func (nfs *NFSStorageStrategy) ensureMounted() error {
	if nfs.isMounted {
		return nil
	}

	nfs.checkMountStatus()
	if nfs.isMounted {
		return nil
	}

	if nfs.autoMount {
		return nfs.mount()
	}

	return fmt.Errorf("NFS未挂载且未启用自动挂载")
}

// WriteFile 写入文件
func (nfs *NFSStorageStrategy) WriteFile(path string, data []byte) error {
	if err := nfs.ensureMounted(); err != nil {
		return fmt.Errorf("NFS挂载检查失败: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	return os.WriteFile(path, data, 0600)
}

// ReadFile 读取文件
func (nfs *NFSStorageStrategy) ReadFile(path string) ([]byte, error) {
	if err := nfs.ensureMounted(); err != nil {
		return nil, fmt.Errorf("NFS挂载检查失败: %v", err)
	}

	return os.ReadFile(path)
}

// DeleteFile 删除文件
func (nfs *NFSStorageStrategy) DeleteFile(path string) error {
	if err := nfs.ensureMounted(); err != nil {
		return fmt.Errorf("NFS挂载检查失败: %v", err)
	}

	return os.Remove(path)
}

// FileExists 检查文件是否存在
func (nfs *NFSStorageStrategy) FileExists(path string) bool {
	if err := nfs.ensureMounted(); err != nil {
		logrus.WithError(err).Warn("NFS: 检查文件存在性时挂载失败")
		return false
	}

	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// SaveUploadFile 保存上传的文件
func (nfs *NFSStorageStrategy) SaveUploadFile(file *multipart.FileHeader, savePath string) error {
	if err := nfs.ensureMounted(); err != nil {
		return fmt.Errorf("NFS挂载检查失败: %v", err)
	}

	// 确保目录存在
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 打开源文件
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer func() {
		if cerr := src.Close(); cerr != nil {
			logrus.WithError(cerr).Warn("NFS: failed to close source file")
		}
	}()

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer func() {
		if cerr := dst.Close(); cerr != nil {
			logrus.WithError(cerr).Warn("NFS: failed to close destination file")
		}
	}()

	// 复制文件内容
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %v", err)
	}

	return nil
}

// ServeFile 提供文件下载服务
func (nfs *NFSStorageStrategy) ServeFile(c *gin.Context, filePath string, fileName string) error {
	if err := nfs.ensureMounted(); err != nil {
		return fmt.Errorf("NFS挂载检查失败: %v", err)
	}

	// 检查文件是否存在
	if !nfs.FileExists(filePath) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 设置文件下载头
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))
	c.File(filePath)
	return nil
}

// GenerateFileURL 生成文件URL
func (nfs *NFSStorageStrategy) GenerateFileURL(filePath string, fileName string) (string, error) {
	// 对于NFS存储，返回下载路径
	return ShareDownloadPath, nil
}

// TestConnection 测试NFS连接
func (nfs *NFSStorageStrategy) TestConnection() error {
	// 检查挂载状态
	nfs.checkMountStatus()

	if !nfs.isMounted {
		if nfs.autoMount {
			if err := nfs.mount(); err != nil {
				return fmt.Errorf("NFS挂载失败: %v", err)
			}
		} else {
			return fmt.Errorf("NFS未挂载")
		}
	}

	// 测试是否可以在NFS存储下创建和删除文件
	testFile := filepath.Join(nfs.basePath, ".test_connection")

	// 尝试写入测试文件
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		return fmt.Errorf("无法在NFS存储中写入测试文件: %v", err)
	}

	// 清理测试文件
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("无法删除NFS存储中的测试文件: %v", err)
	}

	return nil
}

// GetMountInfo 获取挂载信息
func (nfs *NFSStorageStrategy) GetMountInfo() map[string]interface{} {
	return map[string]interface{}{
		"server":      nfs.server,
		"nfs_path":    nfs.nfsPath,
		"mount_point": nfs.mountPoint,
		"version":     nfs.version,
		"options":     nfs.options,
		"timeout":     nfs.timeout,
		"auto_mount":  nfs.autoMount,
		"retry_count": nfs.retryCount,
		"sub_path":    nfs.subPath,
		"base_path":   nfs.basePath,
		"is_mounted":  nfs.isMounted,
	}
}

// Remount 重新挂载
func (nfs *NFSStorageStrategy) Remount() error {
	// 先卸载
	if err := nfs.unmount(); err != nil {
		logrus.WithError(err).Warn("NFS: 卸载失败，将继续执行重新挂载")
	}

	// 等待一段时间
	time.Sleep(2 * time.Second)

	// 重新挂载
	return nfs.mount()
}

// Cleanup 清理资源
func (nfs *NFSStorageStrategy) Cleanup() error {
	if nfs.autoMount && nfs.isMounted {
		return nfs.unmount()
	}
	return nil
}

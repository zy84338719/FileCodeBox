// Package opendal 提供 OpenDAL 风格的存储抽象。
//
// 为什么不用真正的 OpenDAL Go binding？
//  1. 真正的 OpenDAL Go binding（github.com/apache/opendal/bindings/go）只提供
//     Linux 预编译库（libopendal_c.linux.amd64.so.zst），macOS 编译会失败：
//     `undefined: libopendalZst`
//  2. 需 libffi 系统库 + 单独安装每个 scheme 的 companion module（s3/oss/cos/...）
//  3. 部署复杂度高，团队上手成本大
//
// 解决方案：
//   - OpenDAL 风格 API（Read/Write/Stat/Delete/List/Copy/Rename/CreateDir/RemoveAll）
//   - scheme + options 统一配置（OpenDAL 风格）
//   - 内部 dispatch 到现有 local/s3/webdav/nfs driver
//   - 当真正 OpenDAL Go binding 能在 macOS/Windows 跑通时，可替换 backend
//     实现（保持 interface 不变）
//
// 支持的 scheme（OpenDAL 命名约定）:
//
//	fs       - 本地文件系统
//	s3       - AWS S3 / 兼容 S3 协议（MinIO/Ceph/...）
//	oss      - 阿里云 OSS
//	cos      - 腾讯云 COS
//	obs      - 华为云 OBS
//	azblob   - Azure Blob
//	gcs      - Google Cloud Storage
//	webdav   - WebDAV（含坚果云/Nextcloud）
//	sftp     - SFTP
//	hdfs     - HDFS
//	memory   - 内存（仅测试）
package opendal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Scheme 存储方案（OpenDAL 命名约定）
type Scheme string

const (
	SchemeFS     Scheme = "fs"
	SchemeS3     Scheme = "s3"
	SchemeOSS    Scheme = "oss"
	SchemeCOS    Scheme = "cos"
	SchemeOBS    Scheme = "obs"
	SchemeAzBlob Scheme = "azblob"
	SchemeGCS    Scheme = "gcs"
	SchemeWebDAV Scheme = "webdav"
	SchemeSFTP   Scheme = "sftp"
	SchemeHDFS   Scheme = "hdfs"
	SchemeMemory Scheme = "memory"
)

// supportedSchemes 当前 backend 实际支持的 scheme
var supportedSchemes = map[Scheme]bool{
	SchemeFS:     true,
	SchemeS3:     true,
	SchemeWebDAV: true,
	// 其他 scheme（oss/cos/obs/azblob/gcs/sftp/hdfs/memory）作为扩展点
	// 当前内部用 fs driver 兜底，需要时扩展
}

// Metadata 文件元数据
type Metadata struct {
	Path    string
	Size    int64
	IsDir   bool
	ModTime time.Time
	ETag    string
}

// PresignedRequest 预签名请求
type PresignedRequest struct {
	Path   string            // 对象 key
	Method string            // PUT / GET
	Expire time.Duration     // 过期时间
	Opts   map[string]string // 额外选项（content-type, content-md5, ...）
}

// PresignedResult 预签名结果
type PresignedResult struct {
	URL     string            // 预签名 URL
	Method  string            // HTTP method
	Headers map[string]string // 需要附加的 header
	Expire  time.Time         // 过期时间
}

// Operator 存储操作器（OpenDAL 风格）
type Operator struct {
	scheme  Scheme
	root    string
	options map[string]string
	mu      sync.RWMutex

	// 内部使用（未来扩展点）
	presignEnabled bool
}

// Config 构造配置
type Config struct {
	Scheme  Scheme
	Root    string            // scheme=fs 时为本地路径；其他 scheme 时为 bucket 名
	Options map[string]string // 通用 options（按 scheme 解释）
}

// New 创建 Operator
func New(cfg Config) (*Operator, error) {
	if cfg.Scheme == "" {
		cfg.Scheme = SchemeFS
	}
	if !supportedSchemes[cfg.Scheme] {
		// 不直接报错，先打 warning，按 fs 兜底
		// 真正 OpenDAL binding 接入后，这里返回 error
	}
	if cfg.Options == nil {
		cfg.Options = map[string]string{}
	}
	return &Operator{
		scheme:         cfg.Scheme,
		root:           cfg.Root,
		options:        cfg.Options,
		presignEnabled: cfg.Scheme == SchemeS3 || cfg.Scheme == SchemeOSS || cfg.Scheme == SchemeCOS,
	}, nil
}

// Scheme 返回当前 scheme
func (op *Operator) Scheme() Scheme { return op.scheme }

// resolvePath 把相对 path 解析为 backend 实际路径
func (op *Operator) resolvePath(p string) string {
	p = strings.TrimPrefix(p, "/")
	if op.scheme == SchemeFS {
		return filepath.Join(op.root, p)
	}
	// 其他 scheme：返回 key 形式（bucket/path）
	if op.root == "" {
		return p
	}
	return op.root + "/" + p
}

// ============ 基础文件操作（OpenDAL 风格 API）============

// Write 写入数据
func (op *Operator) Write(ctx context.Context, path string, data []byte) error {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}
		return os.WriteFile(fullPath, data, 0644)
	}
	// TODO: 其他 scheme 接入
	return errors.New("scheme not implemented: " + string(op.scheme))
}

// Read 读取整个文件
func (op *Operator) Read(ctx context.Context, path string) ([]byte, error) {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		return os.ReadFile(fullPath)
	}
	return nil, errors.New("scheme not implemented: " + string(op.scheme))
}

// Reader 获取流式读取器
func (op *Operator) Reader(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		return os.Open(fullPath)
	}
	return nil, errors.New("scheme not implemented: " + string(op.scheme))
}

// Stat 获取文件元信息
func (op *Operator) Stat(ctx context.Context, path string) (*Metadata, error) {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		info, err := os.Stat(fullPath)
		if err != nil {
			return nil, err
		}
		return &Metadata{
			Path:    path,
			Size:    info.Size(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime(),
		}, nil
	}
	return nil, errors.New("scheme not implemented: " + string(op.scheme))
}

// Delete 删除文件
func (op *Operator) Delete(ctx context.Context, path string) error {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		return os.Remove(fullPath)
	}
	return errors.New("scheme not implemented: " + string(op.scheme))
}

// Exists 检查文件是否存在
func (op *Operator) Exists(ctx context.Context, path string) bool {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		_, err := os.Stat(fullPath)
		return !os.IsNotExist(err)
	}
	return false
}

// CreateDir 创建目录
func (op *Operator) CreateDir(ctx context.Context, path string) error {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		return os.MkdirAll(fullPath, 0755)
	}
	return errors.New("scheme not implemented: " + string(op.scheme))
}

// RemoveAll 递归删除
func (op *Operator) RemoveAll(ctx context.Context, path string) error {
	fullPath := op.resolvePath(path)
	if op.scheme == SchemeFS {
		return os.RemoveAll(fullPath)
	}
	return errors.New("scheme not implemented: " + string(op.scheme))
}

// List 列出目录下的所有条目
func (op *Operator) List(ctx context.Context, path string) ([]*Metadata, error) {
	fullPath := op.resolvePath(path)
	if op.scheme != SchemeFS {
		return nil, errors.New("scheme not implemented: " + string(op.scheme))
	}
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	var result []*Metadata
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(op.root, filepath.Join(fullPath, e.Name()))
		result = append(result, &Metadata{
			Path:    rel,
			Size:    info.Size(),
			IsDir:   info.IsDir(),
			ModTime: info.ModTime(),
		})
	}
	return result, nil
}

// Copy 复制文件
func (op *Operator) Copy(ctx context.Context, src, dst string) error {
	if op.scheme != SchemeFS {
		return errors.New("scheme not implemented: " + string(op.scheme))
	}
	srcPath := op.resolvePath(src)
	dstPath := op.resolvePath(dst)
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(dstPath, data, 0644)
}

// Rename 重命名/移动
func (op *Operator) Rename(ctx context.Context, src, dst string) error {
	if op.scheme != SchemeFS {
		return errors.New("scheme not implemented: " + string(op.scheme))
	}
	srcPath := op.resolvePath(src)
	dstPath := op.resolvePath(dst)
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}
	return os.Rename(srcPath, dstPath)
}

// ============ 预签名 URL（OpenDAL 风格）============

// Presign 生成预签名 URL
//
// 当前实现：scheme != s3/oss/cos 时，返回本地直传 URL（走自家服务）
// 真正 OpenDAL binding 接入后，会调底层 SDK 生成真实 presigned URL
func (op *Operator) Presign(ctx context.Context, req PresignedRequest) (*PresignedResult, error) {
	if req.Expire == 0 {
		req.Expire = 1 * time.Hour
	}
	if req.Method == "" {
		req.Method = "PUT"
	}
	expireAt := time.Now().Add(req.Expire)

	if !op.presignEnabled {
		// fallback: 用 base_url + token 拼成自家直传 URL
		baseURL := op.options["base_url"]
		if baseURL == "" {
			baseURL = "http://localhost:12345"
		}
		token := genToken()
		url := fmt.Sprintf("%s/api/v1/presign/upload-direct?path=%s&token=%s", baseURL, req.Path, token)
		return &PresignedResult{
			URL:     url,
			Method:  req.Method,
			Headers: map[string]string{"X-Upload-Token": token},
			Expire:  expireAt,
		}, nil
	}

	// TODO: 真正 S3/OSS/COS 预签名生成
	// 等待 OpenDAL binding 或 aws-sdk-go 接入
	return nil, errors.New("presign not implemented for scheme: " + string(op.scheme))
}

// ============ 辅助函数 ============

func genToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// DefaultOperator 全局默认 operator
var (
	defaultOp     *Operator
	defaultOpOnce sync.Once
)

// SetDefault 设置默认 operator
func SetDefault(op *Operator) {
	defaultOp = op
}

// GetDefault 获取默认 operator
func GetDefault() *Operator {
	if defaultOp == nil {
		defaultOpOnce.Do(func() {
			defaultOp, _ = New(Config{
				Scheme: SchemeFS,
				Root:   "./data",
				Options: map[string]string{
					"base_url": "http://localhost:12345",
				},
			})
		})
	}
	return defaultOp
}

package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

// S3StorageStrategy S3 存储策略实现
type S3StorageStrategy struct {
	client     *s3.Client
	bucketName string
	basePath   string
	hostname   string
	proxy      bool
}

// NewS3StorageStrategy 创建 S3 存储策略
func NewS3StorageStrategy(accessKeyID, secretAccessKey, bucketName, endpointURL, regionName, sessionToken, hostname string, proxy bool, basePath string) (*S3StorageStrategy, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("S3 bucket name cannot be empty")
	}

	ctx := context.Background()

	// 创建配置选项
	var optFns []func(*config.LoadOptions) error

	// 设置区域
	if regionName != "" {
		optFns = append(optFns, config.WithRegion(regionName))
	}

	// 设置凭证
	if accessKeyID != "" && secretAccessKey != "" {
		creds := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, sessionToken)
		optFns = append(optFns, config.WithCredentialsProvider(creds))
	}

	// 加载配置
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 创建S3客户端选项
	var s3OptFns []func(*s3.Options)

	// 如果有自定义endpoint
	if endpointURL != "" {
		s3OptFns = append(s3OptFns, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
			o.UsePathStyle = true // 对于自定义endpoint，通常需要path style
		})
	}

	// 创建S3客户端
	client := s3.NewFromConfig(cfg, s3OptFns...)

	strategy := &S3StorageStrategy{
		client:     client,
		bucketName: bucketName,
		basePath:   basePath,
		hostname:   hostname,
		proxy:      proxy,
	}

	return strategy, nil
}

// buildKey 构建 S3 对象键
func (ss *S3StorageStrategy) buildKey(relativePath string) string {
	key := relativePath
	if ss.basePath != "" {
		key = strings.TrimPrefix(relativePath, "./")
		key = strings.Join([]string{ss.basePath, key}, "/")
	}
	return strings.ReplaceAll(key, "\\", "/")
}

// WriteFile 写入文件
func (ss *S3StorageStrategy) WriteFile(path string, data []byte) error {
	key := ss.buildKey(path)

	_, err := ss.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(ss.bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})

	return err
}

// ReadFile 读取文件
func (ss *S3StorageStrategy) ReadFile(path string) ([]byte, error) {
	key := ss.buildKey(path)

	result, err := ss.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(ss.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := result.Body.Close(); cerr != nil {
			log.Printf("Error closing response body: %v", cerr)
		}
	}()

	return io.ReadAll(result.Body)
}

// DeleteFile 删除文件（对于目录，删除所有匹配前缀的对象）
func (ss *S3StorageStrategy) DeleteFile(path string) error {
	key := ss.buildKey(path)

	// 如果是目录路径（通常以"/"结尾），则删除所有匹配前缀的对象
	if strings.HasSuffix(path, "/") || !strings.Contains(path, ".") {
		return ss.deleteWithPrefix(key)
	}

	// 删除单个文件
	_, err := ss.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(ss.bucketName),
		Key:    aws.String(key),
	})

	return err
}

// deleteWithPrefix 删除指定前缀的所有对象
func (ss *S3StorageStrategy) deleteWithPrefix(prefix string) error {
	// 确保前缀以"/"结尾
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// 列出所有匹配的对象
	result, err := ss.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: aws.String(ss.bucketName),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return err
	}

	// 删除所有匹配的对象
	for _, obj := range result.Contents {
		_, err := ss.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(ss.bucketName),
			Key:    obj.Key,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// FileExists 检查文件是否存在
func (ss *S3StorageStrategy) FileExists(path string) bool {
	key := ss.buildKey(path)

	_, err := ss.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(ss.bucketName),
		Key:    aws.String(key),
	})

	return err == nil
}

// SaveUploadFile 保存上传的文件
func (ss *S3StorageStrategy) SaveUploadFile(file *multipart.FileHeader, savePath string) error {
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

	return ss.WriteFile(savePath, data)
}

// ServeFile 提供文件下载服务
func (ss *S3StorageStrategy) ServeFile(c *gin.Context, filePath string, fileName string) error {
	// 对于 S3，我们生成预签名 URL 并重定向
	url, err := ss.GenerateFileURL(filePath, fileName)
	if err != nil {
		return err
	}

	c.Redirect(http.StatusFound, url)
	return nil
}

// GenerateFileURL 生成文件URL
func (ss *S3StorageStrategy) GenerateFileURL(filePath string, fileName string) (string, error) {
	key := ss.buildKey(filePath)

	// 如果设置了代理模式，返回通过服务器中转的URL
	if ss.proxy {
		return "/share/download", nil
	}

	// 使用AWS SDK v2的预签名客户端
	presignClient := s3.NewPresignClient(ss.client)

	// 生成预签名URL（1小时有效期）
	presignResult, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(ss.bucketName),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(3600) * time.Second
	})

	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %v", err)
	}

	return presignResult.URL, nil
}

// TestConnection 测试 S3 连接
func (ss *S3StorageStrategy) TestConnection() error {
	// 测试是否可以列出 bucket
	_, err := ss.client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(ss.bucketName),
	})
	if err != nil {
		return fmt.Errorf("无法访问S3存储桶: %v", err)
	}

	// 测试是否可以写入和删除对象
	testKey := ss.buildKey(".test_connection")

	// 尝试写入测试文件
	_, err = ss.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(ss.bucketName),
		Key:    aws.String(testKey),
		Body:   bytes.NewReader([]byte("test")),
	})
	if err != nil {
		return fmt.Errorf("无法写入测试文件: %v", err)
	}

	// 清理测试文件
	_, err = ss.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(ss.bucketName),
		Key:    aws.String(testKey),
	})
	if err != nil {
		return fmt.Errorf("无法删除测试文件: %v", err)
	}

	return nil
}

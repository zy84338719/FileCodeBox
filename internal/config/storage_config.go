// Package config 存储配置模块
package config

import (
	"fmt"
	"strings"
)

// S3Config S3存储配置
type S3Config struct {
	AccessKeyID      string `json:"s3_access_key_id"`
	SecretAccessKey  string `json:"s3_secret_access_key"`
	BucketName       string `json:"s3_bucket_name"`
	EndpointURL      string `json:"s3_endpoint_url"`
	RegionName       string `json:"s3_region_name"`
	SignatureVersion string `json:"s3_signature_version"`
	Hostname         string `json:"s3_hostname"`
	Proxy            int    `json:"s3_proxy"`
	SessionToken     string `json:"aws_session_token"`
}

// WebDAVConfig WebDAV存储配置
type WebDAVConfig struct {
	Hostname string `json:"webdav_hostname"`
	RootPath string `json:"webdav_root_path"`
	Proxy    int    `json:"webdav_proxy"`
	URL      string `json:"webdav_url"`
	Password string `json:"webdav_password"`
	Username string `json:"webdav_username"`
}

// OneDriveConfig OneDrive存储配置
type OneDriveConfig struct {
	Domain   string `json:"onedrive_domain"`
	ClientID string `json:"onedrive_client_id"`
	Username string `json:"onedrive_username"`
	Password string `json:"onedrive_password"`
	RootPath string `json:"onedrive_root_path"`
	Proxy    int    `json:"onedrive_proxy"`
}

// NFSConfig NFS存储配置
type NFSConfig struct {
	Server     string `json:"nfs_server"`      // NFS服务器地址
	Path       string `json:"nfs_path"`        // NFS路径
	MountPoint string `json:"nfs_mount_point"` // 本地挂载点
	Version    string `json:"nfs_version"`     // NFS版本
	Options    string `json:"nfs_options"`     // 挂载选项
	Timeout    int    `json:"nfs_timeout"`     // 超时时间(秒)
	AutoMount  int    `json:"nfs_auto_mount"`  // 是否自动挂载
	RetryCount int    `json:"nfs_retry_count"` // 重试次数
	SubPath    string `json:"nfs_sub_path"`    // 子路径
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type        string          `json:"file_storage"` // local, s3, webdav, onedrive, nfs
	StoragePath string          `json:"storage_path"`
	S3          *S3Config       `json:"s3,omitempty"`
	WebDAV      *WebDAVConfig   `json:"webdav,omitempty"`
	OneDrive    *OneDriveConfig `json:"onedrive,omitempty"`
	NFS         *NFSConfig      `json:"nfs,omitempty"`
}

// NewS3Config 创建S3配置
func NewS3Config() *S3Config {
	return &S3Config{
		RegionName:       "auto",
		SignatureVersion: "s3v2",
		Proxy:            0,
	}
}

// NewWebDAVConfig 创建WebDAV配置
func NewWebDAVConfig() *WebDAVConfig {
	return &WebDAVConfig{
		RootPath: "filebox_storage",
		Proxy:    0,
	}
}

// NewOneDriveConfig 创建OneDrive配置
func NewOneDriveConfig() *OneDriveConfig {
	return &OneDriveConfig{
		RootPath: "filebox_storage",
		Proxy:    0,
	}
}

// NewNFSConfig 创建NFS配置
func NewNFSConfig() *NFSConfig {
	return &NFSConfig{
		Path:       "/nfs/storage",
		MountPoint: "/mnt/nfs",
		Version:    "4",
		Options:    "rw,sync,hard,intr",
		Timeout:    30,
		AutoMount:  0,
		RetryCount: 3,
		SubPath:    "filebox_storage",
	}
}

// NewStorageConfig 创建存储配置
func NewStorageConfig() *StorageConfig {
	return &StorageConfig{
		Type:        "local",
		StoragePath: "",
		S3:          NewS3Config(),
		WebDAV:      NewWebDAVConfig(),
		OneDrive:    NewOneDriveConfig(),
		NFS:         NewNFSConfig(),
	}
}

// Validate 验证S3配置
func (s3c *S3Config) Validate() error {
	var errors []string

	if strings.TrimSpace(s3c.AccessKeyID) == "" {
		errors = append(errors, "S3 Access Key ID不能为空")
	}

	if strings.TrimSpace(s3c.SecretAccessKey) == "" {
		errors = append(errors, "S3 Secret Access Key不能为空")
	}

	if strings.TrimSpace(s3c.BucketName) == "" {
		errors = append(errors, "S3 Bucket名称不能为空")
	}

	validVersions := []string{"s3v2", "s3v4"}
	if !contains(validVersions, s3c.SignatureVersion) {
		errors = append(errors, "S3签名版本必须是s3v2或s3v4")
	}

	if len(errors) > 0 {
		return fmt.Errorf("S3配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Validate 验证WebDAV配置
func (wc *WebDAVConfig) Validate() error {
	var errors []string

	if strings.TrimSpace(wc.URL) == "" {
		errors = append(errors, "WebDAV URL不能为空")
	}

	if strings.TrimSpace(wc.Username) == "" {
		errors = append(errors, "WebDAV用户名不能为空")
	}

	if strings.TrimSpace(wc.RootPath) == "" {
		errors = append(errors, "WebDAV根路径不能为空")
	}

	if len(errors) > 0 {
		return fmt.Errorf("WebDAV配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Validate 验证OneDrive配置
func (oc *OneDriveConfig) Validate() error {
	var errors []string

	if strings.TrimSpace(oc.ClientID) == "" {
		errors = append(errors, "OneDrive Client ID不能为空")
	}

	if strings.TrimSpace(oc.Username) == "" {
		errors = append(errors, "OneDrive用户名不能为空")
	}

	if strings.TrimSpace(oc.RootPath) == "" {
		errors = append(errors, "OneDrive根路径不能为空")
	}

	if len(errors) > 0 {
		return fmt.Errorf("OneDrive配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Validate 验证NFS配置
func (nc *NFSConfig) Validate() error {
	var errors []string

	if strings.TrimSpace(nc.Server) == "" {
		errors = append(errors, "NFS服务器地址不能为空")
	}

	if strings.TrimSpace(nc.Path) == "" {
		errors = append(errors, "NFS路径不能为空")
	}

	if strings.TrimSpace(nc.MountPoint) == "" {
		errors = append(errors, "NFS挂载点不能为空")
	}

	validVersions := []string{"3", "4", "4.1"}
	if !contains(validVersions, nc.Version) {
		errors = append(errors, "NFS版本必须是3, 4或4.1")
	}

	if nc.Timeout < 1 || nc.Timeout > 3600 {
		errors = append(errors, "NFS超时时间必须在1-3600秒之间")
	}

	if nc.RetryCount < 0 || nc.RetryCount > 10 {
		errors = append(errors, "NFS重试次数必须在0-10之间")
	}

	if len(errors) > 0 {
		return fmt.Errorf("NFS配置验证失败: %s", strings.Join(errors, "; "))
	}

	return nil
}

// Validate 验证存储配置
func (sc *StorageConfig) Validate() error {
	validTypes := []string{"local", "s3", "webdav", "onedrive", "nfs"}
	if !contains(validTypes, sc.Type) {
		return fmt.Errorf("存储类型必须是: %s", strings.Join(validTypes, ", "))
	}

	switch sc.Type {
	case "s3":
		if sc.S3 != nil {
			return sc.S3.Validate()
		}
	case "webdav":
		if sc.WebDAV != nil {
			return sc.WebDAV.Validate()
		}
	case "onedrive":
		if sc.OneDrive != nil {
			return sc.OneDrive.Validate()
		}
	case "nfs":
		if sc.NFS != nil {
			return sc.NFS.Validate()
		}
	}

	return nil
}

// IsLocal 判断是否为本地存储
func (sc *StorageConfig) IsLocal() bool {
	return sc.Type == "local"
}

// IsS3 判断是否为S3存储
func (sc *StorageConfig) IsS3() bool {
	return sc.Type == "s3"
}

// IsWebDAV 判断是否为WebDAV存储
func (sc *StorageConfig) IsWebDAV() bool {
	return sc.Type == "webdav"
}

// IsOneDrive 判断是否为OneDrive存储
func (sc *StorageConfig) IsOneDrive() bool {
	return sc.Type == "onedrive"
}

// IsNFS 判断是否为NFS存储
func (sc *StorageConfig) IsNFS() bool {
	return sc.Type == "nfs"
}

// Clone 克隆S3配置
func (s3c *S3Config) Clone() *S3Config {
	return &S3Config{
		AccessKeyID:      s3c.AccessKeyID,
		SecretAccessKey:  s3c.SecretAccessKey,
		BucketName:       s3c.BucketName,
		EndpointURL:      s3c.EndpointURL,
		RegionName:       s3c.RegionName,
		SignatureVersion: s3c.SignatureVersion,
		Hostname:         s3c.Hostname,
		Proxy:            s3c.Proxy,
		SessionToken:     s3c.SessionToken,
	}
}

// Clone 克隆WebDAV配置
func (wc *WebDAVConfig) Clone() *WebDAVConfig {
	return &WebDAVConfig{
		Hostname: wc.Hostname,
		RootPath: wc.RootPath,
		Proxy:    wc.Proxy,
		URL:      wc.URL,
		Password: wc.Password,
		Username: wc.Username,
	}
}

// Clone 克隆OneDrive配置
func (oc *OneDriveConfig) Clone() *OneDriveConfig {
	return &OneDriveConfig{
		Domain:   oc.Domain,
		ClientID: oc.ClientID,
		Username: oc.Username,
		Password: oc.Password,
		RootPath: oc.RootPath,
		Proxy:    oc.Proxy,
	}
}

// Clone 克隆NFS配置
func (nc *NFSConfig) Clone() *NFSConfig {
	return &NFSConfig{
		Server:     nc.Server,
		Path:       nc.Path,
		MountPoint: nc.MountPoint,
		Version:    nc.Version,
		Options:    nc.Options,
		Timeout:    nc.Timeout,
		AutoMount:  nc.AutoMount,
		RetryCount: nc.RetryCount,
		SubPath:    nc.SubPath,
	}
}

// Clone 克隆存储配置
func (sc *StorageConfig) Clone() *StorageConfig {
	clone := &StorageConfig{
		Type:        sc.Type,
		StoragePath: sc.StoragePath,
	}

	if sc.S3 != nil {
		clone.S3 = sc.S3.Clone()
	}
	if sc.WebDAV != nil {
		clone.WebDAV = sc.WebDAV.Clone()
	}
	if sc.OneDrive != nil {
		clone.OneDrive = sc.OneDrive.Clone()
	}
	if sc.NFS != nil {
		clone.NFS = sc.NFS.Clone()
	}

	return clone
}

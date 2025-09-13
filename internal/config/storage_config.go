// Package config 存储配置模块
package config

import (
	"fmt"
	"strconv"
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

// ToMap S3配置转换为map格式
func (s3c *S3Config) ToMap() map[string]string {
	return map[string]string{
		"s3_access_key_id":     s3c.AccessKeyID,
		"s3_secret_access_key": s3c.SecretAccessKey,
		"s3_bucket_name":       s3c.BucketName,
		"s3_endpoint_url":      s3c.EndpointURL,
		"s3_region_name":       s3c.RegionName,
		"s3_signature_version": s3c.SignatureVersion,
		"s3_hostname":          s3c.Hostname,
		"s3_proxy":             fmt.Sprintf("%d", s3c.Proxy),
		"aws_session_token":    s3c.SessionToken,
	}
}

// ToMap WebDAV配置转换为map格式
func (wc *WebDAVConfig) ToMap() map[string]string {
	return map[string]string{
		"webdav_hostname":  wc.Hostname,
		"webdav_root_path": wc.RootPath,
		"webdav_proxy":     fmt.Sprintf("%d", wc.Proxy),
		"webdav_url":       wc.URL,
		"webdav_password":  wc.Password,
		"webdav_username":  wc.Username,
	}
}

// ToMap OneDrive配置转换为map格式
func (oc *OneDriveConfig) ToMap() map[string]string {
	return map[string]string{
		"onedrive_domain":    oc.Domain,
		"onedrive_client_id": oc.ClientID,
		"onedrive_username":  oc.Username,
		"onedrive_password":  oc.Password,
		"onedrive_root_path": oc.RootPath,
		"onedrive_proxy":     fmt.Sprintf("%d", oc.Proxy),
	}
}

// ToMap NFS配置转换为map格式
func (nc *NFSConfig) ToMap() map[string]string {
	return map[string]string{
		"nfs_server":      nc.Server,
		"nfs_path":        nc.Path,
		"nfs_mount_point": nc.MountPoint,
		"nfs_version":     nc.Version,
		"nfs_options":     nc.Options,
		"nfs_timeout":     fmt.Sprintf("%d", nc.Timeout),
		"nfs_auto_mount":  fmt.Sprintf("%d", nc.AutoMount),
		"nfs_retry_count": fmt.Sprintf("%d", nc.RetryCount),
		"nfs_sub_path":    nc.SubPath,
	}
}

// ToMap 存储配置转换为map格式
func (sc *StorageConfig) ToMap() map[string]string {
	result := map[string]string{
		"file_storage": sc.Type,
		"storage_path": sc.StoragePath,
	}

	// 根据存储类型添加对应配置
	switch sc.Type {
	case "s3":
		if sc.S3 != nil {
			for k, v := range sc.S3.ToMap() {
				result[k] = v
			}
		}
	case "webdav":
		if sc.WebDAV != nil {
			for k, v := range sc.WebDAV.ToMap() {
				result[k] = v
			}
		}
	case "onedrive":
		if sc.OneDrive != nil {
			for k, v := range sc.OneDrive.ToMap() {
				result[k] = v
			}
		}
	case "nfs":
		if sc.NFS != nil {
			for k, v := range sc.NFS.ToMap() {
				result[k] = v
			}
		}
	}

	return result
}

// FromMap 从map加载S3配置
func (s3c *S3Config) FromMap(data map[string]string) error {
	if val, ok := data["s3_access_key_id"]; ok {
		s3c.AccessKeyID = val
	}
	if val, ok := data["s3_secret_access_key"]; ok {
		s3c.SecretAccessKey = val
	}
	if val, ok := data["s3_bucket_name"]; ok {
		s3c.BucketName = val
	}
	if val, ok := data["s3_endpoint_url"]; ok {
		s3c.EndpointURL = val
	}
	if val, ok := data["s3_region_name"]; ok {
		s3c.RegionName = val
	}
	if val, ok := data["s3_signature_version"]; ok {
		s3c.SignatureVersion = val
	}
	if val, ok := data["s3_hostname"]; ok {
		s3c.Hostname = val
	}
	if val, ok := data["s3_proxy"]; ok {
		if proxy, err := strconv.Atoi(val); err == nil {
			s3c.Proxy = proxy
		}
	}
	if val, ok := data["aws_session_token"]; ok {
		s3c.SessionToken = val
	}

	return nil
}

// FromMap 从map加载WebDAV配置
func (wc *WebDAVConfig) FromMap(data map[string]string) error {
	if val, ok := data["webdav_hostname"]; ok {
		wc.Hostname = val
	}
	if val, ok := data["webdav_root_path"]; ok {
		wc.RootPath = val
	}
	if val, ok := data["webdav_proxy"]; ok {
		if proxy, err := strconv.Atoi(val); err == nil {
			wc.Proxy = proxy
		}
	}
	if val, ok := data["webdav_url"]; ok {
		wc.URL = val
	}
	if val, ok := data["webdav_password"]; ok {
		wc.Password = val
	}
	if val, ok := data["webdav_username"]; ok {
		wc.Username = val
	}

	return nil
}

// FromMap 从map加载OneDrive配置
func (oc *OneDriveConfig) FromMap(data map[string]string) error {
	if val, ok := data["onedrive_domain"]; ok {
		oc.Domain = val
	}
	if val, ok := data["onedrive_client_id"]; ok {
		oc.ClientID = val
	}
	if val, ok := data["onedrive_username"]; ok {
		oc.Username = val
	}
	if val, ok := data["onedrive_password"]; ok {
		oc.Password = val
	}
	if val, ok := data["onedrive_root_path"]; ok {
		oc.RootPath = val
	}
	if val, ok := data["onedrive_proxy"]; ok {
		if proxy, err := strconv.Atoi(val); err == nil {
			oc.Proxy = proxy
		}
	}

	return nil
}

// FromMap 从map加载NFS配置
func (nc *NFSConfig) FromMap(data map[string]string) error {
	if val, ok := data["nfs_server"]; ok {
		nc.Server = val
	}
	if val, ok := data["nfs_path"]; ok {
		nc.Path = val
	}
	if val, ok := data["nfs_mount_point"]; ok {
		nc.MountPoint = val
	}
	if val, ok := data["nfs_version"]; ok {
		nc.Version = val
	}
	if val, ok := data["nfs_options"]; ok {
		nc.Options = val
	}
	if val, ok := data["nfs_timeout"]; ok {
		if timeout, err := strconv.Atoi(val); err == nil {
			nc.Timeout = timeout
		}
	}
	if val, ok := data["nfs_auto_mount"]; ok {
		if autoMount, err := strconv.Atoi(val); err == nil {
			nc.AutoMount = autoMount
		}
	}
	if val, ok := data["nfs_retry_count"]; ok {
		if retryCount, err := strconv.Atoi(val); err == nil {
			nc.RetryCount = retryCount
		}
	}
	if val, ok := data["nfs_sub_path"]; ok {
		nc.SubPath = val
	}

	return nil
}

// FromMap 从map加载存储配置
func (sc *StorageConfig) FromMap(data map[string]string) error {
	if val, ok := data["file_storage"]; ok {
		sc.Type = val
	}
	if val, ok := data["storage_path"]; ok {
		sc.StoragePath = val
	}

	// 根据存储类型加载对应配置
	switch sc.Type {
	case "s3":
		if sc.S3 == nil {
			sc.S3 = NewS3Config()
		}
		if err := sc.S3.FromMap(data); err != nil {
			return fmt.Errorf("failed to parse S3 config: %w", err)
		}
	case "webdav":
		if sc.WebDAV == nil {
			sc.WebDAV = NewWebDAVConfig()
		}
		if err := sc.WebDAV.FromMap(data); err != nil {
			return fmt.Errorf("failed to parse WebDAV config: %w", err)
		}
	case "onedrive":
		if sc.OneDrive == nil {
			sc.OneDrive = NewOneDriveConfig()
		}
		if err := sc.OneDrive.FromMap(data); err != nil {
			return fmt.Errorf("failed to parse OneDrive config: %w", err)
		}
	case "nfs":
		if sc.NFS == nil {
			sc.NFS = NewNFSConfig()
		}
		if err := sc.NFS.FromMap(data); err != nil {
			return fmt.Errorf("failed to parse NFS config: %w", err)
		}
	}

	return sc.Validate()
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

package dto

// ConfigUpdateFields 配置更新字段结构体
type ConfigUpdateFields struct {
	Base     *BaseConfigUpdate     `json:"base,omitempty"`
	Transfer *TransferConfigUpdate `json:"transfer,omitempty"`
	User     *UserConfigUpdate     `json:"user,omitempty"`
	Storage  *StorageConfigUpdate  `json:"storage,omitempty"`
	MCP      *MCPConfigUpdate      `json:"mcp,omitempty"`

	// 其他配置字段
	NotifyTitle   *string `json:"notify_title,omitempty"`
	NotifyContent *string `json:"notify_content,omitempty"`
	PageExplain   *string `json:"page_explain,omitempty"`
	Opacity       *int    `json:"opacity,omitempty"`
	ThemesSelect  *string `json:"themes_select,omitempty"`
}

// BaseConfigUpdate 基础配置更新
type BaseConfigUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Keywords    *string `json:"keywords,omitempty"`
	Port        *int    `json:"port,omitempty"`
	Host        *string `json:"host,omitempty"`
	DataPath    *string `json:"data_path,omitempty"`
	Production  *bool   `json:"production,omitempty"`
}

// TransferConfigUpdate 传输配置更新
type TransferConfigUpdate struct {
	Upload   *UploadConfigUpdate   `json:"upload,omitempty"`
	Download *DownloadConfigUpdate `json:"download,omitempty"`
}

// UploadConfigUpdate 上传配置更新
type UploadConfigUpdate struct {
	OpenUpload     *int   `json:"open_upload,omitempty"`
	UploadSize     *int64 `json:"upload_size,omitempty"`
	EnableChunk    *int   `json:"enable_chunk,omitempty"`
	ChunkSize      *int64 `json:"chunk_size,omitempty"`
	MaxSaveSeconds *int   `json:"max_save_seconds,omitempty"`
}

// DownloadConfigUpdate 下载配置更新
type DownloadConfigUpdate struct {
	EnableConcurrentDownload *int `json:"enable_concurrent_download,omitempty"`
	MaxConcurrentDownloads   *int `json:"max_concurrent_downloads,omitempty"`
	DownloadTimeout          *int `json:"download_timeout,omitempty"`
}

// UserConfigUpdate 用户配置更新
type UserConfigUpdate struct {
	AllowUserRegistration *int    `json:"allow_user_registration,omitempty"`
	RequireEmailVerify    *int    `json:"require_email_verify,omitempty"`
	UserUploadSize        *int64  `json:"user_upload_size,omitempty"`
	UserStorageQuota      *int64  `json:"user_storage_quota,omitempty"`
	SessionExpiryHours    *int    `json:"session_expiry_hours,omitempty"`
	MaxSessionsPerUser    *int    `json:"max_sessions_per_user,omitempty"`
	JWTSecret             *string `json:"jwt_secret,omitempty"`
}

// StorageConfigUpdate 存储配置更新
type StorageConfigUpdate struct {
	FileStorage *string               `json:"file_storage,omitempty"`
	StoragePath *string               `json:"storage_path,omitempty"`
	S3          *S3ConfigUpdate       `json:"s3,omitempty"`
	WebDAV      *WebDAVConfigUpdate   `json:"webdav,omitempty"`
	OneDrive    *OneDriveConfigUpdate `json:"onedrive,omitempty"`
	NFS         *NFSConfigUpdate      `json:"nfs,omitempty"`
}

// S3ConfigUpdate S3存储配置更新
type S3ConfigUpdate struct {
	S3AccessKeyID      *string `json:"s3_access_key_id,omitempty"`
	S3SecretAccessKey  *string `json:"s3_secret_access_key,omitempty"`
	S3BucketName       *string `json:"s3_bucket_name,omitempty"`
	S3EndpointURL      *string `json:"s3_endpoint_url,omitempty"`
	S3RegionName       *string `json:"s3_region_name,omitempty"`
	S3SignatureVersion *string `json:"s3_signature_version,omitempty"`
	S3Hostname         *string `json:"s3_hostname,omitempty"`
	S3Proxy            *int    `json:"s3_proxy,omitempty"`
	AWSSessionToken    *string `json:"aws_session_token,omitempty"`
}

// WebDAVConfigUpdate WebDAV存储配置更新
type WebDAVConfigUpdate struct {
	WebDAVHostname *string `json:"webdav_hostname,omitempty"`
	WebDAVRootPath *string `json:"webdav_root_path,omitempty"`
	WebDAVProxy    *int    `json:"webdav_proxy,omitempty"`
	WebDAVURL      *string `json:"webdav_url,omitempty"`
	WebDAVPassword *string `json:"webdav_password,omitempty"`
	WebDAVUsername *string `json:"webdav_username,omitempty"`
}

// OneDriveConfigUpdate OneDrive存储配置更新
type OneDriveConfigUpdate struct {
	OneDriveDomain   *string `json:"onedrive_domain,omitempty"`
	OneDriveClientID *string `json:"onedrive_client_id,omitempty"`
	OneDriveUsername *string `json:"onedrive_username,omitempty"`
	OneDrivePassword *string `json:"onedrive_password,omitempty"`
	OneDriveRootPath *string `json:"onedrive_root_path,omitempty"`
	OneDriveProxy    *int    `json:"onedrive_proxy,omitempty"`
}

// NFSConfigUpdate NFS存储配置更新
type NFSConfigUpdate struct {
	NFSServer     *string `json:"nfs_server,omitempty"`
	NFSPath       *string `json:"nfs_path,omitempty"`
	NFSMountPoint *string `json:"nfs_mount_point,omitempty"`
	NFSVersion    *string `json:"nfs_version,omitempty"`
	NFSOptions    *string `json:"nfs_options,omitempty"`
	NFSTimeout    *int    `json:"nfs_timeout,omitempty"`
	NFSAutoMount  *int    `json:"nfs_auto_mount,omitempty"`
	NFSRetryCount *int    `json:"nfs_retry_count,omitempty"`
	NFSSubPath    *string `json:"nfs_sub_path,omitempty"`
}

// MCPConfigUpdate MCP配置更新
type MCPConfigUpdate struct {
	EnableMCPServer *int    `json:"enable_mcp_server,omitempty"`
	MCPPort         *string `json:"mcp_port,omitempty"`
	MCPHost         *string `json:"mcp_host,omitempty"`
}

// ToMap 将结构体转换为 map，只包含非空字段
func (c *ConfigUpdateFields) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if c.Base != nil {
		baseMap := c.Base.ToMap()
		if len(baseMap) > 0 {
			updates["base"] = baseMap
		}
	}

	if c.Transfer != nil {
		transferMap := c.Transfer.ToMap()
		if len(transferMap) > 0 {
			updates["transfer"] = transferMap
		}
	}

	if c.User != nil {
		userMap := c.User.ToMap()
		if len(userMap) > 0 {
			updates["user"] = userMap
		}
	}

	if c.Storage != nil {
		storageMap := c.Storage.ToMap()
		if len(storageMap) > 0 {
			updates["storage"] = storageMap
		}
	}

	if c.MCP != nil {
		mcpMap := c.MCP.ToMap()
		if len(mcpMap) > 0 {
			updates["mcp"] = mcpMap
		}
	}

	if c.NotifyTitle != nil {
		updates["notify_title"] = *c.NotifyTitle
	}
	if c.NotifyContent != nil {
		updates["notify_content"] = *c.NotifyContent
	}
	if c.PageExplain != nil {
		updates["page_explain"] = *c.PageExplain
	}
	if c.Opacity != nil {
		updates["opacity"] = *c.Opacity
	}
	if c.ThemesSelect != nil {
		updates["themes_select"] = *c.ThemesSelect
	}

	return updates
}

// ToMap 将基础配置转换为 map
func (b *BaseConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if b.Name != nil {
		updates["name"] = *b.Name
	}
	if b.Description != nil {
		updates["description"] = *b.Description
	}
	if b.Keywords != nil {
		updates["keywords"] = *b.Keywords
	}
	if b.Port != nil {
		updates["port"] = *b.Port
	}
	if b.Host != nil {
		updates["host"] = *b.Host
	}
	if b.DataPath != nil {
		updates["data_path"] = *b.DataPath
	}
	if b.Production != nil {
		updates["production"] = *b.Production
	}

	return updates
}

// ToMap 将传输配置转换为 map
func (t *TransferConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if t.Upload != nil {
		uploadMap := t.Upload.ToMap()
		if len(uploadMap) > 0 {
			updates["upload"] = uploadMap
		}
	}

	if t.Download != nil {
		downloadMap := t.Download.ToMap()
		if len(downloadMap) > 0 {
			updates["download"] = downloadMap
		}
	}

	return updates
}

// ToMap 将上传配置转换为 map
func (u *UploadConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if u.OpenUpload != nil {
		updates["open_upload"] = *u.OpenUpload
	}
	if u.UploadSize != nil {
		updates["upload_size"] = *u.UploadSize
	}
	if u.EnableChunk != nil {
		updates["enable_chunk"] = *u.EnableChunk
	}
	if u.ChunkSize != nil {
		updates["chunk_size"] = *u.ChunkSize
	}
	if u.MaxSaveSeconds != nil {
		updates["max_save_seconds"] = *u.MaxSaveSeconds
	}

	return updates
}

// ToMap 将下载配置转换为 map
func (d *DownloadConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if d.EnableConcurrentDownload != nil {
		updates["enable_concurrent_download"] = *d.EnableConcurrentDownload
	}
	if d.MaxConcurrentDownloads != nil {
		updates["max_concurrent_downloads"] = *d.MaxConcurrentDownloads
	}
	if d.DownloadTimeout != nil {
		updates["download_timeout"] = *d.DownloadTimeout
	}

	return updates
}

// ToMap 将用户配置转换为 map
func (u *UserConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if u.AllowUserRegistration != nil {
		updates["allow_user_registration"] = *u.AllowUserRegistration
	}
	if u.RequireEmailVerify != nil {
		updates["require_email_verify"] = *u.RequireEmailVerify
	}
	if u.UserUploadSize != nil {
		updates["user_upload_size"] = *u.UserUploadSize
	}
	if u.UserStorageQuota != nil {
		updates["user_storage_quota"] = *u.UserStorageQuota
	}
	if u.SessionExpiryHours != nil {
		updates["session_expiry_hours"] = *u.SessionExpiryHours
	}
	if u.MaxSessionsPerUser != nil {
		updates["max_sessions_per_user"] = *u.MaxSessionsPerUser
	}
	if u.JWTSecret != nil {
		updates["jwt_secret"] = *u.JWTSecret
	}

	return updates
}

// ToMap 将MCP配置转换为 map
func (m *MCPConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if m.EnableMCPServer != nil {
		updates["enable_mcp_server"] = *m.EnableMCPServer
	}
	if m.MCPPort != nil {
		updates["mcp_port"] = *m.MCPPort
	}
	if m.MCPHost != nil {
		updates["mcp_host"] = *m.MCPHost
	}

	return updates
}

// ToMap 将存储配置转换为 map
func (s *StorageConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if s.FileStorage != nil {
		updates["file_storage"] = *s.FileStorage
	}
	if s.StoragePath != nil {
		updates["storage_path"] = *s.StoragePath
	}
	if s.S3 != nil {
		s3Map := s.S3.ToMap()
		if len(s3Map) > 0 {
			updates["s3"] = s3Map
		}
	}
	if s.WebDAV != nil {
		webdavMap := s.WebDAV.ToMap()
		if len(webdavMap) > 0 {
			updates["webdav"] = webdavMap
		}
	}
	if s.OneDrive != nil {
		onedriveMap := s.OneDrive.ToMap()
		if len(onedriveMap) > 0 {
			updates["onedrive"] = onedriveMap
		}
	}
	if s.NFS != nil {
		nfsMap := s.NFS.ToMap()
		if len(nfsMap) > 0 {
			updates["nfs"] = nfsMap
		}
	}

	return updates
}

// ToMap 将S3配置转换为 map
func (s *S3ConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if s.S3AccessKeyID != nil {
		updates["s3_access_key_id"] = *s.S3AccessKeyID
	}
	if s.S3SecretAccessKey != nil {
		updates["s3_secret_access_key"] = *s.S3SecretAccessKey
	}
	if s.S3BucketName != nil {
		updates["s3_bucket_name"] = *s.S3BucketName
	}
	if s.S3EndpointURL != nil {
		updates["s3_endpoint_url"] = *s.S3EndpointURL
	}
	if s.S3RegionName != nil {
		updates["s3_region_name"] = *s.S3RegionName
	}
	if s.S3SignatureVersion != nil {
		updates["s3_signature_version"] = *s.S3SignatureVersion
	}
	if s.S3Hostname != nil {
		updates["s3_hostname"] = *s.S3Hostname
	}
	if s.S3Proxy != nil {
		updates["s3_proxy"] = *s.S3Proxy
	}
	if s.AWSSessionToken != nil {
		updates["aws_session_token"] = *s.AWSSessionToken
	}

	return updates
}

// ToMap 将WebDAV配置转换为 map
func (w *WebDAVConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if w.WebDAVHostname != nil {
		updates["webdav_hostname"] = *w.WebDAVHostname
	}
	if w.WebDAVRootPath != nil {
		updates["webdav_root_path"] = *w.WebDAVRootPath
	}
	if w.WebDAVProxy != nil {
		updates["webdav_proxy"] = *w.WebDAVProxy
	}
	if w.WebDAVURL != nil {
		updates["webdav_url"] = *w.WebDAVURL
	}
	if w.WebDAVPassword != nil {
		updates["webdav_password"] = *w.WebDAVPassword
	}
	if w.WebDAVUsername != nil {
		updates["webdav_username"] = *w.WebDAVUsername
	}

	return updates
}

// ToMap 将OneDrive配置转换为 map
func (o *OneDriveConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if o.OneDriveDomain != nil {
		updates["onedrive_domain"] = *o.OneDriveDomain
	}
	if o.OneDriveClientID != nil {
		updates["onedrive_client_id"] = *o.OneDriveClientID
	}
	if o.OneDriveUsername != nil {
		updates["onedrive_username"] = *o.OneDriveUsername
	}
	if o.OneDrivePassword != nil {
		updates["onedrive_password"] = *o.OneDrivePassword
	}
	if o.OneDriveRootPath != nil {
		updates["onedrive_root_path"] = *o.OneDriveRootPath
	}
	if o.OneDriveProxy != nil {
		updates["onedrive_proxy"] = *o.OneDriveProxy
	}

	return updates
}

// ToMap 将NFS配置转换为 map
func (n *NFSConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if n.NFSServer != nil {
		updates["nfs_server"] = *n.NFSServer
	}
	if n.NFSPath != nil {
		updates["nfs_path"] = *n.NFSPath
	}
	if n.NFSMountPoint != nil {
		updates["nfs_mount_point"] = *n.NFSMountPoint
	}
	if n.NFSVersion != nil {
		updates["nfs_version"] = *n.NFSVersion
	}
	if n.NFSOptions != nil {
		updates["nfs_options"] = *n.NFSOptions
	}
	if n.NFSTimeout != nil {
		updates["nfs_timeout"] = *n.NFSTimeout
	}
	if n.NFSAutoMount != nil {
		updates["nfs_auto_mount"] = *n.NFSAutoMount
	}
	if n.NFSRetryCount != nil {
		updates["nfs_retry_count"] = *n.NFSRetryCount
	}
	if n.NFSSubPath != nil {
		updates["nfs_sub_path"] = *n.NFSSubPath
	}

	return updates
}

// HasUpdates 检查是否有任何更新字段
func (c *ConfigUpdateFields) HasUpdates() bool {
	return c.Base != nil || c.Transfer != nil || c.User != nil ||
		c.Storage != nil || c.MCP != nil ||
		c.NotifyTitle != nil || c.NotifyContent != nil ||
		c.PageExplain != nil || c.Opacity != nil || c.ThemesSelect != nil
}

// FlatConfigUpdate 平面化配置更新（用于兼容老的API格式）
type FlatConfigUpdate struct {
	// 基础配置
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Keywords    *string `json:"keywords,omitempty"`
	Port        *int    `json:"port,omitempty"`
	Host        *string `json:"host,omitempty"`
	DataPath    *string `json:"data_path,omitempty"`
	Production  *bool   `json:"production,omitempty"`

	// 传输配置
	OpenUpload               *int   `json:"open_upload,omitempty"`
	UploadSize               *int64 `json:"upload_size,omitempty"`
	EnableChunk              *int   `json:"enable_chunk,omitempty"`
	ChunkSize                *int64 `json:"chunk_size,omitempty"`
	MaxSaveSeconds           *int   `json:"max_save_seconds,omitempty"`
	EnableConcurrentDownload *int   `json:"enable_concurrent_download,omitempty"`
	MaxConcurrentDownloads   *int   `json:"max_concurrent_downloads,omitempty"`
	DownloadTimeout          *int   `json:"download_timeout,omitempty"`

	// 用户配置
	AllowUserRegistration *int    `json:"allow_user_registration,omitempty"`
	RequireEmailVerify    *int    `json:"require_email_verify,omitempty"`
	UserUploadSize        *int64  `json:"user_upload_size,omitempty"`
	UserStorageQuota      *int64  `json:"user_storage_quota,omitempty"`
	SessionExpiryHours    *int    `json:"session_expiry_hours,omitempty"`
	MaxSessionsPerUser    *int    `json:"max_sessions_per_user,omitempty"`
	JWTSecret             *string `json:"jwt_secret,omitempty"`

	// MCP配置
	EnableMCPServer *int    `json:"enable_mcp_server,omitempty"`
	MCPPort         *string `json:"mcp_port,omitempty"`
	MCPHost         *string `json:"mcp_host,omitempty"`

	// 其他配置
	NotifyTitle   *string `json:"notify_title,omitempty"`
	NotifyContent *string `json:"notify_content,omitempty"`
	PageExplain   *string `json:"page_explain,omitempty"`
	Opacity       *int    `json:"opacity,omitempty"`
	ThemesSelect  *string `json:"themes_select,omitempty"`
}

// ToMap 将平面化配置转换为 map
func (f *FlatConfigUpdate) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	// 基础配置
	if f.Name != nil {
		updates["name"] = *f.Name
	}
	if f.Description != nil {
		updates["description"] = *f.Description
	}
	if f.Keywords != nil {
		updates["keywords"] = *f.Keywords
	}
	if f.Port != nil {
		updates["port"] = *f.Port
	}
	if f.Host != nil {
		updates["host"] = *f.Host
	}
	if f.DataPath != nil {
		updates["data_path"] = *f.DataPath
	}
	if f.Production != nil {
		updates["production"] = *f.Production
	}

	// 传输配置
	if f.OpenUpload != nil {
		updates["open_upload"] = *f.OpenUpload
	}
	if f.UploadSize != nil {
		updates["upload_size"] = *f.UploadSize
	}
	if f.EnableChunk != nil {
		updates["enable_chunk"] = *f.EnableChunk
	}
	if f.ChunkSize != nil {
		updates["chunk_size"] = *f.ChunkSize
	}
	if f.MaxSaveSeconds != nil {
		updates["max_save_seconds"] = *f.MaxSaveSeconds
	}
	if f.EnableConcurrentDownload != nil {
		updates["enable_concurrent_download"] = *f.EnableConcurrentDownload
	}
	if f.MaxConcurrentDownloads != nil {
		updates["max_concurrent_downloads"] = *f.MaxConcurrentDownloads
	}
	if f.DownloadTimeout != nil {
		updates["download_timeout"] = *f.DownloadTimeout
	}

	// 用户配置
	if f.AllowUserRegistration != nil {
		updates["allow_user_registration"] = *f.AllowUserRegistration
	}
	if f.RequireEmailVerify != nil {
		updates["require_email_verify"] = *f.RequireEmailVerify
	}
	if f.UserUploadSize != nil {
		updates["user_upload_size"] = *f.UserUploadSize
	}
	if f.UserStorageQuota != nil {
		updates["user_storage_quota"] = *f.UserStorageQuota
	}
	if f.SessionExpiryHours != nil {
		updates["session_expiry_hours"] = *f.SessionExpiryHours
	}
	if f.MaxSessionsPerUser != nil {
		updates["max_sessions_per_user"] = *f.MaxSessionsPerUser
	}
	if f.JWTSecret != nil {
		updates["jwt_secret"] = *f.JWTSecret
	}

	// MCP配置
	if f.EnableMCPServer != nil {
		updates["enable_mcp_server"] = *f.EnableMCPServer
	}
	if f.MCPPort != nil {
		updates["mcp_port"] = *f.MCPPort
	}
	if f.MCPHost != nil {
		updates["mcp_host"] = *f.MCPHost
	}

	// 其他配置
	if f.NotifyTitle != nil {
		updates["notify_title"] = *f.NotifyTitle
	}
	if f.NotifyContent != nil {
		updates["notify_content"] = *f.NotifyContent
	}
	if f.PageExplain != nil {
		updates["page_explain"] = *f.PageExplain
	}
	if f.Opacity != nil {
		updates["opacity"] = *f.Opacity
	}
	if f.ThemesSelect != nil {
		updates["themes_select"] = *f.ThemesSelect
	}

	return updates
}

// HasUpdates 检查是否有任何更新字段
func (f *FlatConfigUpdate) HasUpdates() bool {
	return f.Name != nil || f.Description != nil || f.Keywords != nil ||
		f.Port != nil || f.Host != nil || f.DataPath != nil || f.Production != nil ||
		f.OpenUpload != nil || f.UploadSize != nil || f.EnableChunk != nil ||
		f.ChunkSize != nil || f.MaxSaveSeconds != nil ||
		f.EnableConcurrentDownload != nil || f.MaxConcurrentDownloads != nil ||
		f.DownloadTimeout != nil || f.AllowUserRegistration != nil ||
		f.RequireEmailVerify != nil || f.UserUploadSize != nil ||
		f.UserStorageQuota != nil || f.SessionExpiryHours != nil ||
		f.MaxSessionsPerUser != nil || f.JWTSecret != nil ||
		f.EnableMCPServer != nil || f.MCPPort != nil || f.MCPHost != nil ||
		f.NotifyTitle != nil || f.NotifyContent != nil ||
		f.PageExplain != nil || f.Opacity != nil || f.ThemesSelect != nil
}

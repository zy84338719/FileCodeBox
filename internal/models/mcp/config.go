package mcp

// SystemConfigResponse MCP系统配置响应结构
type SystemConfigResponse struct {
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Port                  int    `json:"port"`
	Host                  string `json:"host"`
	DataPath              string `json:"data_path"`
	FileStorage           string `json:"file_storage"`
	AllowUserRegistration bool   `json:"allow_user_registration"`
	UploadSize            int64  `json:"upload_size"`
	MaxSaveSeconds        int    `json:"max_save_seconds"`
}

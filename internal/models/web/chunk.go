package web

// ChunkUploadInitRequest 分片上传初始化请求
type ChunkUploadInitRequest struct {
	FileName  string `json:"file_name" binding:"required"`
	FileSize  int64  `json:"file_size" binding:"required"`
	ChunkSize int    `json:"chunk_size" binding:"required"`
	FileHash  string `json:"file_hash" binding:"required"`
}

// ChunkUploadInitResponse 分片上传初始化响应
type ChunkUploadInitResponse struct {
	UploadID      string  `json:"upload_id"`
	TotalChunks   int     `json:"total_chunks"`
	ChunkSize     int     `json:"chunk_size"`
	UploadedCount int     `json:"uploaded_count"`
	Progress      float64 `json:"progress"`
}

// ChunkUploadResponse 分片上传响应
type ChunkUploadResponse struct {
	ChunkHash  string  `json:"chunk_hash"`
	ChunkIndex int     `json:"chunk_index"`
	Progress   float64 `json:"progress"`
}

// ChunkUploadCompleteRequest 分片上传完成请求
type ChunkUploadCompleteRequest struct {
	ExpireValue int    `json:"expire_value" binding:"required"`
	ExpireStyle string `json:"expire_style" binding:"required"`
	RequireAuth bool   `json:"require_auth"`
}

// ChunkUploadCompleteResponse 分片上传完成响应
type ChunkUploadCompleteResponse struct {
	Code     string `json:"code"`
	ShareURL string `json:"share_url"`
	FileName string `json:"file_name"`
}

// ChunkUploadStatusResponse 分片上传状态响应
type ChunkUploadStatusResponse struct {
	UploadID      string  `json:"upload_id"`
	Status        string  `json:"status"`
	Progress      float64 `json:"progress"`
	TotalChunks   int     `json:"total_chunks"`
	UploadedCount int     `json:"uploaded_count"`
	FileName      string  `json:"file_name"`
	FileSize      int64   `json:"file_size"`
}

// ChunkValidationResponse 分片验证响应
type ChunkValidationResponse struct {
	Valid bool `json:"valid"`
}

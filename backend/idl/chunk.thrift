// =====================================================================
// chunk.thrift — 分片上传
// =====================================================================
// 迁移自: idl/chunk.proto
// =====================================================================

namespace go chunk

// ==================== 初始化分片上传 ====================

struct ChunkUploadInitReq {
    1: required string file_name    (api.body = "file_name"),
    2: required i64    file_size    (api.body = "file_size"),
    3: required string file_hash    (api.body = "file_hash"),
    4: required i32    chunk_size   (api.body = "chunk_size"),
    5: required i32    total_chunks (api.body = "total_chunks"),
}

struct ChunkUploadInitData {
    1: required string upload_id        (api.body = "upload_id"),
    2: required string chunk_size       (api.body = "chunk_size"),
    3: required string total_chunks     (api.body = "total_chunks"),
    4: required bool   is_quick_upload  (api.body = "is_quick_upload"),     // Quick upload flag
    5: required string share_code       (api.body = "share_code"),          // If quick upload, return share code directly
}

struct ChunkUploadInitResp {
    1: required i32                  code    (api.body = "code"),
    2: required string               message (api.body = "message"),
    3: required ChunkUploadInitData  data    (api.body = "data"),
}

// ==================== 上传单个分片 ====================

struct ChunkUploadReq {
    1: required string upload_id   (api.path = "upload_id"),
    2: required i32    chunk_index (api.path = "chunk_index"),
    // 分片数据通过 multipart/form-data 上传
}

struct ChunkUploadData {
    1: required i32    chunk_index (api.body = "chunk_index"),
    2: required string chunk_hash  (api.body = "chunk_hash"),
}

struct ChunkUploadResp {
    1: required i32           code    (api.body = "code"),
    2: required string        message (api.body = "message"),
    3: required ChunkUploadData data   (api.body = "data"),
}

// ==================== 查询上传进度 ====================

struct ChunkUploadStatusReq {
    1: required string upload_id (api.path = "upload_id"),
}

struct ChunkUploadStatusData {
    1: required string upload_id        (api.body = "upload_id"),
    2: required i32    total_chunks     (api.body = "total_chunks"),
    3: required i32    uploaded_chunks  (api.body = "uploaded_chunks"),
    4: required list<i32> uploaded_indexes (api.body = "uploaded_indexes"),
    5: required i32    progress         (api.body = "progress"),   // 百分比
    6: required string status           (api.body = "status"),     // uploading, completed, cancelled
}

struct ChunkUploadStatusResp {
    1: required i32                     code    (api.body = "code"),
    2: required string                  message (api.body = "message"),
    3: required ChunkUploadStatusData   data    (api.body = "data"),
}

// ==================== 完成上传 ====================

struct ChunkUploadCompleteReq {
    1: required string upload_id     (api.path = "upload_id"),
    2: required i32    expire_value  (api.body = "expire_value"),
    3: required string expire_style  (api.body = "expire_style"),
    4: required bool   require_auth  (api.body = "require_auth"),
}

struct ChunkUploadCompleteData {
    1: required string share_code (api.body = "share_code"),
    2: required string share_url  (api.body = "share_url"),
    3: required string file_name  (api.body = "file_name"),
    4: required i64    file_size  (api.body = "file_size"),
}

struct ChunkUploadCompleteResp {
    1: required i32                     code    (api.body = "code"),
    2: required string                  message (api.body = "message"),
    3: required ChunkUploadCompleteData data    (api.body = "data"),
}

// ==================== 取消上传 ====================

struct ChunkUploadCancelReq {
    1: required string upload_id (api.path = "upload_id"),
}

struct ChunkUploadCancelResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================

service ChunkService {
    // ChunkUploadInit 初始化分片上传
    ChunkUploadInitResp ChunkUploadInit(1: ChunkUploadInitReq req) (api.post = "/chunk/upload/init/")

    // ChunkUpload 上传单个分片
    ChunkUploadResp ChunkUpload(1: ChunkUploadReq req) (api.post = "/chunk/upload/chunk/:upload_id/:chunk_index")

    // ChunkUploadStatus 查询上传进度
    ChunkUploadStatusResp ChunkUploadStatus(1: ChunkUploadStatusReq req) (api.get = "/chunk/upload/status/:upload_id")

    // ChunkUploadComplete 完成上传
    ChunkUploadCompleteResp ChunkUploadComplete(1: ChunkUploadCompleteReq req) (api.post = "/chunk/upload/complete/:upload_id")

    // ChunkUploadCancel 取消上传
    ChunkUploadCancelResp ChunkUploadCancel(1: ChunkUploadCancelReq req) (api.delete = "/chunk/upload/cancel/:upload_id")
}

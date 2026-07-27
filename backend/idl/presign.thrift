// =====================================================================
// presign.thrift — 预签名上传
// =====================================================================
// 客户端直传对象存储，减轻服务器压力
// 流程: init → 客户端 PUT 直传 → complete 通知服务
// =====================================================================

namespace go presign

// ==================== Init 申请预签名 URL ====================

struct InitReq {
    1: required string file_name    (api.body = "file_name"),
    2: required i64    file_size    (api.body = "file_size"),
    3: required string content_type (api.body = "content_type"),
    4: optional string scheme       (api.body = "scheme"),  // s3/oss/cos 等，留空用默认
    5: optional i32    expire_value  (api.body = "expire_value"),
    6: optional string expire_style  (api.body = "expire_style"),
    7: optional bool   require_auth  (api.body = "require_auth"),
}

struct InitData {
    1: required string upload_id    (api.body = "upload_id"),
    2: required string upload_url   (api.body = "upload_url"),
    3: required string method       (api.body = "method"),     // PUT
    4: required map<string,string> headers (api.body = "headers"), // 需要附加的 header
    5: required i32    expire_seconds (api.body = "expire_seconds"),
    6: required string object_key   (api.body = "object_key"),
    7: required string scheme       (api.body = "scheme"),
    8: required string token        (api.body = "token"),    // 校验令牌
}

struct InitResp {
    1: required i32      code    (api.body = "code"),
    2: required string   message (api.body = "message"),
    3: required InitData data    (api.body = "data"),
}

// ==================== Complete 上传完成通知 ====================

struct CompleteReq {
    1: required string upload_id   (api.body = "upload_id"),
    2: required string token       (api.body = "token"),
    3: optional string object_key  (api.body = "object_key"),
    4: optional string file_hash   (api.body = "file_hash"),
}

struct CompleteData {
    1: required string code         (api.body = "code"),         // share code
    2: required string url          (api.body = "url"),          // 分享 URL
    3: required string file_name    (api.body = "file_name"),
    4: required i64    file_size    (api.body = "file_size"),
    5: required string download_url (api.body = "download_url"),
}

struct CompleteResp {
    1: required i32          code    (api.body = "code"),
    2: required string       message (api.body = "message"),
    3: required CompleteData data    (api.body = "data"),
}

// ==================== Abort 取消 ====================

struct AbortReq {
    1: required string upload_id (api.body = "upload_id"),
    2: required string token     (api.body = "token"),
}

struct AbortResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================

service PresignService {
    // Init 申请预签名上传 URL
    InitResp Init(1: InitReq req) (api.post = "/api/v1/presign/upload")

    // Complete 上传完成后通知服务写 share 表
    CompleteResp Complete(1: CompleteReq req) (api.post = "/api/v1/presign/complete")

    // Abort 取消预签名上传
    AbortResp Abort(1: AbortReq req) (api.post = "/api/v1/presign/abort")
}

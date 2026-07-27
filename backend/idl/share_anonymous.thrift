// =====================================================================
// share_anonymous.thrift — 匿名取件（vastsa/FileCodeBox 核心 UX）
// =====================================================================
// 仿 vastsa: 取件人无账号，按 6 位取件码取文件
// 场景: 工作群发文件 / 老师发资料给家长 / 跨账号文件传递
// =====================================================================

namespace go share_anonymous

// ==================== 生成取件码 ====================

// GenerateCodeReq 上传文件时生成 6 位取件码
struct GenerateCodeReq {
    1: required string file_name    (api.body = "file_name"),
    2: required i64    file_size    (api.body = "file_size"),
    3: optional string expire_value  (api.body = "expire_value"),  // 默认 24
    4: optional string expire_style  (api.body = "expire_style"),  // 默认 hour
    5: optional i32    max_pickup_count (api.body = "max_pickup_count"),  // 0 = 无限
    6: optional string password      (api.body = "password"),     // 可选密码
}

struct GenerateCodeData {
    1: required string code     (api.body = "code"),         // 6 位取件码
    2: required string file_key (api.body = "file_key"),     // 存储 key
    3: required string url      (api.body = "url"),          // 可选下载 URL
    4: required i32    expire_seconds (api.body = "expire_seconds"),
    5: required i32    max_pickup_count (api.body = "max_pickup_count"),
}

struct GenerateCodeResp {
    1: required i32              code    (api.body = "code"),
    2: required string           message (api.body = "message"),
    3: required GenerateCodeData data    (api.body = "data"),
}

// ==================== 按取件码取件 ====================

struct RetrieveReq {
    1: required string code     (api.body = "code"),
    2: optional string password (api.body = "password"),
}

struct RetrieveData {
    1: required string file_name       (api.body = "file_name"),
    2: required i64    file_size       (api.body = "file_size"),
    3: required string content_type    (api.body = "content_type"),
    4: required string download_url    (api.body = "download_url"),
    5: required i32    remaining_count (api.body = "remaining_count"),  // 剩余取件次数
    6: required i64    expire_at       (api.body = "expire_at"),         // Unix 时间戳
    7: required bool   require_password (api.body = "require_password"),
}

struct RetrieveResp {
    1: required i32          code    (api.body = "code"),
    2: required string       message (api.body = "message"),
    3: optional RetrieveData data    (api.body = "data"),
}

// ==================== 下载 ====================

struct DownloadReq {
    1: required string code     (api.path = "code"),
    2: optional string password (api.query = "password"),
}

// DownloadResp 流式响应（用于 JSON 错误响应）
struct DownloadResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 按码查询分享信息（不下载） ====================

struct SearchByCodeReq {
    1: required string code (api.path = "code"),
}

struct SearchByCodeData {
    1: required string file_name     (api.body = "file_name"),
    2: required i64    file_size     (api.body = "file_size"),
    3: required string created_at    (api.body = "created_at"),
    4: required string expire_at     (api.body = "expire_at"),
    5: required i32    pickup_count  (api.body = "pickup_count"),
    6: required i32    max_pickup_count (api.body = "max_pickup_count"),
    7: required bool   require_password (api.body = "require_password"),
}

struct SearchByCodeResp {
    1: required i32             code    (api.body = "code"),
    2: required string          message (api.body = "message"),
    3: optional SearchByCodeData data   (api.body = "data"),
}

// ==================== 服务定义 ====================

service ShareAnonymousService {
    // GenerateCode 生成取件码（上传文件后调用）
    GenerateCodeResp GenerateCode(1: GenerateCodeReq req) (api.post = "/anonymous/generate")

    // Retrieve 按码取件（校验密码 + 返回下载信息）
    RetrieveResp Retrieve(1: RetrieveReq req) (api.post = "/anonymous/retrieve")

    // Download 下载文件
    DownloadResp Download(1: DownloadReq req) (api.get = "/anonymous/download/:code")

    // SearchByCode 按码查询分享信息（不下载）
    SearchByCodeResp SearchByCode(1: SearchByCodeReq req) (api.get = "/anonymous/search/:code")
}

// =====================================================================
// share.thrift — 分享
// =====================================================================
// 迁移自: idl/share.proto
// =====================================================================

namespace go share

// ==================== 分享文本 ====================

struct ShareTextReq {
    1: required string text          (api.body = "text"),
    2: required i32    expire_value  (api.body = "expire_value"),
    3: required string expire_style  (api.body = "expire_style"),  // minute, hour, day, week, month, year, forever
    4: required bool   require_auth  (api.body = "require_auth"),
}

struct ShareData {
    1: required string code (api.body = "code"),
    2: required string url  (api.body = "url"),
}

struct ShareTextResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required ShareData data    (api.body = "data"),
}

// ==================== 分享文件 ====================

struct ShareFileReq {
    // 文件通过 multipart/form-data 上传
    1: required i32    expire_value (api.form = "expire_value"),
    2: required string expire_style (api.form = "expire_style"),
    3: required bool   require_auth (api.form = "require_auth"),
}

struct ShareFileResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required ShareData data    (api.body = "data"),
}

// ==================== 获取分享 ====================

struct GetShareReq {
    1: required string code     (api.path = "code"),
    2: optional string password (api.query = "password"),
}

struct ShareDetail {
    1: required string code         (api.body = "code"),
    2: required string text         (api.body = "text"),
    3: required string file_name    (api.body = "file_name"),
    4: required string file_size    (api.body = "file_size"),
    5: required string url          (api.body = "url"),
    6: required bool   has_password (api.body = "has_password"),
    7: required string expire_time  (api.body = "expire_time"),
}

struct GetShareResp {
    1: required i32         code    (api.body = "code"),
    2: required string      message (api.body = "message"),
    3: required ShareDetail data    (api.body = "data"),
}

// ==================== 下载分享 ====================

struct DownloadFileReq {
    1: required string code     (api.query = "code"),
    2: optional string password (api.query = "password"),
}

struct DownloadFileResp {
    // 文件下载使用流式响应，此消息用于 JSON 错误响应
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================

service ShareService {
    // ShareText 分享文本
    ShareTextResp ShareText(1: ShareTextReq req) (api.post = "/share/text/")

    // ShareFile 分享文件
    ShareFileResp ShareFile(1: ShareFileReq req) (api.post = "/share/file/")

    // GetShare 获取分享内容
    GetShareResp GetShare(1: GetShareReq req) (api.get = "/share/select/")

    // DownloadFile 下载文件
    DownloadFileResp DownloadFile(1: DownloadFileReq req) (api.get = "/share/download")
}

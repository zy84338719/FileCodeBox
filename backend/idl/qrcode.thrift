// =====================================================================
// qrcode.thrift — 二维码生成
// =====================================================================
// 迁移自: idl/qrcode.proto
// =====================================================================

namespace go qrcode

// ==================== 生成二维码 ====================

struct GenerateQRCodeReq {
    1: required string data           (api.body = "data"),
    2: required i32    size           (api.body = "size"),
    3: required bool   return_base64  (api.body = "return_base64"),
}

struct QRCodeData {
    1: required string id          (api.body = "id"),
    2: required string data        (api.body = "data"),
    3: required i32    size        (api.body = "size"),
    4: required string image_url   (api.body = "image_url"),
    5: required string base64_data (api.body = "base64_data"),
}

struct GenerateQRCodeResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required QRCodeData data    (api.body = "data"),
}

// ==================== 获取二维码 ====================

struct GetQRCodeReq {
    1: required string id (api.path = "id"),
}

// ==================== 服务定义 ====================

service QRCodeService {
    // GenerateQRCode 生成二维码
    GenerateQRCodeResp GenerateQRCode(1: GenerateQRCodeReq req) (api.post = "/qrcode/generate")

    // GetQRCode 获取二维码图片
    GetQRCodeReq GetQRCode(1: GetQRCodeReq req) (api.get = "/qrcode/:id")
}

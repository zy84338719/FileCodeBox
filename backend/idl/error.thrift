// =====================================================================
// error.thrift — 统一错误响应 schema
// =====================================================================
// 全站统一响应 envelope: {code, message, data, trace_id}
// 业务码段位规划:
//   0       成功
//   1xxxx   通用错误（参数/认证/限流/资源不存在）
//   2xxxx   业务错误（分享/取件）
//   3xxxx   存储错误
//   4xxxx   用户/认证错误
//   5xxxx   系统/内部错误
//   9xxxx   第三方错误
// =====================================================================

namespace go errcode

// ==================== 统一响应 envelope ====================

struct ErrorInfo {
    1: required i32    code     (api.body = "code"),
    2: required string message  (api.body = "message"),
    3: optional string trace_id (api.body = "trace_id"),
    4: optional string detail   (api.body = "detail"),
    5: optional i64    timestamp (api.body = "timestamp"),
}

struct ErrorResp {
    1: required i32      code    (api.body = "code"),
    2: required string   message (api.body = "message"),
    3: optional ErrorInfo error  (api.body = "error"),
}

// ==================== 业务码常量（仅文档，实际值在 pkg/errcode） ====================

// 1xxxx 通用
// 0    成功
// 10000 未知错误
// 10001 参数错误
// 10002 未登录
// 10003 无权限
// 10004 资源不存在
// 10005 限流
// 10006 服务不可用
// 10007 请求超时
// 10008 内部错误
//
// 2xxxx 分享/取件
// 20001 分享不存在
// 20002 分享过期
// 20003 密码错误
// 20004 超过取件次数
// 20005 取件码不存在
// 20006 取件码已过期
// 20007 取件码已用完
// 20008 文件不存在
//
// 3xxxx 存储
// 30001 存储初始化失败
// 30002 存储配额超限
// 30003 上传失败
// 30004 下载失败
// 30005 删除失败
// 30006 存储方案不支持
//
// 4xxxx 用户/认证
// 40001 用户不存在
// 40002 用户已存在
// 40003 密码错误
// 40004 用户已禁用
// 40005 Token 无效
// 40006 Token 过期
// 40007 API Key 无效
// 40008 权限不足
//
// 5xxxx 系统
// 50001 数据库错误
// 50002 缓存错误
// 50003 配置错误
// 50004 调度错误
//
// 9xxxx 第三方
// 90001 短信发送失败
// 90002 邮件发送失败
// 90003 第三方 API 失败

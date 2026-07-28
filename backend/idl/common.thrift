// =====================================================================
// common.thrift — 通用服务（index/ping）
// =====================================================================
// 迁移自: idl/common.proto
// 迁移日期: 2026-07-27
// 字段 1:1 保留 proto 编号（API 兼容）
// =====================================================================
// 注意：健康检查 /health 统一由 health.thrift 的 HealthService 提供
// （/health、/ready、/live、/version、/ping）。本服务历史上也注册过
// /health，与 health.thrift 重复，会导致 Hertz 启动 panic
// （"handlers are already registered for path '/health'"），故移除。
// 本服务仅保留首页 Index (/) 与 Ping (/api/v1/ping)。

namespace go common

// ==================== 通用响应 ====================

struct EmptyReq {
}

// HealthResp 健康状态响应
struct HealthResp {
    1: required string status (api.body = "status"),
}

// IndexResp 服务基本信息
struct IndexResp {
    1: required string name    (api.body = "name"),
    2: required string version (api.body = "version"),
    3: required string status  (api.body = "status"),
}

// PingResp Ping 响应
struct PingResp {
    1: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================
//
// 注意：根路径 "/" 不在此注册。前端 SPA 首页由 bootstrap.customizedRegister
// 直接服务 static/index.html，避免与 API 路由冲突。本服务仅提供 Ping。

service CommonService {
    // Ping Ping 检查
    PingResp Ping(1: EmptyReq req) (api.get = "/api/v1/ping")
}

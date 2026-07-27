// =====================================================================
// common.thrift — 通用服务（health/index/ping）
// =====================================================================
// 迁移自: idl/common.proto
// 迁移日期: 2026-07-27
// 字段 1:1 保留 proto 编号（API 兼容）
// =====================================================================

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

service CommonService {
    // Health 健康检查
    HealthResp Health(1: EmptyReq req) (api.get = "/health")

    // Index 首页
    IndexResp Index(1: EmptyReq req) (api.get = "/")

    // Ping Ping 检查
    PingResp Ping(1: EmptyReq req) (api.get = "/api/v1/ping")
}

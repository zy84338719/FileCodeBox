// =====================================================================
// health.thrift — 健康检查（独立子目录版，与 common.thrift 互补）
// =====================================================================
// 迁移自: idl/http/health.proto
// common.thrift 提供 health/index/ping，本文件提供 readiness/liveness/version
// =====================================================================

namespace go health

// ==================== 健康检查 ====================

struct HealthCheckReq {
}

struct HealthCheckResp {
    1: required string status    (api.body = "status"),
    2: required string timestamp (api.body = "timestamp"),
}

// ==================== 就绪检查 ====================

struct ReadinessReq {
}

struct ReadinessResp {
    1: required string status   (api.body = "status"),
    2: required bool   database (api.body = "database"),
    3: required bool   redis    (api.body = "redis"),
}

// ==================== 存活检查 ====================

struct LivenessReq {
}

struct LivenessResp {
    1: required string status (api.body = "status"),
}

// ==================== 版本信息 ====================

struct VersionReq {
}

struct VersionResp {
    1: required string name       (api.body = "name"),
    2: required string version    (api.body = "version"),
    3: required string build_time (api.body = "build_time"),
    4: required string git_commit (api.body = "git_commit"),
    5: required string go_version (api.body = "go_version"),
}

// ==================== Ping/Pong ====================

struct PingReq {
}

struct PingResp {
    1: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================

service HealthService {
    // Health 健康检查
    HealthCheckResp Health(1: HealthCheckReq req) (api.get = "/health")

    // Readiness 就绪检查（K8s readinessProbe）
    ReadinessResp Readiness(1: ReadinessReq req) (api.get = "/ready")

    // Liveness 存活检查（K8s livenessProbe）
    LivenessResp Liveness(1: LivenessReq req) (api.get = "/live")

    // Version 版本信息
    VersionResp Version(1: VersionReq req) (api.get = "/version")

    // Ping Ping 检查
    PingResp Ping(1: PingReq req) (api.get = "/ping")
}

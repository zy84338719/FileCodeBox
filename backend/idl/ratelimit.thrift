// =====================================================================
// ratelimit.thrift — 限流管理
// =====================================================================
// IP 维度 + 接口维度双层限流
// 配置存 runtime_configs 表，运行时可调
// =====================================================================

namespace go ratelimit

// ==================== 限流配置 ====================

struct RateLimitConfig {
    1: required i32  global_qps     (api.body = "global_qps"),     // 全局 IP 维度 QPS
    2: required i32  upload_qps     (api.body = "upload_qps"),     // 上传接口 QPS（per IP）
    3: required i32  download_qps   (api.body = "download_qps"),   // 下载接口 QPS（per IP）
    4: required i32  login_qps      (api.body = "login_qps"),      // 登录接口 QPS（per IP）
    5: required i32  burst          (api.body = "burst"),          // 突发容量
    6: required bool enabled        (api.body = "enabled"),        // 总开关
    7: required bool block_on_limit (api.body = "block_on_limit"), // true: 429 阻断; false: 放过
    8: required i32  block_seconds  (api.body = "block_seconds"),  // 阻断持续秒数
}

// ==================== 获取配置 ====================

struct GetConfigReq {
}

struct GetConfigResp {
    1: required i32             code    (api.body = "code"),
    2: required string          message (api.body = "message"),
    3: required RateLimitConfig data    (api.body = "data"),
}

// ==================== 更新配置 ====================

struct UpdateConfigReq {
    1: required RateLimitConfig config (api.body = "config"),
}

struct UpdateConfigResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 压测 ====================

struct TestReq {
    1: required string scope     (api.body = "scope"),    // global/upload/download/login
    2: required i32    qps       (api.body = "qps"),      // 目标 QPS
    3: required i32    duration  (api.body = "duration"), // 秒
    4: optional string client_ip (api.body = "client_ip"), // 模拟客户端 IP
}

struct TestData {
    1: required i32    total_requests     (api.body = "total_requests"),
    2: required i32    success_count      (api.body = "success_count"),
    3: required i32    blocked_count      (api.body = "blocked_count"),
    4: required double actual_qps         (api.body = "actual_qps"),
    5: required double p50_latency_ms     (api.body = "p50_latency_ms"),
    6: required double p99_latency_ms     (api.body = "p99_latency_ms"),
    7: required i32    limit_triggered_at (api.body = "limit_triggered_at"),
}

struct TestResp {
    1: required i32     code    (api.body = "code"),
    2: required string  message (api.body = "message"),
    3: required TestData data    (api.body = "data"),
}

// ==================== 实时状态 ====================

struct StatusReq {
}

struct StatusData {
    1: required i32    active_clients (api.body = "active_clients"),
    2: required double current_qps   (api.body = "current_qps"),
    3: required i32    blocked_ips   (api.body = "blocked_ips"),
    4: required RateLimitConfig config (api.body = "config"),
}

struct StatusResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required StatusData data    (api.body = "data"),
}

// ==================== 服务定义 ====================

service RatelimitService {
    // GetConfig 获取限流配置
    GetConfigResp GetConfig(1: GetConfigReq req) (api.get = "/admin/ratelimit/config")

    // UpdateConfig 更新限流配置
    UpdateConfigResp UpdateConfig(1: UpdateConfigReq req) (api.put = "/admin/ratelimit/config")

    // Test 压力测试
    TestResp Test(1: TestReq req) (api.post = "/admin/ratelimit/test")

    // Status 实时状态
    StatusResp Status(1: StatusReq req) (api.get = "/admin/ratelimit/status")
}

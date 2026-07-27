// =====================================================================
// setup.thrift — 系统初始化
// =====================================================================
// 迁移自: idl/setup.proto
// =====================================================================

namespace go setup

// ==================== 系统初始化检查 ====================

struct CheckInitializationReq {
}

struct CheckInitializationResp {
    1: required bool   initialized (api.body = "initialized"),
    2: required string message    (api.body = "message"),
}

// ==================== 系统初始化 ====================

struct BaseConfig {
    1: required string name        (api.body = "name"),
    2: required string description (api.body = "description"),
    3: required i32    port        (api.body = "port"),
    4: required string host        (api.body = "host"),
}

struct InitializeReq {
    1: required string     admin_username (api.body = "admin_username"),
    2: required string     admin_password (api.body = "admin_password"),
    3: required string     admin_email    (api.body = "admin_email"),
    4: optional BaseConfig base_config   (api.body = "base_config"),
}

struct InitializeResp {
    1: required string message  (api.body = "message"),
    2: required string username (api.body = "username"),
}

// ==================== 服务定义 ====================

service SetupService {
    // CheckInitialization 检查系统是否已初始化
    CheckInitializationResp CheckInitialization(1: CheckInitializationReq req) (api.get = "/setup/check")

    // Initialize 初始化系统
    InitializeResp Initialize(1: InitializeReq req) (api.post = "/setup")
}

// =====================================================================
// maintenance.thrift — 维护
// =====================================================================
// 迁移自: idl/maintenance.proto
// =====================================================================

namespace go maintenance

// ==================== 清理过期文件 ====================

struct CleanExpiredFilesReq {
}

struct CleanExpiredFilesData {
    1: required i64 deleted_count (api.body = "deleted_count"),
    2: required i64 freed_space   (api.body = "freed_space"),
}

struct CleanExpiredFilesResp {
    1: required i32                   code    (api.body = "code"),
    2: required string                message (api.body = "message"),
    3: required CleanExpiredFilesData data    (api.body = "data"),
}

// ==================== 清理临时文件 ====================

struct CleanTempFilesReq {
}

struct CleanTempFilesData {
    1: required i64 deleted_count (api.body = "deleted_count"),
    2: required i64 freed_space   (api.body = "freed_space"),
}

struct CleanTempFilesResp {
    1: required i32                code    (api.body = "code"),
    2: required string             message (api.body = "message"),
    3: required CleanTempFilesData data    (api.body = "data"),
}

// ==================== 系统信息 ====================

struct GetSystemInfoReq {
}

struct SystemInfoData {
    1: required string version       (api.body = "version"),
    2: required string os            (api.body = "os"),
    3: required string arch          (api.body = "arch"),
    4: required string uptime        (api.body = "uptime"),
    5: required i64    goroutines    (api.body = "goroutines"),
    6: required i64    memory_alloc  (api.body = "memory_alloc"),
    7: required i64    memory_total  (api.body = "memory_total"),
    8: required i64    memory_sys    (api.body = "memory_sys"),
}

struct GetSystemInfoResp {
    1: required i32            code    (api.body = "code"),
    2: required string         message (api.body = "message"),
    3: required SystemInfoData data    (api.body = "data"),
}

// ==================== 存储状态 ====================

struct GetStorageStatusReq {
}

struct StorageStatusData {
    1: required string storage_type   (api.body = "storage_type"),
    2: required i64    total_space    (api.body = "total_space"),
    3: required i64    used_space     (api.body = "used_space"),
    4: required i64    free_space     (api.body = "free_space"),
    5: required i64    file_count     (api.body = "file_count"),
    6: required double usage_percent  (api.body = "usage_percent"),
}

struct GetStorageStatusResp {
    1: required i32              code    (api.body = "code"),
    2: required string           message (api.body = "message"),
    3: required StorageStatusData data   (api.body = "data"),
}

// ==================== 系统日志 ====================

struct GetSystemLogsReq {
    1: optional string level     (api.query = "level"),
    2: optional i32    page      (api.query = "page"),
    3: optional i32    page_size (api.query = "page_size"),
}

struct LogEntry {
    1: required i64    id         (api.body = "id"),
    2: required string level      (api.body = "level"),
    3: required string message    (api.body = "message"),
    4: required string created_at (api.body = "created_at"),
    5: required string module     (api.body = "module"),
    6: required string user_id    (api.body = "user_id"),
}

struct SystemLogsData {
    1: required list<LogEntry> logs      (api.body = "logs"),
    2: required i64            total     (api.body = "total"),
    3: required i32            page      (api.body = "page"),
    4: required i32            page_size (api.body = "page_size"),
}

struct GetSystemLogsResp {
    1: required i32            code    (api.body = "code"),
    2: required string         message (api.body = "message"),
    3: required SystemLogsData data    (api.body = "data"),
}

// ==================== 服务定义 ====================

service MaintenanceService {
    // CleanExpiredFiles 清理过期文件
    CleanExpiredFilesResp CleanExpiredFiles(1: CleanExpiredFilesReq req) (api.post = "/admin/maintenance/clean-expired")

    // CleanTempFiles 清理临时文件
    CleanTempFilesResp CleanTempFiles(1: CleanTempFilesReq req) (api.post = "/admin/maintenance/clean-temp")

    // GetSystemInfo 获取系统信息
    GetSystemInfoResp GetSystemInfo(1: GetSystemInfoReq req) (api.get = "/admin/maintenance/system-info")

    // GetStorageStatus 获取存储状态
    GetStorageStatusResp GetStorageStatus(1: GetStorageStatusReq req) (api.get = "/admin/maintenance/monitor/storage")

    // GetSystemLogs 获取系统日志
    GetSystemLogsResp GetSystemLogs(1: GetSystemLogsReq req) (api.get = "/admin/maintenance/logs")
}

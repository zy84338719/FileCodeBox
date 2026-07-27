// =====================================================================
// storage.thrift — 存储管理（OpenDAL 配置）
// =====================================================================
// 迁移自: idl/storage.proto
// 改造点: storage 改用 OpenDAL 风格 scheme + options
// =====================================================================

namespace go storage

// ==================== 存储详情 ====================

struct WebDAVConfig {
    1: required string hostname   (api.body = "hostname"),
    2: required string username   (api.body = "username"),
    3: required string password   (api.body = "password"),
    4: required string root_path  (api.body = "root_path"),
    5: required string url        (api.body = "url"),
}

struct S3Config {
    1: required string access_key_id     (api.body = "access_key_id"),
    2: required string secret_access_key (api.body = "secret_access_key"),
    3: required string bucket_name       (api.body = "bucket_name"),
    4: required string endpoint_url      (api.body = "endpoint_url"),
    5: required string region_name       (api.body = "region_name"),
    6: required string hostname          (api.body = "hostname"),
    7: required string proxy             (api.body = "proxy"),
}

struct NFSConfig {
    1: required string server      (api.body = "server"),
    2: required string path        (api.body = "path"),
    3: required string mount_point (api.body = "mount_point"),
    4: required string version     (api.body = "version"),
    5: required string options     (api.body = "options"),
    6: required i32    timeout     (api.body = "timeout"),
    7: required i32    auto_mount  (api.body = "auto_mount"),
    8: required i32    retry_count (api.body = "retry_count"),
    9: required string sub_path    (api.body = "sub_path"),
}

struct StorageDetail {
    1: required string type          (api.body = "type"),
    2: required bool   available     (api.body = "available"),
    3: required string storage_path  (api.body = "storage_path"),
    4: required i32    usage_percent (api.body = "usage_percent"),
    5: required string error         (api.body = "error"),
}

struct StorageConfig {
    1: required string         type         (api.body = "type"),
    2: required string         storage_path (api.body = "storage_path"),
    3: required WebDAVConfig   webdav       (api.body = "webdav"),
    4: required S3Config       s3           (api.body = "s3"),
    5: required NFSConfig      nfs          (api.body = "nfs"),
}

// ==================== 获取存储信息 ====================

struct GetStorageInfoReq {
}

struct StorageInfoData {
    1: required string                    current          (api.body = "current"),           // 当前存储类型
    2: required list<string>              available        (api.body = "available"),         // 可用存储类型列表
    3: required map<string,StorageDetail> storage_details  (api.body = "storage_details"),   // 各存储类型详细信息
    4: required StorageConfig             storage_config   (api.body = "storage_config"),     // 存储配置
}

struct GetStorageInfoResp {
    1: required i32              code    (api.body = "code"),
    2: required string           message (api.body = "message"),
    3: required StorageInfoData  data    (api.body = "data"),
}

// ==================== 切换存储 ====================

struct SwitchStorageReq {
    1: required string type (api.body = "type"),
}

struct SwitchStorageData {
    1: required bool   success      (api.body = "success"),
    2: required string message      (api.body = "message"),
    3: required string current_type (api.body = "current_type"),
}

struct SwitchStorageResp {
    1: required i32              code    (api.body = "code"),
    2: required string           message (api.body = "message"),
    3: required SwitchStorageData data    (api.body = "data"),
}

// ==================== 测试存储连接 ====================

struct TestStorageConnectionReq {
    1: required string type (api.path = "type"),
}

struct TestStorageData {
    1: required string type   (api.body = "type"),
    2: required string status (api.body = "status"),
}

struct TestStorageConnectionResp {
    1: required i32              code    (api.body = "code"),
    2: required string           message (api.body = "message"),
    3: required TestStorageData  data    (api.body = "data"),
}

// ==================== 更新存储配置 ====================

struct UpdateConfig {
    1: required string       storage_path (api.body = "storage_path"),
    2: required WebDAVConfig webdav       (api.body = "webdav"),
    3: required S3Config     s3           (api.body = "s3"),
    4: required NFSConfig    nfs          (api.body = "nfs"),
}

struct UpdateStorageConfigReq {
    1: required string       type   (api.body = "type"),
    2: required UpdateConfig config (api.body = "config"),
}

struct UpdateStorageConfigResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================

service StorageService {
    // GetStorageInfo 获取存储信息
    GetStorageInfoResp GetStorageInfo(1: GetStorageInfoReq req) (api.get = "/admin/storage")

    // SwitchStorage 切换存储
    SwitchStorageResp SwitchStorage(1: SwitchStorageReq req) (api.post = "/admin/storage/switch")

    // TestStorageConnection 测试存储连接
    TestStorageConnectionResp TestStorageConnection(1: TestStorageConnectionReq req) (api.get = "/admin/storage/test/:type")

    // UpdateStorageConfig 更新存储配置
    UpdateStorageConfigResp UpdateStorageConfig(1: UpdateStorageConfigReq req) (api.put = "/admin/storage/config")
}

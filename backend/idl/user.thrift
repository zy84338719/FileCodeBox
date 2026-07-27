// =====================================================================
// user.thrift — 用户模块
// =====================================================================
// 迁移自: idl/user.proto
// =====================================================================

namespace go user

// ==================== 用户注册 ====================

struct RegisterReq {
    1: required string username (api.body = "username"),
    2: required string email    (api.body = "email"),
    3: required string password (api.body = "password"),
    4: required string nickname (api.body = "nickname"),
}

struct UserData {
    1: required i64    id         (api.body = "id"),
    2: required string username   (api.body = "username"),
    3: required string email      (api.body = "email"),
    4: required string nickname   (api.body = "nickname"),
    5: required string avatar     (api.body = "avatar"),
    6: required i32    status     (api.body = "status"),
    7: required string created_at (api.body = "created_at"),
}

struct RegisterResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required UserData  data    (api.body = "data"),
}

// ==================== 用户登录 ====================

struct LoginReq {
    1: required string username (api.body = "username"),
    2: required string password (api.body = "password"),
}

struct LoginData {
    1: required string   token (api.body = "token"),
    2: required UserData user  (api.body = "user"),
}

struct LoginResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required LoginData data    (api.body = "data"),
}

// ==================== 用户信息 ====================

struct UserInfoReq {
}

struct UserInfoResp {
    1: required i32      code    (api.body = "code"),
    2: required string   message (api.body = "message"),
    3: required UserData data    (api.body = "data"),
}

// ==================== 用户资料更新 ====================

struct UpdateProfileReq {
    1: required string nickname (api.body = "nickname"),
    2: required string avatar   (api.body = "avatar"),
}

struct UpdateProfileResp {
    1: required i32      code    (api.body = "code"),
    2: required string   message (api.body = "message"),
    3: required UserData data    (api.body = "data"),
}

// ==================== 修改密码 ====================

struct ChangePasswordReq {
    1: required string old_password (api.body = "old_password"),
    2: required string new_password (api.body = "new_password"),
}

struct ChangePasswordResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 用户统计信息 ====================

struct UserStatsReq {
}

struct UserStats {
    1: required i64 total_uploads (api.body = "total_uploads"),
    2: required i64 total_size    (api.body = "total_size"),
    3: required i64 quota_used    (api.body = "quota_used"),
    4: required i64 quota_limit   (api.body = "quota_limit"),
}

struct UserStatsResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required UserStats data    (api.body = "data"),
}

// ==================== 空请求 ====================

struct EmptyReq {
}

// ==================== 用户文件列表 ====================

struct UserFilesReq {
    1: required i32 page      (api.query = "page"),
    2: required i32 page_size (api.query = "page_size"),
}

struct UserFileItem {
    1: required i64    id            (api.body = "id"),
    2: required string code          (api.body = "code"),
    3: required string prefix        (api.body = "prefix"),
    4: required string suffix        (api.body = "suffix"),
    5: required string file_name     (api.body = "file_name"),
    6: required string file_path     (api.body = "file_path"),
    7: required i64    size          (api.body = "size"),
    8: required string expired_at    (api.body = "expired_at"),
    9: required i32    expired_count (api.body = "expired_count"),
    10: required i32   used_count    (api.body = "used_count"),
    11: required string created_at   (api.body = "created_at"),
    12: required string updated_at   (api.body = "updated_at"),
}

struct UserFilePagination {
    1: required i32    page        (api.body = "page"),
    2: required i32    page_size   (api.body = "page_size"),
    3: required i64    total       (api.body = "total"),
    4: required i32    total_pages (api.body = "total_pages"),
    5: required bool   has_next    (api.body = "has_next"),
    6: required bool   has_prev    (api.body = "has_prev"),
}

struct UserFileList {
    1: required list<UserFileItem>     files      (api.body = "files"),
    2: required UserFilePagination     pagination (api.body = "pagination"),
}

struct UserFilesResp {
    1: required i32          code    (api.body = "code"),
    2: required string       message (api.body = "message"),
    3: required UserFileList data    (api.body = "data"),
}

// ==================== API Key 管理 ====================

struct APIKeyData {
    1: required i64    id           (api.body = "id"),
    2: required string name         (api.body = "name"),
    3: required string prefix       (api.body = "prefix"),
    4: required string last_used_at (api.body = "last_used_at"),
    5: required string expires_at   (api.body = "expires_at"),
    6: required string created_at   (api.body = "created_at"),
    7: required bool   revoked      (api.body = "revoked"),
}

struct APIKeyListResp {
    1: required i32              code    (api.body = "code"),
    2: required string           message (api.body = "message"),
    3: required list<APIKeyData> keys    (api.body = "keys"),
}

struct CreateAPIKeyReq {
    1: required string name             (api.body = "name"),
    2: optional string expires_at       (api.body = "expires_at"),
    3: optional i64    expires_in_days  (api.body = "expires_in_days"),
}

struct CreateAPIKeyData {
    1: required string    key     (api.body = "key"),
    2: required APIKeyData api_key (api.body = "api_key"),
}

struct CreateAPIKeyResp {
    1: required i32             code    (api.body = "code"),
    2: required string          message (api.body = "message"),
    3: required CreateAPIKeyData data    (api.body = "data"),
}

struct DeleteAPIKeyReq {
    1: required i64 id (api.path = "id"),
}

struct DeleteAPIKeyResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================

service UserService {
    // Register 用户注册
    RegisterResp Register(1: RegisterReq req) (api.post = "/user/register")

    // Login 用户登录
    LoginResp Login(1: LoginReq req) (api.post = "/user/login")

    // UserInfo 获取用户信息
    UserInfoResp UserInfo(1: UserInfoReq req) (api.get = "/user/info")

    // UpdateProfile 更新用户资料
    UpdateProfileResp UpdateProfile(1: UpdateProfileReq req) (api.put = "/user/profile")

    // ChangePassword 修改密码
    ChangePasswordResp ChangePassword(1: ChangePasswordReq req) (api.post = "/user/change-password")

    // UserStats 用户统计信息
    UserStatsResp UserStats(1: UserStatsReq req) (api.get = "/user/stats")

    // UserFiles 获取用户文件列表
    UserFilesResp UserFiles(1: UserFilesReq req) (api.get = "/user/files")

    // ListAPIKeys 获取 API Key 列表
    APIKeyListResp ListAPIKeys(1: EmptyReq req) (api.get = "/user/api-keys")

    // CreateAPIKey 创建 API Key
    CreateAPIKeyResp CreateAPIKey(1: CreateAPIKeyReq req) (api.post = "/user/api-keys")

    // DeleteAPIKey 删除 API Key
    DeleteAPIKeyResp DeleteAPIKey(1: DeleteAPIKeyReq req) (api.delete = "/user/api-keys/:id")
}

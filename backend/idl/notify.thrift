// =====================================================================
// notify.thrift — 系统通知（管理员公告）
// =====================================================================
// 管理员发布公告，前端首页/管理面板 banner 显示
// 支持时间窗口、级别、类型
// =====================================================================

namespace go notify

// ==================== 通用 ====================

// NotifyType 通知类型
struct NotifyType {
    1: required string value (api.body = "value"),  // system / feature / maintenance
}

// NotifyLevel 通知级别
struct NotifyLevel {
    1: required string value (api.body = "value"),  // info / warning / error / success
}

// ==================== 通知数据 ====================

struct NotifyItem {
    1: required i64    id         (api.body = "id"),
    2: required string title      (api.body = "title"),
    3: required string content    (api.body = "content"),
    4: required string type       (api.body = "type"),        // system/feature/maintenance
    5: required string level      (api.body = "level"),       // info/warning/error/success
    6: required i32    status     (api.body = "status"),      // 0=草稿 1=发布 2=下线
    7: optional string start_at   (api.body = "start_at"),    // 生效时间（RFC3339）
    8: optional string end_at     (api.body = "end_at"),      // 失效时间
    9: required string created_at (api.body = "created_at"),
    10: required string updated_at (api.body = "updated_at"),
}

// ==================== 列表查询 ====================

struct ListReq {
    1: optional i32    page      (api.query = "page"),
    2: optional i32    page_size (api.query = "page_size"),
    3: optional string type      (api.query = "type"),
    4: optional string level     (api.query = "level"),
    5: optional i32    status    (api.query = "status"),
}

struct ListData {
    1: required list<NotifyItem> items     (api.body = "items"),
    2: required i64              total     (api.body = "total"),
    3: required i32              page      (api.body = "page"),
    4: required i32              page_size (api.body = "page_size"),
}

struct ListResp {
    1: required i32      code    (api.body = "code"),
    2: required string   message (api.body = "message"),
    3: required ListData data    (api.body = "data"),
}

// ==================== 获取活跃通知（公开 API） ====================

struct ActiveReq {
    1: optional string type (api.query = "type"),
}

struct ActiveData {
    1: required list<NotifyItem> items (api.body = "items"),
}

struct ActiveResp {
    1: required i32        code    (api.body = "code"),
    2: required string     message (api.body = "message"),
    3: required ActiveData data    (api.body = "data"),
}

// ==================== 获取单条 ====================

struct GetReq {
    1: required i64 id (api.path = "id"),
}

struct GetResp {
    1: required i32       code    (api.body = "code"),
    2: required string    message (api.body = "message"),
    3: required NotifyItem data   (api.body = "data"),
}

// ==================== 创建 ====================

struct CreateReq {
    1: required string title    (api.body = "title"),
    2: required string content  (api.body = "content"),
    3: required string type     (api.body = "type"),
    4: required string level    (api.body = "level"),
    5: optional i32    status   (api.body = "status"),
    6: optional string start_at (api.body = "start_at"),
    7: optional string end_at   (api.body = "end_at"),
}

struct CreateResp {
    1: required i32        code    (api.body = "code"),
    2: required string     message (api.body = "message"),
    3: required NotifyItem data    (api.body = "data"),
}

// ==================== 更新 ====================

struct UpdateReq {
    1: required i64    id       (api.path = "id"),
    2: optional string title    (api.body = "title"),
    3: optional string content  (api.body = "content"),
    4: optional string type     (api.body = "type"),
    5: optional string level    (api.body = "level"),
    6: optional i32    status   (api.body = "status"),
    7: optional string start_at (api.body = "start_at"),
    8: optional string end_at   (api.body = "end_at"),
}

struct UpdateResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 删除 ====================

struct DeleteReq {
    1: required i64 id (api.path = "id"),
}

struct DeleteResp {
    1: required i32    code    (api.body = "code"),
    2: required string message (api.body = "message"),
}

// ==================== 服务定义 ====================

service NotifyService {
    // List 通知列表（管理）
    ListResp List(1: ListReq req) (api.get = "/admin/notifies")

    // Active 当前活跃通知（公开）
    ActiveResp Active(1: ActiveReq req) (api.get = "/notifies/active")

    // Get 获取单条
    GetResp Get(1: GetReq req) (api.get = "/admin/notifies/:id")

    // Create 创建
    CreateResp Create(1: CreateReq req) (api.post = "/admin/notifies")

    // Update 更新
    UpdateResp Update(1: UpdateReq req) (api.put = "/admin/notifies/:id")

    // Delete 删除
    DeleteResp Delete(1: DeleteReq req) (api.delete = "/admin/notifies/:id")
}

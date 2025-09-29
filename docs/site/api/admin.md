# 管理类 API

管理接口需要管理员 Token（`Authorization: Bearer <token>`）。可用于构建自动化运维脚本、审计平台或自定义控制台。

## 登录

`POST /admin/login`

```json
{
  "username": "admin",
  "password": "StrongPassw0rd!"
}
```

返回

```json
{
  "code": 200,
  "data": {
    "token": "<jwt>",
    "expires_in": 259200
  }
}
```

## 仪表盘数据

`GET /admin/dashboard`

返回最近 24 小时/7 天的上传、下载统计以及系统信息。

## 文件管理

### 列表

`GET /admin/files?page=1&size=20&status=active`

返回字段包括：

- `id`, `filename`, `size`, `uploader`
- `share_code`, `expire_at`, `download_count`
- `storage_provider`, `created_at`

### 删除 / 撤销

`POST /admin/files/delete`

```json
{
  "ids": ["f_1001", "f_1002"],
  "hard_delete": false
}
```

- `hard_delete=true` 时会同时删除存储后端文件

## 用户管理

`GET /admin/users`

`POST /admin/users`

```json
{
  "username": "alice",
  "password": "P@ssword!",
  "role": "user",
  "quota": 1073741824
}
```

`PATCH /admin/users/:id`

支持禁用 / 重置密码 / 调整配额。

## 配置管理

`GET /admin/config`

返回当前配置（已合并文件、环境变量、数据库）。

`PUT /admin/config`

```json
{
  "transfer": {
    "upload": {
      "upload_size": 52428800
    }
  }
}
```

更新成功后会触发 ConfigManager 热刷新。

## API Key

`POST /admin/api-keys`

```json
{
  "name": "ci-pipeline",
  "scopes": ["files:read", "files:write"],
  "expire_type": "days",
  "expire_value": 30
}
```

返回一次性明文 Key，需立即保存。

`DELETE /admin/api-keys/:id`

撤销 API Key。

## 系统任务

- `POST /admin/tasks/cleanup` — 立即执行过期文件清理
- `POST /admin/tasks/rebuild-stats` — 重建统计信息

## 响应字段说明

| 字段 | 说明 |
| --- | --- |
| `code` | 业务状态码 |
| `message` | 简要描述 |
| `data` | 业务数据 |
| `trace_id` | （可选）链路追踪 ID |

## 安全建议

- 对 `/admin` 接口启用 HTTPS + IP 白名单
- 定期轮换管理员 Token 或改用 API Key 模式
- 为自动化脚本分配最小权限的 Scope

更多接口请参考 Swagger 文档或直接阅读 `internal/handlers` 目录下的实现。

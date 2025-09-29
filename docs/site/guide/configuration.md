# 系统配置详解

FileCodeBox 使用 `ConfigManager` 管理配置：**文件 → 环境变量 → 数据库** 三层覆盖。以下介绍常见配置项及最佳实践。

## 配置加载顺序

1. 启动时读取 `config.yaml`
2. 读取环境变量（`FILECODEBOX__BASE__PORT` 等）覆盖
3. 初始化完成后写入数据库 `key_value` 表
4. 后台或 API 修改配置会同步刷新内存缓存并写回数据库

## 配置示例

```yaml
base:
  name: 文件快递柜
  description: "安全、便捷的文件快传平台"
  port: 12345
  data_path: "./data"

transfer:
  upload:
    open_upload: 1
    upload_size: 10485760      # 单文件 10 MiB，0 为无限制
    enable_chunk: 1
    chunk_size: 2097152        # 2 MiB 分片
  download:
    enable_concurrent_download: 1
    max_concurrent_downloads: 10
    require_login: 0

user:
  allow_user_registration: 1
  session_expiry_hours: 72

ui:
  themes_select: "themes/2025"
  opacity: 0.95
  show_admin_addr: 1

storage:
  type: "local"

mcp:
  enable_mcp_server: 0
```

## 基础配置（`base`）

| 字段 | 作用 |
| --- | --- |
| `name` / `description` | 页面标题与描述 |
| `port` / `host` | 服务监听地址 |
| `data_path` | 数据目录（数据库/上传） |
| `production` | 控制日志级别、静态资源缓存策略 |

## 上传与下载（`transfer`）

- `upload_size`：单次上传大小限制（字节），值为 0 表示不限制
- `chunk_size`：分片大小，推荐 1~5 MiB
- `max_save_seconds`：临时片段自动清理时间
- `require_login`：上传/下载是否强制登录

## 用户系统（`user`）

- `allow_user_registration`：允许用户自助注册
- `require_email_verify`：是否启用邮箱验证流程
- `user_storage_quota`：单用户存储配额
- `max_sessions_per_user`：限制同一账号同时在线的设备数量

## UI 与主题（`ui`）

- `themes_select`：当前主题路径，默认 `themes/2025`
- `notify_title` / `notify_content`：登录后的公告
- `page_explain`：首页底部说明文字
- `background`：自定义背景图 URL

## 存储（`storage`）

参见 [存储适配指南](./storage.md)。切换存储需填写对应凭证，否则健康检查会失败。

## MCP 服务器（`mcp`）

- `enable_mcp_server`：是否启用 Model Context Protocol Server
- `mcp_host` / `mcp_port`：监听地址
- 启用后可通过脚本 `scripts/test_mcp_client.py` 进行联调

## 环境变量速查

| 环境变量 | 对应配置 |
| --- | --- |
| `FILECODEBOX__BASE__PORT=8080` | `base.port` |
| `FILECODEBOX__STORAGE__TYPE=s3` | `storage.type` |
| `FILECODEBOX__TRANSFER__UPLOAD__UPLOAD_SIZE=52428800` | 上传大小限制 |

> 命名规则：将 YAML 层级转换为大写并使用 `__` 连接。

## 在线修改

- 管理后台 → 系统设置
- API：`PUT /admin/config`（需管理员 Token）
- 修改后 ConfigManager 会自动刷新缓存并持久化

## 配置备份

- `scripts/export_config_from_db.go`：从数据库导出配置为 YAML
- `scripts/export_config_from_db.py`：Python 版本脚本

## 常见问题

| 问题 | 解决方案 |
| --- | --- |
| 配置修改后未生效 | 确认是否存在更高优先级的环境变量或数据库配置覆盖 |
| 配置格式错误导致启动失败 | 检查日志，或先恢复上一次导出的配置文件 |
| 需要回滚配置 | 使用 ConfigManager 的历史记录（开发中），或手动导入备份

下一章节：[安全加固最佳实践](./security.md)

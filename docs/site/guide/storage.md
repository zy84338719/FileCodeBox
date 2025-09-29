# 存储适配指南

FileCodeBox 采用可插拔的 Storage Manager。只需在配置中切换存储类型即可完成迁移。本章介绍常见存储方案与注意事项。

## 配置入口

- 初始化向导：部署后首次访问 `/setup`
- 管理后台：系统设置 → 存储配置
- 配置文件：`config.yaml` → `storage` 段

完整示例可参考仓库 `docs/config.new.yaml`。

## 本地存储（默认）

```yaml
storage:
  type: local
  local:
    path: "./uploads"
```

- 文件保存在服务端磁盘上（默认为 `data/uploads`）
- 适用于单机部署或小规模团队使用
- 建议结合定期备份脚本，将目录同步至对象存储或 NAS

## S3 / MinIO / OSS

```yaml
storage:
  type: s3
  s3:
    endpoint: "https://s3.your-cloud.com"
    access_key: "AK..."
    secret_key: "SK..."
    bucket: "filecodebox"
    region: "ap-southeast-1"
```

- 支持任何兼容 S3 API 的厂商，如 AWS S3、阿里云 OSS、MinIO
- 建议开启版本控制或生命周期策略处理过期对象
- 配置完成后，可在后台执行健康检查验证凭证有效性

## WebDAV

```yaml
storage:
  type: webdav
  webdav:
    url: "https://dav.example.com/remote.php/webdav/filecodebox"
    username: "dav-user"
    password: "dav-pass"
```

- 适合已有 NAS / Nextcloud 环境
- 通过 HTTPS 确保传输安全
- 可结合 Basic Auth + IP 限制增强访问安全

## OneDrive / Microsoft 365

```yaml
storage:
  type: onedrive
  onedrive:
    client_id: "..."
    client_secret: "..."
    refresh_token: "..."
    drive_id: "..."
```

- 需在 Azure AD 中注册应用并获取 Client ID/Secret
- 使用 Graph API 刷新 Token，保持长期可用
- 适合教育/企业订阅环境，方便与 Office 套件协同

## 迁移流程

1. 确认目标存储已配置好凭证与访问权限
2. 后台“存储配置”中切换类型并保存
3. 执行“健康检查”确保写入成功
4. 使用“迁移向导”（规划中）或脚本将历史文件同步到新存储

> 切换存储仅影响后续上传。历史文件仍保留在旧存储中，需手动迁移或保留只读访问。

## 常用脚本

仓库 `scripts/` 目录提供了若干辅助脚本：

- `scripts/export_config_from_db.go`：导出数据库内的配置，便于备份
- `scripts/test_nfs_storage.sh`：快速验证 NFS / WebDAV 连通性

## 故障排查

| 场景 | 排查步骤 |
| --- | --- |
| 上传失败（S3 403） | 检查 Bucket 策略、AccessKey 权限、Region 是否匹配 |
| WebDAV 超时 | 确认网络连通性、提高 `download_timeout`、检查 SSL 证书 |
| OneDrive 刷新失败 | 更新 Refresh Token、确认应用权限包含 `Files.ReadWrite.All` |

下一章节：[系统配置详解](./configuration.md)

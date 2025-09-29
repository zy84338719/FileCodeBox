# API 使用指南

FileCodeBox 公开了覆盖上传、分享、用户、管理的 REST API，便于集成到自动化流程或自定义客户端中。

> 完整的 Swagger 文档位于 `docs/swagger-enhanced.yaml`，部署后可访问 `/swagger/index.html` 在线查看。

## 基础信息

| 项目 | 说明 |
| --- | --- |
| 基础路径 | `https://your-domain.com` |
| 返回格式 | JSON（统一 `code`、`message`、`data` 字段） |
| 编码 | UTF-8 |

## 认证方式

| 类型 | 说明 |
| --- | --- |
| 公共接口 | 上传初始化、获取分享信息等无需认证 |
| 用户登录 | `POST /user/login`，返回用户 Token |
| 管理员登录 | `POST /admin/login`，返回 Bearer Token |
| API Key | （可选）在后台生成，使用 `Authorization: ApiKey <key>` |

Token 默认 72 小时过期，可在配置中调整。

## 通用错误码

| `code` | 说明 |
| --- | --- |
| `200` | 请求成功 |
| `400` | 参数错误 |
| `401` | 未授权或 Token 失效 |
| `403` | 无权限访问 |
| `404` | 资源不存在 |
| `429` | 触发限流 |
| `500` | 服务异常 |

客户端应根据 `code` 和 `message` 进行错误处理。

## 主要模块

- [上传与分片 API](./upload.md)
- [管理类 API](./admin.md)
- 用户系统、分享资源 API（待补充）

## 速查示例

```bash
# 登录获取管理员 Token
curl -X POST https://your-domain.com/admin/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"***"}'

# 使用 Token 获取仪表盘数据
curl https://your-domain.com/admin/dashboard \
  -H 'Authorization: Bearer <token>'
```

> 建议结合 HTTPS、IP 白名单与 API Key 控制访问来源。

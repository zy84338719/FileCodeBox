# Upload / Chunk API

以下接口用于实现大文件上传、秒传与断点续传逻辑。所有接口返回统一格式：

```json
{
  "code": 200,
  "message": "success",
  "data": {...}
}
```

## 1. 初始化上传任务

`POST /chunk/upload/init/`

### 请求体

```json
{
  "filename": "report.zip",
  "filesize": 10485760,
  "hash": "<md5 or sha1>",
  "mime_type": "application/zip"
}
```

### 返回示例

```json
{
  "code": 200,
  "data": {
    "upload_id": "2c5f7c98-...",
    "chunk_size": 2097152,
    "uploaded_chunks": []
  }
}
```

- 若服务端检测到同哈希文件已存在，将返回 `uploaded=true` 并直接生成分享记录（秒传）。

## 2. 上传分片

`POST /chunk/upload/chunk/:upload_id/:index`

- `upload_id`：初始化返回的任务 ID
- `index`：分片序号，从 0 或 1 开始（以返回值为准）

### 请求头

```
Content-Type: application/octet-stream
Content-Range: bytes <start>-<end>/<total>
Authorization: Bearer <token>  # 若需登录权限
```

### 返回

- `204 No Content` 或 `200`，表示分片上传成功
- 若重复上传已完成的分片，服务端会忽略并返回成功

## 3. 查询任务状态

`GET /chunk/upload/status/:upload_id`

```json
{
  "code": 200,
  "data": {
    "upload_id": "2c5f7c98-...",
    "chunk_size": 2097152,
    "uploaded_chunks": [0,1,2]
  }
}
```

可用于前端断点续传或 CLI 监控任务进度。

## 4. 合并分片

`POST /chunk/upload/complete/:upload_id`

### 请求体

```json
{
  "expire_style": "day",
  "expire_value": 7,
  "require_auth": false,
  "filename": "report.zip",
  "metadata": {
    "description": "季度报告"
  }
}
```

### 返回值

```json
{
  "code": 200,
  "data": {
    "share_code": "ABCD",
    "share_url": "https://your-domain.com/s/abcd1234",
    "expire_at": "2025-10-01T13:00:00Z"
  }
}
```

> 合并过程会执行哈希校验，确保所有分片完整且顺序正确。

## 5. 取消上传

`DELETE /chunk/upload/cancel/:upload_id`

- 会清理临时分片文件
- 返回 200 表示取消成功

## 6. 普通上传 API

若无需分片，可使用简化接口：

`POST /upload/file/`

- 表单字段：`file`、`expire_style`、`expire_value`
- 适合小文件或脚本场景

## 错误处理

| 错误码 | 说明 |
| --- | --- |
| `413` | 文件太大，超出 `upload_size` 限制 |
| `409` | 上传任务已存在或状态异常 |
| `422` | 分片缺失或顺序错误 |
| `500` | 分片合并失败，可重试或查看服务日志 |

## 最佳实践

- 分片大小建议保持在 1~5 MiB，避免网络抖动
- 上传前计算 MD5/sha1，利用秒传能力节省带宽
- 对自动化任务，可在完成上传后调用管理 API 更新描述或标签

继续阅读：[管理类 API](./admin.md)

# FileCodeBox 运维指南

面向生产环境运维：监控、日志、健康检查、备份、升级、故障排查。

## 1. 监控

### Prometheus 指标

服务暴露 `/metrics`（默认开启，可用 `FCB_METRICS_ENABLED=false` 关闭）：

| 指标 | 类型 | 说明 |
| --- | --- | --- |
| `http_requests_total{method,path,status}` | Counter | 请求总数 |
| `http_request_duration_seconds{method,path}` | Histogram | 请求延迟分布 |
| `http_requests_in_flight{method}` | Gauge | 当前处理中请求数 |
| `go_*` / `process_*` | — | Go runtime / 进程指标 |

K8s 集群若已安装 Prometheus Operator，`deploy/k8s/base/servicemonitor.yaml` 会自动被识别。

**Grafana 常用 PromQL 示例**：
- QPS：`rate(http_requests_total[1m])`
- P99 延迟：`histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))`
- 错误率：`rate(http_requests_total{status=~"5.."}[1m]) / rate(http_requests_total[1m])`

### 健康检查端点

| 端点 | 用途 | K8s 探针 |
| --- | --- | --- |
| `GET /live` | 进程存活（轻量） | livenessProbe |
| `GET /readyz` | 就绪（含 DB ping，依赖不可用返回 503） | readinessProbe |
| `GET /health` | 通用健康检查 | — |

## 2. 日志

日志为 zap 结构化 JSON（输出到 stdout），每条访问日志包含：
`method / path / status / ip / latency / size / trace_id`

**trace_id 全链路**：每个请求自动生成/透传 `X-Trace-Id`（客户端可传入串联上游），响应头与响应体（`trace_id` 字段）均会返回，可在日志中按 trace_id 串联一次请求的全部日志。

排查示例（按 trace_id 过滤）：
```bash
docker logs filecodebox-app 2>&1 | grep '"trace_id":"abc-123"'
```

## 3. 备份与恢复

### SQLite（默认）
```bash
# 备份（使用 sqlite3 在线备份，避免锁冲突）
sqlite3 /app/data/fileCodeBox.db ".backup /backup/fileCodeBox-$(date +%F).db"

# 或直接复制文件（需短暂停止写入）
cp /app/data/fileCodeBox.db /backup/
```

### 上传文件
数据目录 `/app/data/uploads` 需定期备份（rsync / 快照）。

### K8s
建议对 PVC `filecodebox-data` 做定期快照（VolumeSnapshot）或 CronJob 备份。

## 4. 升级

1. 备份数据（见上）
2. 拉取新镜像：`docker pull ghcr.io/zy84338719/filecodebox:<new-tag>`
3. 更新 K8s Deployment 镜像 tag，触发滚动更新
4. 观察 `/readyz` 与 `/metrics` 确认实例就绪

> 注意：跨大版本升级前请阅读 CHANGELOG，确认是否有破坏性变更。

## 5. 故障排查

| 现象 | 排查方向 |
| --- | --- |
| 启动失败 "requires a secure jwt_secret" | 生产模式检测到默认密钥；设置 `FCB_JWT_SECRET` 环境变量 |
| `/readyz` 返回 503 | DB 不可达；检查 `database` 配置 / 网络连通性 |
| 前端打不开 | 确认 `static/` 目录存在且包含 index.html（单二进制需先 `make copy-frontend`） |
| CORS 报错 | 配置 `FCB_CORS_ALLOW_ORIGINS` 白名单（生产不要留空） |
| 上传 413 | 调大 `FCB_UPLOAD_SIZE`；若经 nginx/ingress 还需调 `proxy-body-size` |

## 6. 默认管理员

首次启动自动创建管理员 `admin / admin123`（bcrypt 哈希）。**生产环境请立即登录后台修改密码**，并考虑通过 `FCB_PRODUCTION=1` 启用生产模式强制 secret 校验。

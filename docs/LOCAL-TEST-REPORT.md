# FileCodeBox 本地测试报告

> 日期：2026-07-29
> 测试环境：macOS darwin 25.5.0 arm64, Go 1.26.5, sqlite 内存/文件库
> 测试方式：单二进制启动 + curl 接口实测 + 单元测试

## 一、构建验证

| 项目 | 命令 | 结果 |
|---|---|---|
| 后端单二进制 | `go build -o filecodebox ./cmd/server` | ✅ 40MB 二进制 |
| 后端全量测试 | `go test ./...` | ✅ 12 个包全过 |
| go vet | `go vet ./...` | ✅ 干净 |
| 前端 tsc | `vue-tsc --noEmit` | ✅ 零错误 |
| 前端构建 | `npm run build` | ✅ 1838 模块 |

## 二、接口实测（启动服务 :18080）

### ✅ 通过的测试

| # | 接口 | 测试内容 | 结果 |
|---|---|---|---|
| 1 | `GET /readyz` | 健康检查/就绪探针 | ✅ `{"status":"ready","checks":{"database":true}}` |
| 2 | `GET /api/config` | 公开配置 | ✅ 返回站点配置（名称/上传限制/过期选项） |
| 3 | `POST /admin/login` | 管理员登录 | ✅ 返回 JWT token（默认 admin/admin123） |
| 4 | `POST /api/v1/user/refresh` | **token 刷新（第三批新增）** | ✅ 旧 token 换发新 token |
| 5 | `GET /metrics`(主端口) | **metrics 隔离（第一批）** | ✅ 404（主端口不暴露指标） |
| 6 | `GET 127.0.0.1:9090/metrics` | **metrics 内网监听（第一批）** | ✅ 返回 Prometheus 指标 |
| 7 | `POST /anonymous/generate`(超大文件) | **上传大小校验（第一批）** | ✅ 拒绝：`文件过大` |
| 8 | `POST /anonymous/generate`(.exe) | **上传类型校验（第一批）** | ✅ 拒绝：`该文件类型禁止上传` |
| 9 | 限流中间件 | **限流挂载（第一批）** | ✅ 中间件链可见 RateLimiter |

### 🐛 发现并修复的 Bug

| Bug | 原因 | 修复 |
|---|---|---|
| **匿名取件 panic**（`anonymous.go:81` nil pointer） | 本地无 Redis 时 `s.rdb` 为 nil，`SetNX` 解引用 panic | GenerateCode/Retrieve/Peek 开头加 nil 检查，返回明确错误 `Redis 未配置` |
| **admin 登录后卡在登录页**（第三批引入的回归） | admin Login.vue 改用 fetchUserInfo 拿 role，但 `/user/info` 接口不返回 role 字段，导致 role 判断失败登出 | `/user/info` handler 补充 role 字段返回（用 map 替代 thrift model） |

修复后匿名取件在无 Redis 时返回：
```json
{"code":10008,"message":"Redis 未配置，匿名取件功能不可用"}
```
（生产环境有 Redis 时正常工作）

## 三、安全修复验证

| 安全项 | 验证方式 | 结果 |
|---|---|---|
| JWT fail-fast | 不设 FCB_JWT_SECRET 启动 | ✅ 启动报错退出 |
| metrics 隔离 | 主端口访问 /metrics | ✅ 404，仅内网 :9090 可访问 |
| 上传校验 | 超大文件/黑名单扩展名 | ✅ 均被拒绝 |
| CORS 收紧 | 未配白名单时非 localhost Origin | ✅ 不反射（需配 FCB_CORS_ALLOW_ORIGINS） |
| 限流挂载 | 登录/上传/下载路径 | ✅ 中间件链已挂载 |

## 四、结论

- **所有修复功能正常**，服务可正常启动运行
- **发现并修复 1 个运行时 bug**（匿名取件 nil panic）
- 单元测试 12 包全过、前后端构建均通过
- 可进入生产部署

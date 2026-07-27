# idl/ - 接口定义语言

此目录存放 IDL（Interface Definition Language）文件。

## 目录结构

```
idl/
├── api.thrift            # hz 注解约定文档（thrift 版本）
├── common.proto          # 旧 proto IDL（迁移中，逐步替换为 .thrift）
├── common.thrift         # 新 thrift IDL（推荐）
├── setup.proto|thrift
├── qrcode.proto|thrift
├── chunk.proto|thrift
├── share.proto|thrift
├── user.proto|thrift
├── admin.proto|thrift
├── storage.proto|thrift
├── maintenance.proto|thrift
├── health.proto|thrift
├── http/                 # HTTP API 定义子目录
│   ├── health.proto|thrift
│   └── ...
└── rpc/                  # Kitex RPC 服务定义
    └── ...
```

## IDL 迁移进度（proto → thrift）

| 状态 | 含义 |
|---|---|
| ✅ thrift | 已迁移到 thrift，hz 用 thrift 路径生成 |
| 🔄 proto | proto 旧文件保留（hz 仍可生成） |
| 🆕 new | thrift 新增能力（无 proto 对应） |

| IDL | 状态 | 说明 |
|---|---|---|
| `common.thrift` | 🔄 proto | health/index/ping |
| `setup.thrift` | 🔄 proto | 系统初始化 |
| `qrcode.thrift` | 🔄 proto | 二维码 |
| `chunk.thrift` | 🔄 proto | 分片上传 |
| `share.thrift` | 🔄 proto | 分享 |
| `share_anonymous.thrift` | 🆕 new | 匿名取件（vastsa UX） |
| `user.thrift` | 🔄 proto | 用户 |
| `admin.thrift` | 🔄 proto | 管理员 |
| `storage.thrift` | 🔄 proto | 存储管理（OpenDAL） |
| `maintenance.thrift` | 🔄 proto | 维护 |
| `presign.thrift` | 🆕 new | 预签名上传 |
| `ratelimit.thrift` | 🆕 new | 限流管理 |
| `notify.thrift` | 🆕 new | 系统通知 |
| `error.thrift` | 🆕 new | 统一错误响应 schema |

## 代码生成

### thrift IDL（推荐）

```bash
# 初始化新项目（首次使用）
make gen-thrift-new IDL=common.thrift

# 更新单个 thrift IDL
make gen-thrift-update IDL=common.thrift
make gen-thrift-update IDL=http/health.thrift

# 批量更新所有 thrift IDL
make gen-thrift-update-all

# 或直接用脚本
./scripts/gen.sh thrift-new idl/common.thrift
./scripts/gen.sh thrift-update idl/http/health.thrift
./scripts/gen.sh thrift-update-all
```

### proto IDL（兼容路径，迁移期使用）

```bash
# 初始化
make gen-http-new IDL=common.proto

# 更新
make gen-http-update IDL=common.proto
make gen-http-update IDL=http/health.proto

# 批量更新
make gen-http-update-all
```

## thrift IDL 编写规范

参见 `idl/api.thrift` 文档，关键规则：

1. `namespace go filecodebox.<sub>`（Go 包名 = namespace 最后一段）
2. type 命名 `XxxReq` / `XxxResp`（与 proto 1:1）
3. service 命名 `XxxService`（与 proto 1:1）
4. field 编号从 1 开始，与原 proto 编号保持一致（API 兼容）
5. 路由注解内联在 method 末尾：`Method(1: Req req) (api.get = "/path")`
6. 路径参数用 `:param` 表示（hz 自动转 `{param}`）

## 安装工具

```bash
./scripts/install-tools.sh           # 装 hz + thriftgo
./scripts/install-tools.sh hz        # 只装 hz
./scripts/install-tools.sh thriftgo  # 只装 thriftgo
```

## 注意事项

- HTTP IDL 生成到 `gen/http/`，proto + thrift 共用同一目录
- 迁移期间 proto + thrift 可并存，但建议逐步切换到 thrift
- 全部迁移完成后会删除所有 .proto 文件
- 路由注解（`api.get` 等）由 hz 工具直接读取 thrift method 末尾的注解
- `api.thrift` 是约定文档，不是服务定义

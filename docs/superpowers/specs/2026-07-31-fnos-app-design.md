# FileCodeBox 飞牛(fnOS)应用化设计

> 状态:已确认(选项 C) · 日期:2026-07-31
> 范围:将 FileCodeBox 包装为可在飞牛 NAS 应用商店上架的第三方应用,并接入飞牛 Open API。

---

## 1. 目标与非目标

### 目标
1. 新建独立项目 `filecodebox-fnos`,把 FileCodeBox 当作库复用,不在原项目写飞牛业务逻辑。
2. 产出符合飞牛第三方应用规范(`source = thirdparty`)的 `.fpk` 打包定义。
3. 接入飞牛 Open API 四项能力:SSO 免登录、共享目录列表读取、通知中心推送、内网穿透/外网分享。
4. 凭证(appid/appsecret)留作可配置项,先抽象接口、后填凭证。

### 非目标
- 不改动 FileCodeBox 的任何业务逻辑(原项目仅做"包位置提升"的机械改造)。
- 不做多架构分发:本期 `platform = x86`,arm 版本后续按需新增(飞牛以独立 fpk 包区分架构)。
- 不在本期实现飞牛应用商店的自动发版流水线(先产出可手动打包的产物)。

---

## 2. 总体架构

### 2.1 两项目关系

```
FileCodeBox(原项目,机械改造)          filecodebox-fnos(新项目,飞牛应用主体)
├─ backend/api/...(原 internal,提升)   ├─ fnos/                      ← .fpk 包定义
│   业务逻辑不变,仅目录改名             │   ├─ manifest
└─ cmd/server/bootstrap(对外库入口)    │   ├─ config/{privilege,resource}
                                        │   ├─ docker/docker-compose.yaml
                                        │   ├─ ui/ cmd/ *.sc ICON*.PNG
                                        │   └─ wizard/
                                        ├─ adapter/(Go)飞牛 Open API 适配层
                                        │   ├─ sso/        免登录
                                        │   ├─ storage/    共享目录列表
                                        │   ├─ notify/     通知中心
                                        │   └─ tunnel/     内网穿透
                                        └─ go.mod(replace → 本地 FileCodeBox)
```

### 2.2 运行时形态:单容器单进程

飞牛 `source = thirdparty` 应用在 docker-compose 中使用 `${TRIM_PKGVAR}` 指向包变量目录,且以**单镜像**为部署单元。因此运行时为单容器:

- 容器入口 = 新项目的 `fnos-adapter` 二进制。
- `fnos-adapter` 启动时以**库调用**方式 `bootstrap.Bootstrap(configPath)` 拉起 FileCodeBox 全部业务,拿到 `*server.Hertz` 句柄。
- adapter 在同一进程内,以**反向代理 + HTTP 中间件**包裹 Hertz server,注入飞牛能力:
  - SSO:拦截 `/api/auth/fnos-login`,用飞牛 ticket 换取飞牛用户信息,映射为本系统用户并签发本系统 JWT。
  - 目录:暴露 `/api/fnos/shares`,代理飞牛"列出共享目录"接口,供前端配置页选择存储目录。
  - 通知:封装飞牛通知推送 client,供 FileCodeBox 通知模块在配置为"fnos"通道时调用。
  - 穿透:在分享生成时,可选调用飞牛内网穿透接口生成本分享的外网链接。

这样既满足"当库用"(同进程 import),又满足飞牛"单容器"约束,避免 sidecar 的跨进程通信复杂度。

---

## 3. 原项目改造(选项 C,机械操作)

### 3.1 改造事实依据(已核实)
- 模块路径:`github.com/zy84338719/fileCodeBox/backend`
- `bootstrap.go` 是 internal 包的"集散地",import **15 个** internal 包。
- 跨模块 import bootstrap 时,编译器沿依赖图检查,15 个 internal 包全部触发 `use of internal package ... not allowed`。
- 故"最小子集提升"不可行,必须整体提升 internal。
- 引用规模:**59 个 .go 文件 / 141 处 import 行**(其中 `gen/` 占 17 文件),全部为可机械替换的文本。

### 3.2 改造步骤
1. `git mv backend/internal backend/api`
2. 全局替换 import 路径:`fileCodeBox/backend/internal/` → `fileCodeBox/backend/api/`
   - 范围:`backend/**/*.go`(含 `gen/`、`cmd/`)
3. 验证:`cd backend && go build ./... && go vet ./... && go test ./...`
4. `cmd/server/bootstrap` 路径不变,但因 internal 已提升,它现在可被跨模块 import,成为对外库入口。

### 3.3 改造的安全性
- 零业务逻辑变更:仅目录改名 + import 路径替换。
- 原项目自身行为完全等价(`go build` 通过即证明)。
- 可回滚:`git revert` 一次提交。
- `bootstrap.Bootstrap(configPath) (*server.Hertz, error)` 签名不变,是稳定的库 API。

---

## 4. 新项目 filecodebox-fnos 结构

```
filecodebox-fnos/
├─ go.mod                       module github.com/zy84338719/filecodebox-fnos
│                               require github.com/zy84338719/fileCodeBox/backend v0.0.0
│                               replace github.com/zy84338719/fileCodeBox/backend => ../FileCodeBox/backend
├─ cmd/
│   └─ fnos-adapter/
│       └─ main.go              入口:Bootstrap() 拉起业务 + 挂载 adapter 中间件 + Spin
├─ adapter/                     飞牛 Open API 适配层
│   ├─ adapter.go               反向代理 + 中间件装配,包裹 *server.Hertz
│   ├─ client.go                飞牛 Open API HTTP client(签名/鉴权,凭证从环境变量读)
│   ├─ config.go                飞牛适配配置(FNOS_APPID/FNOS_APPSECRET/FNOS_API_BASE 等)
│   ├─ sso/                     SSO 免登录
│   │   └─ sso.go               /api/fnos/login:ticket→飞牛用户→本系统用户→本系统JWT
│   ├─ storage/                 共享目录列表
│   │   └─ storage.go           /api/fnos/shares:代理飞牛"列出共享目录"
│   ├─ notify/                  通知中心
│   │   └─ notify.go            实现 FileCodeBox 通知接口的 fnos channel
│   └─ tunnel/                  内网穿透
│       └─ tunnel.go            分享外网链接生成(可选,凭证就绪后启用)
├─ fnos/                        飞牛 .fpk 打包定义
│   ├─ manifest                 应用元数据(键值对格式)
│   ├─ config/
│   │   ├─ privilege            运行权限声明(run-as/username/groupname)
│   │   └─ resource             资源声明
│   ├─ docker/
│   │   └─ docker-compose.yaml  飞牛容器编排(飞牛变量占位)
│   ├─ ui/                      桌面入口 UI
│   ├─ cmd/                     生命周期脚本(install/uninstall/upgrade)
│   ├─ wizard/                  安装向导
│   ├─ FileCodeBox.sc           端口转发声明
│   ├─ ICON.PNG                 应用图标(48/256)
│   └─ ICON_256.PNG
├─ Dockerfile                   单容器构建:编译 fnos-adapter + 内嵌 FileCodeBox
└─ README.md
```

---

## 5. 飞牛 Open API 适配层设计

### 5.1 凭证与配置(先抽象,后填)
凭证从环境变量读取,缺省时 adapter 以**降级模式**运行(飞牛能力关闭,FileCodeBox 业务正常):

| 环境变量 | 说明 | 缺省行为 |
|---------|------|---------|
| `FNOS_APPID` | 飞牛应用 appid | 关闭所有飞牛能力 |
| `FNOS_APPSECRET` | 飞牛应用 secret | 同上 |
| `FNOS_API_BASE` | 飞牛 Open API 基址 | 默认飞牛官方地址 |
| `FNOS_ENABLED` | 总开关 `true/false` | false 时纯 passthrough |

### 5.2 client.go:飞牛 API client
- 统一封装签名(按飞牛文档的签名算法,凭证就绪后补全)、请求重试、错误码映射。
- 所有飞牛调用经此 client,便于凭证注入与降级判断。

### 5.3 四项能力(按实现优先级)

**① SSO 免登录(最高优先)**
- 路由:`POST /api/fnos/login { ticket }`
- 流程:校验 ticket → 调飞牛"换取用户信息" → 查/建本系统用户(用户名=飞牛用户名) → 签发本系统 JWT → 返回。
- 前端:登录页增加"飞牛账号登录"按钮,跳飞牛授权拿 ticket 回跳。

**② 共享目录列表**
- 路由:`GET /api/fnos/shares`
- 流程:调飞牛"列出共享目录" → 返回目录名/路径/容量。
- 前端:存储配置页用此填充"存储目录"下拉,替代手填。

**③ 通知中心**
- 实现 FileCodeBox 通知模块的一个 channel(命名 `fnos`),在通知配置选 `fnos` 时,经 client 推送到飞牛通知中心。
- 不侵入 FileCodeBox 通知抽象(若现有抽象不支持外部 channel,adapter 以独立 goroutine 订阅 DB 通知表的方式实现,零侵入)。

**④ 内网穿透/外网分享(可选,凭证就绪后)**
- 分享生成时,可选调飞牛内网穿透,为本分享生成本机端口的外网映射链接。
- 默认关闭,配置开启后生效。

### 5.4 adapter.go:中间件装配
- 反向代理方式包裹 Hertz:adapter 监听对外端口,把 `/api/fnos/*` 路由到 adapter handler,其余全部透传给 Hertz server。
- 不修改 Hertz 内部路由,保证 FileCodeBox 原有 API 行为不变。

---

## 6. 飞牛 .fpk 打包定义

### 6.1 manifest(键值对格式,参照真实仓库)
```
appname         = filecodebox
version         = 2.0.0
display_name    = FileCodeBox
platform        = x86
maintainer      = zhangyi
maintainer_url  = https://github.com/zy84338719/FileCodeBox
distributor     = zhangyi
distributor_url = https://github.com/zy84338719/FileCodeBox
desktop_uidir   = ui
desktop_applaunchname = filecodebox.Application
service_port    = 12345
desc            = 高性能文件/文本分享平台,支持拖拽上传、匿名取件、多存储后端。
source          = thirdparty
checksum        =
```

### 6.2 docker/docker-compose.yaml(飞牛变量占位,参照真实仓库)
```yaml
services:
  filecodebox:
    image: ${DOCKER_MIRROR}docker.io/zy84338719/filecodebox-fnos:${VERSION}
    container_name: filecodebox
    ports:
      - "${TRIM_SERVICE_PORT}:12345"
    volumes:
      - ${TRIM_PKGVAR:?}/data:/app/data       # NAS 共享文件夹挂载
    environment:
      - TZ=Asia/Shanghai
      - FCB_SERVER_PORT=12345
      - FCB_DATA_PATH=/app/data
      - FCB_DATABASE_DRIVER=sqlite
      - FNOS_APPID=${FNOS_APPID:-}
      - FNOS_APPSECRET=${FNOS_APPSECRET:-}
    restart: unless-stopped
```
- `${TRIM_PKGVAR}` = 飞牛包变量目录(用户可见的 NAS 共享路径),映射到容器 `/app/data`,实现"挂载 NAS 共享文件夹"。
- SQLite 库与上传文件均落于此,用户可在飞牛文件管理中直接看到、备份、迁移。
- `service_port` = 12345,与 FileCodeBox 默认端口一致。

### 6.3 config/privilege(权限声明)
```json
{
  "defaults": { "run-as": "package" },
  "username": "filecodebox",
  "groupname": "filecodebox"
}
```

### 6.4 图标
- `ICON.PNG`(48×48)、`ICON_256.PNG`(256×256),沿用 FileCodeBox 现有 favicon/品牌图,按需缩放。

---

## 7. 错误处理与降级

- **凭证缺失**:adapter 启动检测 `FNOS_APPID/FNOS_APPSECRET`,缺失则日志告警并进入降级模式(飞牛路由返回 503 提示"未配置飞牛凭证",其余 API 正常)。
- **飞牛 API 不可达**:client 设超时(3s)+ 重试(2 次),失败返回明确错误码,不阻塞 FileCodeBox 主流程。
- **SSO 失败**:回退到 FileCodeBox 原生登录页(飞牛登录按钮点击报错提示)。

---

## 8. 验证策略

1. **原项目改造后**:`go build ./... && go vet ./... && go test ./...` 全绿,证明行为等价。
2. **新项目编译**:`go build ./cmd/fnos-adapter` 通过,证明库式调用可行。
3. **本地运行**:无凭证降级模式启动,FileCodeBox 全功能可用(上传/取件/管理后台)。
4. **飞牛能力**:凭证就绪后,逐项验证 SSO/目录/通知/穿透(本期产出接口与桩,凭证后补真实验证)。
5. **打包**:手动产出 `.fpk`,结构对照真实仓库(dpanel/alist)一致。

---

## 9. 实施顺序

1. 原项目改造:`internal`→`api`(机械操作 + 验证)。
2. 新建 `filecodebox-fnos`:go.mod + replace + cmd/fnos-adapter 骨架,验证库式拉起。
3. adapter 骨架:反向代理 + 降级模式 + config/client。
4. 四项飞牛能力(接口抽象 + 桩实现)。
5. fnos 打包定义(manifest/compose/config/图标)。
6. 本地联调 + 文档。

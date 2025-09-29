# 快速开始

本章节帮助你在 10 分钟内完成 FileCodeBox 的部署与初始化。

## 环境要求

| 组件 | 说明 |
| --- | --- |
| 操作系统 | Linux / macOS / Windows 均可（推荐 Linux / 容器化部署） |
| Go | 1.21 及以上（本地开发推荐 1.24+） |
| 数据库 | SQLite（默认）或 MySQL / PostgreSQL |
| 可选 | Docker 20+、docker-compose v2、Make、Nginx/Traefik |

## 获取源码

```bash
# 克隆仓库
git clone https://github.com/zy84338719/FileCodeBox.git
cd FileCodeBox

# 拉取依赖
go mod download
```

> 使用国内镜像时可预先设置 `GOPROXY=https://goproxy.cn,direct`。

## 方式一：本地开发（Go 运行）

```bash
# 运行服务
go run ./...

# 或先构建再启动
make build
./build/filecodebox
```

- 默认监听地址：`http://127.0.0.1:12345`
- 首次访问会自动跳转 `/setup` 引导创建数据库和管理员

## 方式二：Docker 部署（推荐）

```bash
docker run -d \
  --name filecodebox \
  -p 12345:12345 \
  -v $(pwd)/data:/data \
  ghcr.io/zy84338719/filecodebox:latest
```

- 数据目录会持久化到本地 `./data`
- 可通过 `-e CONFIG_PATH=/data/config.yaml` 指定配置文件

### docker-compose

```bash
cp docker-compose.yml docker-compose.override.yml  # 如需自定义
# 按需修改 override 中的存储、数据库、环境变量

docker compose up -d
```

## 方式三：生产环境二进制

1. 从 Release 页面下载对应平台的压缩包
2. 解压后放置到 `/opt/filecodebox` 等目录
3. 创建 `config.yaml`（可复制 `docs/config.new.yaml`）
4. 配置 systemd：

```ini
[Unit]
Description=FileCodeBox Service
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/filecodebox
ExecStart=/opt/filecodebox/filecodebox --config=/opt/filecodebox/config.yaml
Restart=always
User=filecodebox

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now filecodebox
```

## 初始化流程

1. 浏览器访问 `http://服务器IP:12345`
2. 根据向导完成以下步骤：
   - 选择数据库类型 & 填写 DSN（默认 SQLite 会自动生成）
   - 创建管理员账户与密码
   - 设置站点名称、下载策略等基础配置
3. 完成后将进入 Dashboard，可开始上传文件

> 初始化完成后 `/setup` 将自动关闭，并拒绝再次访问。

## 目录结构

```
FileCodeBox/
├── main.go              # 入口
├── internal/            # 分层业务逻辑
├── themes/2025/         # 前端静态资源（用户中心 + 管理后台）
├── data/                # SQLite / 上传文件默认位置
└── docs/                # 参考文档与专题文章
```

## 下一步

- [文件上传体验](./upload.md) — 了解拖拽、分片与秒传
- [分享与领取](./share.md) — 控制有效期、提取码与下载次数
- [存储适配指南](./storage.md) — 切换到云存储或第三方 NAS
- [管理后台总览](./management.md) — 掌握后台操作面板

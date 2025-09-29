# FileCodeBox 文档中心

> 基于 Go 的 FileCodeBox 让“像拿快递一样取文件”成为现实。文档站点涵盖部署、功能配置、安全与 API 参考。

## 快速入口

- [功能导览](./guide/introduction.md)
- [快速上手部署](./guide/getting-started.md)
- [上传分片指南](./guide/upload.md)
- [安全加固手册](./guide/security.md)
- [API 使用指南](./api/README.md)

## 使用 VitePress 本地预览

此目录使用 [VitePress](https://vitepress.dev/) 构建。

### 安装依赖（首次执行）

在仓库根目录执行一次安装（已配置 npm workspace）：

```bash
npm install
```

### 启动开发服务器

```bash
npm run docs:dev
```

### 构建静态站点

```bash
npm run docs:build
```

构建结果默认输出到 `docs/site/.vitepress/dist`，可配合静态站点托管服务发布。

## 目录结构

```text
site/
├── index.md                 # 首页英雄区与导航索引
├── README.md                # 当前文件，开发指导
├── guide/                   # 部署、配置、运维等主题文档
│   ├── introduction.md
│   ├── getting-started.md
│   ├── upload.md
│   ├── share.md
│   ├── management.md
│   ├── storage.md
│   ├── configuration.md
│   ├── security.md
│   └── troubleshooting.md
└── api/                     # REST API 参考
	├── README.md
	├── upload.md
	└── admin.md
```

其他静态资源可直接放在 `docs/site/public/` 或通过 VitePress 自定义主题引入。

## 贡献建议

1. 新增或修改文档时请保持章节编号与导航一致。
2. 若调整页面结构，记得同步更新 `index.md` 与（待添加的） `.vitepress/config.ts` 导航。
3. 文档示例中的域名与凭证请使用占位符，避免泄露生产信息。

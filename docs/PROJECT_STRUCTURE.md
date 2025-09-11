# FileCodeBox 项目目录结构

## 📁 整体目录结构

```
FileCodeBox/
├── assets/                     # 项目资源文件
│   └── images/
│       └── logos/             # Logo 相关图片
│           ├── logo.svg           # 主 Logo (64x64px, 带动画)
│           ├── logo-horizontal.svg # 横版 Logo (280x80px)
│           ├── logo-small.svg     # 小尺寸 Logo (32x32px)
│           ├── logo-monochrome.svg # 黑白版 Logo
│           └── favicon.svg        # 网站图标 (16x16px)
│
├── docs/                       # 文档目录
│   ├── changelogs/            # 改动记录
│   │   ├── LOGO_UPDATE_REPORT.md  # Logo 更新报告
│   │   └── REFACTOR_SUMMARY.md    # 重构总结
│   ├── design/                # 设计相关文档
│   │   ├── LOGO_DESIGN.md         # Logo 设计文档
│   │   └── logo-showcase.html     # Logo 展示页面
│   ├── API_SWAGGER_GUIDE.md   # API 接口文档
│   ├── API-README.md          # API 使用说明
│   ├── database-config.md     # 数据库配置
│   └── ... (其他已有文档)
│
├── scripts/                    # 脚本文件
│   ├── build-docker.sh           # Docker 构建脚本
│   ├── deploy.sh                  # 部署脚本
│   ├── generate_favicon.sh       # Favicon 生成脚本
│   ├── quick-push.sh             # 快速推送脚本
│   ├── release.sh                # 发布脚本
│   ├── tag-manager.sh            # 标签管理脚本
│   ├── test_mcp_client.py        # MCP 客户端测试
│   └── test_nfs_storage.sh       # NFS 存储测试
│
├── themes/                     # 主题文件
│   └── 2024/
│       ├── assets/
│       │   └── images/            # 主题专用图片
│       │       ├── logo.svg           # 主题 Logo (带动画)
│       │       ├── logo-small.svg     # 主题小尺寸 Logo
│       │       ├── logo-lock.svg      # 安全主题 Logo
│       │       └── favicon.svg        # 主题 Favicon
│       ├── index.html             # 主页面
│       ├── admin.html             # 管理后台
│       ├── login.html             # 登录页面
│       ├── register.html          # 注册页面
│       ├── forgot-password.html   # 忘记密码页面
│       ├── dashboard.html         # 用户仪表板
│       └── README.md              # 主题说明
│
├── internal/                   # Go 源代码
│   ├── config/                    # 配置管理
│   ├── database/                  # 数据库
│   ├── handlers/                  # HTTP 处理器
│   ├── middleware/                # 中间件
│   ├── models/                    # 数据模型
│   ├── routes/                    # 路由
│   ├── services/                  # 业务逻辑
│   └── storage/                   # 存储接口
│
├── tests/                      # 测试文件
├── deploy/                     # 部署配置
├── data/                       # 数据目录
├── nginx/                      # Nginx 配置
│
├── main.go                     # 主程序入口
├── go.mod                      # Go 模块定义
├── go.sum                      # Go 依赖校验
├── Dockerfile                  # Docker 构建文件
├── docker-compose.yml          # Docker 编排
├── Makefile                    # 构建脚本
├── README.md                   # 项目说明
└── LICENSE                     # 许可证
```

## 📋 目录分类说明

### 1. 资源文件 (`assets/`)
- **用途**: 存放项目级别的静态资源
- **内容**: Logo、图标、图片等
- **特点**: 全项目共享，版本控制管理

### 2. 文档目录 (`docs/`)
- **changelogs/**: 项目改动记录、版本更新日志
- **design/**: 设计文档、UI/UX 相关资料
- **其他**: API 文档、配置说明、用户指南

### 3. 脚本目录 (`scripts/`)
- **构建脚本**: Docker 构建、发布打包
- **部署脚本**: 自动部署、环境配置
- **测试脚本**: 功能测试、性能测试
- **工具脚本**: 开发辅助工具

### 4. 主题目录 (`themes/`)
- **版本化管理**: 按年份或版本号组织
- **独立资源**: 每个主题有自己的 assets 目录
- **模板文件**: HTML 模板和样式文件

## 🔧 路径引用规范

### 项目级别资源
```html
<!-- 项目 Logo -->
<img src="assets/images/logos/logo.svg" alt="FileCodeBox Logo"/>

<!-- 项目文档链接 -->
<a href="docs/API-README.md">API 文档</a>
```

### 主题级别资源
```html
<!-- 主题 Logo -->
<img src="assets/images/logo.svg" alt="Theme Logo"/>

<!-- 主题内相对路径 -->
<link rel="icon" href="assets/images/favicon.svg">
```

### 文档内部引用
```markdown
<!-- 引用项目 Logo -->
![Logo](../../assets/images/logos/logo.svg)

<!-- 引用其他文档 -->
[设计文档](../design/LOGO_DESIGN.md)
```

## 📝 维护建议

### 1. 新增资源
- 图片文件放入 `assets/images/` 对应子目录
- 文档放入 `docs/` 相应分类目录
- 脚本放入 `scripts/` 目录并添加执行权限

### 2. 版本管理
- 重要改动记录放入 `docs/changelogs/`
- 设计变更记录放入 `docs/design/`
- 保持文档与代码同步更新

### 3. 清理原则
- 定期清理过时文档和资源
- 保持目录结构简洁明了
- 避免重复文件和无用资源

---

**整理时间**: 2025年9月11日  
**整理人员**: GitHub Copilot  
**版本**: v1.0

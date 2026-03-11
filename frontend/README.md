# FileCodeBox 前端项目

基于 Vue 3 + Vite + Element Plus 的现代化文件分享平台前端。

## 技术栈

- **框架**: Vue 3 (Composition API + script setup)
- **构建工具**: Vite 5
- **UI 组件库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router 4
- **数据请求**: Axios + TanStack Query
- **工具函数**: VueUse
- **类型检查**: TypeScript
- **样式**: SCSS

## 项目结构

```
src/
├── api/                # API 接口封装
│   ├── share.ts        # 分享相关 API
│   ├── user.ts         # 用户相关 API
│   └── admin.ts        # 管理后台 API
├── components/         # 组件
│   ├── upload/         # 上传相关组件
│   │   ├── FileUpload.vue
│   │   ├── TextShare.vue
│   │   └── GetShare.vue
│   └── common/         # 通用组件
├── composables/        # 组合式函数
├── router/             # 路由配置
├── stores/             # Pinia 状态管理
│   └── user.ts         # 用户状态
├── styles/             # 全局样式
│   └── main.scss
├── types/              # TypeScript 类型定义
│   ├── common.ts
│   └── user.ts
├── utils/              # 工具函数
│   └── request.ts      # Axios 封装
├── views/              # 页面组件
│   ├── home/           # 首页
│   ├── share/          # 分享详情页
│   ├── user/           # 用户相关页面
│   │   ├── Login.vue
│   │   ├── Register.vue
│   │   └── Dashboard.vue
│   └── admin/          # 管理后台
│       ├── index.vue
│       ├── Dashboard.vue
│       ├── Files.vue
│       ├── Users.vue
│       └── Config.vue
├── App.vue
└── main.ts
```

## 开发指南

### 安装依赖

```bash
npm install
```

### 启动开发服务器

```bash
npm run dev
```

访问 http://localhost:3000

### 构建生产版本

```bash
npm run build
```

### 类型检查

```bash
npm run type-check
```

## API 代理配置

开发环境下，API 请求会被代理到后端服务器：

- `/api/*` → `http://localhost:8888`
- `/share/*` → `http://localhost:8888`
- `/user/*` → `http://localhost:8888`
- `/admin/*` → `http://localhost:8888`
- `/chunk/*` → `http://localhost:8888`

## 主要功能

### 1. 文件分享
- 拖拽上传
- 进度显示
- 过期时间设置
- 密码保护

### 2. 文本分享
- 大文本支持
- 字数限制
- 格式保留

### 3. 获取分享
- 分享码输入
- 密码验证
- 文件下载
- 文本复制

### 4. 用户系统
- 注册/登录
- 用户中心
- 配额管理
- 上传统计

### 5. 管理后台
- 仪表盘统计
- 文件管理
- 用户管理
- 系统配置

## 代码规范

- 使用 Composition API + `<script setup>`
- 使用 TypeScript 类型检查
- 组件命名：PascalCase
- 文件命名：kebab-case
- 样式使用 SCSS + scoped

## 下一步开发

### 待完成功能

1. **用户页面**
   - [ ] Login.vue - 登录页面
   - [ ] Register.vue - 注册页面
   - [ ] Dashboard.vue - 用户中心

2. **管理后台页面**
   - [ ] admin/index.vue - 后台布局
   - [ ] admin/Dashboard.vue - 仪表盘
   - [ ] admin/Files.vue - 文件管理
   - [ ] admin/Users.vue - 用户管理
   - [ ] admin/Config.vue - 系统配置

3. **分片上传**
   - [ ] ChunkUpload 组件
   - [ ] 断点续传
   - [ ] 秒传功能

4. **性能优化**
   - [ ] 路由懒加载
   - [ ] 组件按需加载
   - [ ] 图片懒加载

5. **功能增强**
   - [ ] 深色模式
   - [ ] 国际化
   - [ ] PWA 支持
   - [ ] WebSocket 实时通知

## 部署

### Docker 构建

```bash
docker build -t filecodebox-frontend .
docker run -p 80:80 filecodebox-frontend
```

### Nginx 配置

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8888;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 许可证

MIT License

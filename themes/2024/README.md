# 2024 主题资源说明

## 📁 目录结构

```
themes/2024/
├── assets/
│   └── images/
│       ├── logo.svg          # 主 Logo (64x64px) - 带动画
│       ├── logo-small.svg    # 小尺寸 Logo (32x32px) - 无动画
│       ├── logo-lock.svg     # 安全主题 Logo (64x64px) - 用于密码重置
│       └── favicon.svg       # 网站图标 (16x16px)
├── index.html               # 主页面
├── admin.html               # 管理后台
├── login.html               # 登录页面
├── register.html            # 注册页面
├── forgot-password.html     # 忘记密码页面
└── dashboard.html           # 用户仪表板
```

## 🎨 Logo 文件说明

### logo.svg (主 Logo)
- **尺寸**: 64x64px
- **用途**: 主页面、登录页面、注册页面
- **特色**: 包含传输动画效果
- **颜色**: 蓝色主题 (#2563eb)

### logo-small.svg (小尺寸 Logo)
- **尺寸**: 32x32px
- **用途**: 管理后台、用户仪表板、作为 favicon 备用
- **特色**: 简化版本，无动画
- **颜色**: 紫色主题 (#667eea)

### logo-lock.svg (安全主题 Logo)
- **尺寸**: 64x64px
- **用途**: 忘记密码页面、安全相关功能
- **特色**: 带锁图标，传输箭头暗化
- **颜色**: 红色主题 (#dc3545)

### favicon.svg (网站图标)
- **尺寸**: 16x16px
- **用途**: 浏览器标签页图标
- **特色**: 极简版本，最小化细节
- **颜色**: 标准蓝色主题

## 🔧 技术优势

### 1. 模块化设计
- ✅ Logo 与 HTML 分离，便于维护
- ✅ 单独修改 Logo 不影响页面结构
- ✅ 支持多主题切换

### 2. 性能优化
- ✅ SVG 文件被浏览器缓存
- ✅ 减少 HTML 文件体积
- ✅ 支持 CDN 分发

### 3. 维护性
- ✅ 统一的资源管理
- ✅ 易于批量更新
- ✅ 版本控制友好

## 🎯 使用方式

### 在 HTML 中引用
```html
<!-- 主 Logo -->
<img src="assets/images/logo.svg" alt="FileCodeBox Logo" class="logo-icon"/>

<!-- 小尺寸 Logo -->
<img src="assets/images/logo-small.svg" alt="FileCodeBox Logo" class="logo-icon"/>

<!-- 安全主题 Logo -->
<img src="assets/images/logo-lock.svg" alt="FileCodeBox Security Logo" class="logo-icon"/>

<!-- Favicon -->
<link rel="icon" type="image/svg+xml" href="assets/images/favicon.svg">
```

### CSS 样式
```css
.logo-icon {
    width: 48px;
    height: 48px;
}

/* 或者根据需要调整尺寸 */
.logo-icon.small {
    width: 32px;
    height: 32px;
}
```

## 🚀 自定义建议

### 创建新主题
1. 复制 `themes/2024/` 目录
2. 修改 `assets/images/` 中的 Logo 文件
3. 调整 HTML 文件中的样式和布局
4. 更新 CSS 配色方案

### Logo 定制
1. 保持 SVG 格式以支持缩放
2. 维持相同的 viewBox 比例
3. 确保在小尺寸下的可读性
4. 测试不同背景下的对比度

## 📱 响应式支持

不同设备自动选择合适的 Logo 尺寸：
- **桌面端**: logo.svg (64px)
- **移动端**: logo.svg (48px，CSS 缩放)
- **导航栏**: logo-small.svg (32px)
- **浏览器标签**: favicon.svg (16px)

---

*更新时间: 2025年9月11日*  
*版本: 2024主题 v1.0*

# FileCodeBox Logo 设计文档

## 设计概念

FileCodeBox 的 Logo 设计围绕项目的核心功能展开：**文件快传和分享**。Logo 采用现代扁平化设计风格，融合了技术感和易用性。

## 设计元素

### 1. 文件盒子 📦
- **含义**：体现项目名称中的"Box"，象征文件的存储和管理
- **设计**：圆角矩形盒子，现代简洁
- **颜色**：白色盒体，浅灰色盖子

### 2. 多彩文件 📁
- **含义**：支持多种文件类型的分享
- **设计**：三个不同颜色的文件图标
- **颜色方案**：
  - 蓝色 (#3b82f6) - 文档类文件
  - 绿色 (#10b981) - 数据类文件  
  - 橙色 (#f59e0b) - 媒体类文件

### 3. 传输箭头 ⚡
- **含义**：快速传输和分享功能
- **设计**：流畅的箭头线条，带动画效果
- **颜色**：白色，突出传输概念

### 4. 动画点 ✨
- **含义**：数据流动和实时传输
- **设计**：沿传输路径移动的青色光点
- **效果**：无限循环动画，增强视觉吸引力

## Logo 变体

### 1. 主 Logo (logo.svg)
- **尺寸**：200x200px
- **用途**：官网首页、宣传材料、大尺寸展示
- **特点**：完整版本，包含所有设计元素和动画

### 2. 横版 Logo (logo-horizontal.svg)
- **尺寸**：280x80px
- **用途**：网站头部导航、横幅、邮件签名
- **特点**：Logo + 文字 + 特性标签的组合

### 3. 小尺寸 Logo (logo-small.svg)
- **尺寸**：64x64px
- **用途**：应用图标、按钮、小尺寸展示
- **特点**：简化版本，保留核心元素

### 4. Favicon (favicon.svg)
- **尺寸**：32x32px
- **用途**：浏览器标签页图标、收藏夹图标
- **特点**：最简化版本，无动画效果

## 颜色规范

### 主色调
```css
/* 主蓝色 - 品牌色 */
#2563eb

/* 辅助色彩 */
#3b82f6  /* 文件蓝 */
#10b981  /* 文件绿 */
#f59e0b  /* 文件橙 */
#22d3ee  /* 动画青 */

/* 中性色 */
#ffffff  /* 白色 */
#f3f4f6  /* 浅灰 */
#e5e7eb  /* 边框灰 */
```

### 渐变效果
```css
/* 文字渐变 */
background: linear-gradient(135deg, #667eea, #764ba2);
```

## 使用指南

### 网站集成
1. **Favicon 设置**
```html
<link rel="icon" type="image/svg+xml" href="/favicon.svg">
<link rel="icon" type="image/png" sizes="32x32" href="/logo-small.svg">
```

2. **主题文件集成**
Logo 已集成到以下模板文件：
- `themes/2024/index.html` - 主页面
- `themes/2024/admin.html` - 管理后台

### 自定义主题
创建新主题时，可参考以下 CSS 类：
```css
.logo {
    display: flex;
    align-items: center;
    gap: 15px;
}

.logo-icon {
    width: 48px;
    height: 48px;
}

.logo-text {
    font-size: 2.5em;
    font-weight: bold;
    background: linear-gradient(135deg, #667eea, #764ba2);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
}
```

## 技术实现

### SVG 优势
- **矢量图形**：任意缩放不失真
- **小文件**：优化的文件大小
- **动画支持**：原生 CSS/SVG 动画
- **浏览器兼容**：现代浏览器完全支持

### 响应式设计
Logo 支持多种尺寸，在不同设备上都能良好显示：
- 桌面端：完整版 Logo
- 平板端：横版 Logo  
- 移动端：小尺寸 Logo
- 浏览器标签：Favicon

## 品牌应用

### 适用场景
✅ 官方网站和文档  
✅ GitHub 仓库  
✅ 宣传材料和海报  
✅ 社交媒体头像  
✅ 开发者工具集成  

### 不适用场景  
❌ 极小尺寸（小于 16px）  
❌ 单色印刷（建议使用单色版本）  
❌ 复杂背景（影响可读性）  

## 文件清单

```
logo.svg              # 主 Logo (200x200)
logo-horizontal.svg   # 横版 Logo (280x80)  
logo-small.svg        # 小尺寸 Logo (64x64)
favicon.svg           # Favicon (32x32)
logo-showcase.html    # Logo 展示页面
generate_favicon.sh   # Favicon 生成脚本
```

## 版权说明

FileCodeBox Logo 遵循项目的开源协议，可在符合协议条款的前提下自由使用和修改。

---

*设计日期：2025年9月11日*  
*设计师：GitHub Copilot*  
*项目：FileCodeBox Go版本*

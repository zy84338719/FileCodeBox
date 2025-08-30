# FileCodeBox 测试脚本集合

本目录包含了 FileCodeBox Go 版本的所有测试脚本，用于验证系统的各种功能和性能。

## 🚀 快速开始

### 前提条件
1. 确保 FileCodeBox 服务已启动在 `http://localhost:12345`
2. 确保有执行权限：`chmod +x tests/*.sh`
3. 系统需要安装 `curl` 和 `jq`（部分脚本需要）

### 启动服务
```bash
# 在项目根目录
./filecodebox
```

## 📋 测试脚本分类

### 🔧 核心功能测试

#### `test_api.sh` - API 基础功能测试
**用途**: 测试基本的 API 端点和响应
**测试内容**:
- 获取系统配置信息
- 文本分享功能
- 文件上传功能
- 分享码查询

**运行**: `./test_api.sh`

#### `test_admin.sh` - 管理后台功能测试  
**用途**: 测试管理员功能和权限控制
**测试内容**:
- 管理员登录验证
- 配置管理
- 分享记录管理
- 文件管理功能
- 权限验证

**运行**: `./test_admin.sh`

#### `test_resume_upload.sh` - 断点续传功能测试
**用途**: 测试文件断点续传功能
**测试内容**:
- 分块上传初始化
- 部分分片上传
- 上传状态查询
- 断点续传恢复
- 分片完整性验证
- 上传完成合并
- 取消上传功能

**运行**: `./test_resume_upload.sh`

#### `test_chunk.sh` - 分块上传测试
**用途**: 测试大文件分块上传功能
**测试内容**:
- 分块上传流程
- 文件完整性验证
- 分块合并功能

**运行**: `./test_chunk.sh`

### 💾 存储相关测试

#### `test_storage_management.sh` - 存储管理测试
**用途**: 测试存储系统管理功能
**测试内容**:
- 存储类型切换
- 存储配置管理
- 文件存储验证

**运行**: `./test_storage_management.sh`

#### `test_storage_switch_fix.sh` - 存储切换修复测试
**用途**: 验证存储类型切换功能的修复
**测试内容**:
- Local ↔ WebDAV 切换
- 配置保存验证
- UI 响应测试

**运行**: `./test_storage_switch_fix.sh`

#### `test_webdav_config.sh` - WebDAV 配置测试
**用途**: 测试 WebDAV 存储配置
**测试内容**:
- WebDAV 连接测试
- 配置参数验证
- 文件操作测试

**运行**: `./test_webdav_config.sh`

### 🗄️ 数据库和配置测试

#### `test_database_config.sh` - 数据库配置测试
**用途**: 测试数据库配置系统
**测试内容**:
- 配置读取和保存
- 数据库初始化
- 配置项验证

**运行**: `./test_database_config.sh`

#### `test_date_grouping.sh` - 日期分组测试
**用途**: 测试文件按日期分组存储
**测试内容**:
- 日期目录创建
- 文件分组逻辑
- 路径生成验证

**运行**: `./test_date_grouping.sh`

### 🌐 前端和UI测试

#### `test_web.sh` - Web 界面测试
**用途**: 测试 Web 前端功能
**测试内容**:
- 页面加载测试
- 表单提交功能
- 响应验证

**运行**: `./test_web.sh`

#### `test_ui_features.sh` - UI 功能测试
**用途**: 测试用户界面特性
**测试内容**:
- 界面元素验证
- 用户交互测试
- 样式检查

**运行**: `./test_ui_features.sh`

#### `test_javascript.sh` - JavaScript 功能测试
**用途**: 测试前端 JavaScript 功能
**测试内容**:
- 客户端脚本验证
- 异步请求测试
- 错误处理验证

**运行**: `./test_javascript.sh`

#### `test_progress.sh` - 进度显示测试
**用途**: 测试上传进度显示功能
**测试内容**:
- 上传进度条
- 实时状态更新
- 完成状态验证

**运行**: `./test_progress.sh`

### 🚨 问题诊断和修复测试

#### `test_upload_limit.sh` - 上传限制测试
**用途**: 测试文件上传大小限制
**测试内容**:
- 上传限制验证
- 错误处理测试
- 限制配置测试

**运行**: `./test_upload_limit.sh`

#### `test_download_issue.sh` - 下载问题测试
**用途**: 诊断和测试文件下载问题
**测试内容**:
- 下载链接验证
- 文件完整性检查
- 错误情况处理

**运行**: `./test_download_issue.sh`

#### `diagnose_storage_issue.sh` - 存储问题诊断
**用途**: 诊断存储系统问题
**测试内容**:
- 存储连接状态
- 文件访问权限
- 路径配置检查

**运行**: `./diagnose_storage_issue.sh`

### ⚡ 性能和压力测试

#### `benchmark.sh` - 性能基准测试
**用途**: 测试系统性能和并发能力
**测试内容**:
- 并发文本分享测试
- 并发文件上传测试
- 响应时间统计
- 吞吐量测试

**运行**: `./benchmark.sh`

#### `simple_test.sh` - 简单功能测试
**用途**: 快速验证基本功能
**测试内容**:
- 基础功能检查
- 快速验证流程

**运行**: `./simple_test.sh`

## 🔄 测试套件运行

### 🎯 自动化测试运行器（推荐）

使用提供的自动化测试运行器，支持分类运行和详细报告：

```bash
cd tests

# 运行所有测试（推荐）
./run_all_tests.sh

# 按分类运行测试
./run_all_tests.sh core        # 核心功能测试
./run_all_tests.sh storage     # 存储功能测试  
./run_all_tests.sh frontend    # 前端功能测试
./run_all_tests.sh performance # 性能测试
./run_all_tests.sh resume      # 断点续传测试（新增）
```

**特性**:
- 🕒 自动检查服务器状态
- 📊 生成详细测试报告
- ⏱️  超时保护（60秒/脚本）
- 📝 时间戳和日志记录
- 🎯 分类测试支持

### 手动运行所有测试
```bash
cd tests
for script in *.sh; do
    if [[ "$script" != "run_all_tests.sh" ]]; then
        echo "=== 运行 $script ==="
        ./"$script"
        echo
    fi
done
```

### 按分类运行测试

#### 核心功能测试
```bash
./test_api.sh && ./test_admin.sh && ./test_chunk.sh
```

#### 存储功能测试
```bash
./test_storage_management.sh && ./test_storage_switch_fix.sh && ./test_webdav_config.sh
```

#### 前端功能测试
```bash
./test_web.sh && ./test_ui_features.sh && ./test_javascript.sh
```

#### 性能测试
```bash
./benchmark.sh
```

## 📊 测试结果说明

### 成功标识
- ✅ - 测试通过
- 📈 - 性能正常
- 🔧 - 功能正常

### 失败标识
- ❌ - 测试失败
- ⚠️  - 警告或需要注意
- 🚨 - 严重问题

## 🛠️ 故障排除

### 常见问题

1. **服务器未启动**
   ```bash
   # 启动服务器
   cd /Users/zhangyi/FileCodeBox/go
   ./filecodebox
   ```

2. **权限问题**
   ```bash
   # 添加执行权限
   chmod +x tests/*.sh
   ```

3. **端口占用**
   ```bash
   # 检查端口占用
   lsof -i :12345
   # 杀掉占用进程
   pkill -f filecodebox
   ```

4. **依赖缺失**
   ```bash
   # 安装必要工具
   brew install curl jq  # macOS
   ```

### 调试模式
在脚本开头添加调试选项：
```bash
#!/bin/bash
set -x  # 显示执行的命令
set -e  # 遇到错误立即退出
```

## 📝 测试报告

运行测试后，建议记录以下信息：
- 测试时间
- 环境信息（OS、Go版本等）
- 测试结果汇总
- 发现的问题
- 性能数据

## 🤝 贡献指南

### 添加新测试
1. 创建新的 `.sh` 文件
2. 使用统一的命名规范：`test_[功能名].sh`
3. 包含详细的测试说明注释
4. 更新本 README 文档

### 测试脚本规范
```bash
#!/bin/bash

# 测试功能描述
# 作者：XXX
# 日期：YYYY-MM-DD

BASE_URL="http://localhost:12345"

echo "=== 测试名称 ==="
echo

# 检查前提条件
if ! curl -s --connect-timeout 2 $BASE_URL > /dev/null; then
    echo "❌ 服务器未运行"
    exit 1
fi

# 测试逻辑
echo "1. 测试项目1..."
# 测试代码

echo "✅ 所有测试完成"
```

---

**最后更新**: 2025年8月30日  
**版本**: FileCodeBox Go v1.0  
**维护者**: FileCodeBox Team

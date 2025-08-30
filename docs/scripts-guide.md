# 发布脚本使用指南

本目录包含了用于自动化代码提交、版本管理和发布的脚本工具。

## 脚本列表

### 1. release.sh - 完整发布脚本 🚀

功能最全面的发布脚本，用于正式版本发布。

**主要功能:**
- 自动版本管理和标签创建
- 代码质量检查和测试
- Docker镜像构建
- 自动生成变更日志
- Git提交和推送
- 完整的发布流程

**使用方法:**
```bash
# 基本用法
./release.sh v1.0.0

# 预发布版本
./release.sh v1.1.0-beta --pre-release

# 自定义提交信息
./release.sh v1.0.1 -m "修复重要安全漏洞"

# 包含Docker构建
./release.sh v1.0.0 --build

# 模拟运行（不执行实际操作）
./release.sh v1.0.0 --dry-run

# 强制执行（跳过确认）
./release.sh v1.0.0 --force
```

**参数说明:**
- `-h, --help`: 显示帮助信息
- `-d, --dry-run`: 模拟运行，不执行实际操作
- `-f, --force`: 强制执行，跳过确认提示
- `-m, --message`: 自定义提交信息
- `-p, --pre-release`: 标记为预发布版本
- `-b, --build`: 发布前构建Docker镜像

### 2. quick-push.sh - 快速推送脚本 ⚡

用于日常开发中的快速代码提交和推送。

**主要功能:**
- 快速添加、提交和推送代码
- 交互式提交信息输入
- 简单的Git状态检查

**使用方法:**
```bash
# 交互式使用
./quick-push.sh

# 直接提供提交信息
./quick-push.sh "修复用户登录问题"

# 自动生成时间戳提交信息
./quick-push.sh ""
```

### 3. tag-manager.sh - 标签管理脚本 🏷️

专门用于Git标签的管理操作。

**主要功能:**
- 创建、删除、推送标签
- 列出和查看标签详情
- 标签格式验证
- 远程标签同步

**使用方法:**
```bash
# 创建标签
./tag-manager.sh create v1.0.0
./tag-manager.sh create v1.0.1 -m "修复重要bug"

# 删除标签
./tag-manager.sh delete v1.0.0
./tag-manager.sh delete v1.0.0 -f  # 强制删除

# 列出所有标签
./tag-manager.sh list

# 推送标签
./tag-manager.sh push v1.0.0
./tag-manager.sh push all  # 推送所有标签

# 拉取远程标签
./tag-manager.sh pull

# 查看标签详情
./tag-manager.sh show v1.0.0
```

## 使用场景

### 日常开发提交
```bash
# 快速提交日常开发更改
./quick-push.sh "优化用户界面"
```

### 功能完成发布
```bash
# 完整的版本发布流程
./release.sh v1.1.0 --build
```

### 紧急修复发布
```bash
# 快速修复发布
./release.sh v1.0.1 -m "紧急修复安全漏洞" --force
```

### 预发布测试
```bash
# 创建预发布版本
./release.sh v1.2.0-beta --pre-release --dry-run  # 先模拟
./release.sh v1.2.0-beta --pre-release             # 实际执行
```

### 标签管理
```bash
# 管理发布标签
./tag-manager.sh create v1.0.0-rc1 -m "候选发布版本"
./tag-manager.sh push v1.0.0-rc1
```

## 版本号规范

所有脚本都遵循语义化版本控制 (SemVer) 规范：

- `v1.0.0` - 正式版本
- `v1.0.1` - 补丁版本
- `v1.1.0` - 次要版本
- `v2.0.0` - 主要版本
- `v1.0.0-alpha` - Alpha版本
- `v1.0.0-beta` - Beta版本
- `v1.0.0-rc1` - 候选版本

## 工作流程建议

### 开发流程
1. 日常开发使用 `quick-push.sh` 进行快速提交
2. 功能完成后使用 `release.sh` 进行版本发布
3. 使用 `tag-manager.sh` 管理特殊标签需求

### 发布流程
1. **开发完成** → 使用 `quick-push.sh` 提交最终更改
2. **测试验证** → 使用 `release.sh --dry-run` 模拟发布
3. **正式发布** → 使用 `release.sh` 进行完整发布
4. **标签管理** → 使用 `tag-manager.sh` 进行后续标签操作

## 注意事项

### 权限要求
- 脚本需要执行权限: `chmod +x *.sh`
- 需要Git仓库的推送权限
- Docker构建需要Docker环境

### 环境要求
- Git 2.0+
- Go 1.25+ (用于构建)
- Docker (可选，用于容器构建)
- Bash 4.0+

### 安全建议
- 在生产环境使用前，先用 `--dry-run` 模拟
- 重要发布前备份代码和数据
- 确认远程仓库地址正确
- 定期更新脚本以适应项目变化

### 故障排除

**脚本权限问题:**
```bash
chmod +x release.sh quick-push.sh tag-manager.sh
```

**Git推送失败:**
```bash
# 检查远程仓库配置
git remote -v

# 检查认证状态
ssh -T git@github.com
```

**版本号格式错误:**
```bash
# 正确格式示例
v1.0.0, v1.2.3-beta, v2.0.0-rc1

# 错误格式
1.0.0, v1.0, v1.0.0.1
```

## 自定义配置

可以通过修改脚本顶部的配置变量来自定义行为：

```bash
# 在脚本中修改这些变量
DEFAULT_BRANCH="main"        # 默认分支
DOCKER_REGISTRY="your-registry"  # Docker仓库地址
TEST_COMMAND="go test ./..."     # 测试命令
```

## 脚本集成

这些脚本可以集成到CI/CD流程中：

```yaml
# GitHub Actions 示例
- name: Release
  run: ./release.sh ${{ github.event.inputs.version }} --force
```

```yaml
# GitLab CI 示例
release:
  script:
    - ./release.sh $CI_COMMIT_TAG --force
  only:
    - tags
```

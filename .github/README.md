# GitHub Actions 工作流程

本项目包含了完整的 CI/CD 工作流程，用于自动化构建、测试和发布。

## 工作流程概览

### 1. CI (持续集成) - `.github/workflows/ci.yml`

**触发条件：**
- 推送到 `main` 或 `develop` 分支
- 创建 Pull Request

**功能：**
- 多版本 Go 测试 (1.20, 1.21)
- 代码质量检查 (golangci-lint)
- 安全扫描 (gosec)
- 测试覆盖率报告

### 2. Build and Release - `.github/workflows/build.yml`

**触发条件：**
- 推送标签 (v*.*.*)
- 推送到 main 分支
- Pull Request

**功能：**
- **多平台构建：**
  - Linux (amd64, arm64)
  - macOS (amd64, arm64) 
  - Windows (amd64, arm64)
- **Docker 镜像构建**
- **自动发布：** 标签推送时自动创建 GitHub Release

### 3. Deploy - `.github/workflows/deploy.yml`

**触发条件：**
- 发布新版本
- 手动触发

**功能：**
- Docker 镜像发布到 Docker Hub
- Docker 镜像发布到 GitHub Container Registry

## 使用指南

### 🏗️ 日常开发

1. **推送代码** 到 `main` 或 `develop` 分支会触发 CI 流程
2. **创建 PR** 会运行完整的测试套件
3. **推送到 main** 会额外触发构建流程

### 📦 发布新版本

1. **创建标签：**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. **自动执行：**
   - ✅ 构建多平台可执行文件
   - ✅ 构建 Docker 镜像
   - ✅ 创建 GitHub Release
   - ✅ 上传构建产物

### 🐳 Docker 镜像

**公开镜像：**
```bash
# Docker Hub
docker pull filecodebox/filecodebox:latest
docker pull filecodebox/filecodebox:v1.0.0

# GitHub Container Registry  
docker pull ghcr.io/zy84338719/filecodebox:latest
```

### 📋 配置要求

为了完整使用所有功能，需要在 GitHub 仓库中配置以下 Secrets：

#### 必需的 Secrets

| Secret Name | 描述 | 用途 |
|-------------|------|------|
| `DOCKER_USERNAME` | Docker Hub 用户名 | 发布 Docker 镜像 |
| `DOCKER_PASSWORD` | Docker Hub 密码/Token | 发布 Docker 镜像 |

#### 可选的 Secrets

| Secret Name | 描述 | 用途 |
|-------------|------|------|
| `CODECOV_TOKEN` | Codecov Token | 上传测试覆盖率 |

### 🔧 自定义构建

#### 修改构建平台

编辑 `.github/workflows/build.yml` 中的 `matrix` 部分：

```yaml
strategy:
  matrix:
    include:
      - goos: linux
        goarch: amd64
        output: filecodebox-linux-amd64
      # 添加或删除平台...
```

#### 修改 Docker 配置

编辑 `Dockerfile` 和相关工作流程文件。

#### 自定义发布说明

修改 `.github/workflows/build.yml` 中的 `Generate release notes` 步骤。

## 构建产物

### 可执行文件

每次发布会生成以下文件：

- `filecodebox-linux-amd64.tar.gz`
- `filecodebox-linux-arm64.tar.gz`  
- `filecodebox-darwin-amd64.tar.gz`
- `filecodebox-darwin-arm64.tar.gz`
- `filecodebox-windows-amd64.zip`
- `filecodebox-windows-arm64.zip`

### Docker 镜像

- 支持 `linux/amd64` 和 `linux/arm64` 架构
- 多标签发布：`latest`, `v1.0.0`, `v1.0`, `v1`

## 版本信息

构建的可执行文件包含版本信息：

```bash
./filecodebox -version
```

输出示例：
```
FileCodeBox v1.0.0
Commit: a1b2c3d4e5f6...
Built: 2024-01-01T12:00:00Z
Go Version: go1.21+
```

## 故障排除

### 常见问题

1. **Docker 推送失败**
   - 检查 `DOCKER_USERNAME` 和 `DOCKER_PASSWORD` 是否正确配置
   - 确认 Docker Hub 仓库权限

2. **构建失败**
   - 检查 Go 版本兼容性
   - 查看构建日志中的具体错误信息

3. **测试失败**
   - 确保所有依赖都在 `go.mod` 中正确声明
   - 检查代码质量问题

### 调试技巧

1. **查看工作流日志：** GitHub Actions 标签页
2. **本地测试：** 
   ```bash
   # 运行测试
   go test ./...
   
   # 代码检查
   golangci-lint run
   
   # 构建测试
   go build -v ./...
   ```

3. **手动触发：** 在 Actions 标签页可以手动触发部署工作流

---

更多信息请参考 [GitHub Actions 文档](https://docs.github.com/en/actions)。

# FileCodeBox Docker 多架构构建指南

这个项目支持多架构Docker构建，可以轻松地在不同的平台上运行。

## 快速开始

### 1. 使用构建脚本（推荐）

我们提供了一个便捷的构建脚本 `build-docker.sh`，支持各种构建选项：

```bash
# 给脚本执行权限
chmod +x build-docker.sh

# 构建当前平台的镜像（自动检测架构）
./build-docker.sh

# 构建指定平台的镜像
./build-docker.sh --single linux/amd64    # 构建 AMD64 架构
./build-docker.sh --single linux/arm64    # 构建 ARM64 架构

# 构建并推送多架构镜像到仓库
./build-docker.sh --push

# 自定义镜像名称和标签
./build-docker.sh --name myregistry/filecodebox --tag v1.0.0

# 查看帮助
./build-docker.sh --help
```

### 2. 手动构建

#### 单架构构建（本地使用）

```bash
# AMD64 架构
docker build --platform linux/amd64 -t filecodebox:amd64 .

# ARM64 架构  
docker build --platform linux/arm64 -t filecodebox:arm64 .

# 自动检测当前平台
docker build -t filecodebox:latest .
```

#### 多架构构建（发布使用）

```bash
# 创建并使用 buildx builder
docker buildx create --name multiarch-builder --driver docker-container --bootstrap
docker buildx use multiarch-builder

# 构建多架构镜像并推送
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag your-registry/filecodebox:latest \
  --push .

# 仅构建不推送（会保存到构建缓存中）
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag filecodebox:latest \
  .
```

## 运行容器

### 基础运行

```bash
# 运行容器（端口映射）
docker run -d -p 12345:12345 filecodebox:latest

# 运行容器（自定义端口）
docker run -d -p 8080:12345 filecodebox:latest

# 运行容器（带数据持久化）
docker run -d \
  -p 12345:12345 \
  -v $(pwd)/data:/app/data \
  filecodebox:latest
```

### 使用 Docker Compose

创建 `docker-compose.yml` 文件：

```yaml
version: '3.8'

services:
  filecodebox:
    image: filecodebox:latest
    container_name: filecodebox
    ports:
      - "12345:12345"
    volumes:
      - ./data:/app/data
    restart: unless-stopped
    environment:
      - GIN_MODE=release  # 生产模式
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:12345/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

运行：

```bash
docker-compose up -d
```

## 构建特性

### Dockerfile 特性

- **多阶段构建**：优化镜像大小，仅包含运行时必需的文件
- **多架构支持**：支持 AMD64 和 ARM64 架构
- **安全性**：使用非 root 用户运行，减少安全风险  
- **优化**：使用编译优化参数，减小二进制文件大小

### 支持的架构

- `linux/amd64` - Intel/AMD 64位处理器
- `linux/arm64` - ARM 64位处理器（如 Apple Silicon、树莓派4等）

### 镜像分层

1. **构建层**：基于 `golang:1.25-alpine`
   - 安装构建依赖（gcc、musl-dev、sqlite-dev）
   - 下载 Go 模块
   - 编译应用程序

2. **运行层**：基于 `alpine:latest`
   - 安装运行时依赖（ca-certificates、tzdata、sqlite）
   - 创建非 root 用户
   - 复制编译好的程序和静态文件

## 环境变量

可以通过环境变量配置应用：

```bash
docker run -d \
  -p 12345:12345 \
  -e GIN_MODE=release \
  -e LOG_LEVEL=info \
  filecodebox:latest
```

常用环境变量：
- `GIN_MODE=release` - 设置为生产模式
- `LOG_LEVEL=info` - 设置日志级别

## 数据持久化

建议挂载数据目录以保持数据持久化：

```bash
# 创建数据目录
mkdir -p ./data

# 运行容器并挂载数据目录
docker run -d \
  -p 12345:12345 \
  -v $(pwd)/data:/app/data \
  filecodebox:latest
```

## 故障排除

### 1. 构建问题

如果遇到构建错误，请检查：

```bash
# 检查 Docker 版本
docker version

# 检查 buildx 可用性
docker buildx version

# 清理构建缓存
docker builder prune
```

### 2. 运行问题

```bash
# 查看容器日志
docker logs <container-id>

# 进入容器排查
docker exec -it <container-id> /bin/sh

# 检查容器状态
docker ps -a
```

### 3. 网络问题

```bash
# 检查端口占用
lsof -i :12345

# 测试容器网络
curl http://localhost:12345/health
```

## 性能优化

### 构建优化

1. **使用构建缓存**：
   ```bash
   # 启用内联缓存
   docker buildx build --cache-from type=inline --cache-to type=inline .
   ```

2. **并行构建**：
   ```bash
   # 设置并行构建数量
   docker buildx build --platform linux/amd64,linux/arm64 .
   ```

### 运行优化

1. **资源限制**：
   ```bash
   docker run -d \
     --memory=512m \
     --cpus=1.0 \
     -p 12345:12345 \
     filecodebox:latest
   ```

2. **健康检查**：
   ```bash
   docker run -d \
     --health-cmd="wget --quiet --tries=1 --spider http://localhost:12345/health || exit 1" \
     --health-interval=30s \
     --health-timeout=10s \
     --health-retries=3 \
     -p 12345:12345 \
     filecodebox:latest
   ```

## 版本发布

推荐的版本发布流程：

1. **标记版本**：
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```

2. **构建并推送**：
   ```bash
   ./build-docker.sh --name your-registry/filecodebox --tag v1.0.0 --push
   ./build-docker.sh --name your-registry/filecodebox --tag latest --push
   ```

## 许可证

本项目基于原有许可证发布。

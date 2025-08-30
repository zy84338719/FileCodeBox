# Docker Compose 部署指南

本目录包含多个 Docker Compose 配置文件，适用于不同的部署场景。

## 文件说明

- `docker-compose.yml` - 开发/测试环境配置
- `docker-compose.prod.yml` - 生产环境配置
- `nginx/nginx.conf` - Nginx 反向代理配置

## 快速开始

### 开发环境

```bash
# 启动开发环境
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 生产环境

```bash
# 构建并启动生产环境
docker-compose -f docker-compose.prod.yml up -d

# 包含 Nginx 反向代理
docker-compose -f docker-compose.prod.yml --profile with-nginx up -d

# 查看服务状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f filecodebox-go
```

## 配置说明

### 开发环境配置 (docker-compose.yml)

- **端口映射**: 12345:12345
- **数据持久化**: 本地 `./data` 和 `./themes` 目录
- **健康检查**: 30秒间隔检查服务状态
- **日志管理**: 限制日志文件大小和数量
- **资源限制**: 内存限制 512MB

### 生产环境配置 (docker-compose.prod.yml)

- **数据卷**: 使用 Docker 命名卷进行数据持久化
- **网络隔离**: 自定义网络确保服务间通信安全
- **资源限制**: CPU 和内存资源限制
- **安全配置**: 禁用新权限、只读文件系统等
- **Nginx 支持**: 可选的反向代理和负载均衡

## 环境变量

可以创建 `.env` 文件来自定义配置：

```bash
# .env 文件示例
COMPOSE_PROJECT_NAME=filecodebox
FILECODEBOX_PORT=12345
TZ=Asia/Shanghai
GIN_MODE=release

# 资源限制
MEMORY_LIMIT=1g
CPU_LIMIT=1.0
MEMORY_RESERVATION=512m
CPU_RESERVATION=0.5
```

## Nginx 配置

### 启用 HTTPS

1. 将 SSL 证书放入 `nginx/ssl/` 目录
2. 修改 `nginx/nginx.conf` 中的 HTTPS 配置
3. 取消注释 HTTPS server 块

### 自定义域名

修改 `nginx/nginx.conf` 中的 `server_name`：

```nginx
server_name your-domain.com;
```

## 健康检查

服务包含健康检查配置，可以通过以下命令查看服务健康状态：

```bash
# 查看服务健康状态
docker-compose ps

# 检查特定服务健康状态
docker inspect --format='{{.State.Health.Status}}' filecodebox
```

## 备份和恢复

### 数据备份

```bash
# 备份数据目录
docker run --rm -v filecodebox-data:/data -v $(pwd):/backup ubuntu tar czf /backup/filecodebox-backup-$(date +%Y%m%d).tar.gz -C /data .

# 或者直接备份本地目录（开发环境）
tar czf filecodebox-backup-$(date +%Y%m%d).tar.gz data/
```

### 数据恢复

```bash
# 恢复数据
docker run --rm -v filecodebox-data:/data -v $(pwd):/backup ubuntu tar xzf /backup/filecodebox-backup-YYYYMMDD.tar.gz -C /data
```

## 监控和日志

### 查看实时日志

```bash
# 所有服务日志
docker-compose logs -f

# 特定服务日志
docker-compose logs -f filecodebox-go

# 最近 100 行日志
docker-compose logs --tail=100 filecodebox-go
```

### 资源监控

```bash
# 查看资源使用情况
docker stats filecodebox

# 查看容器详细信息
docker inspect filecodebox
```

## 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   lsof -i :12345
   
   # 修改端口映射
   # 在 docker-compose.yml 中修改 ports 配置
   ```

2. **权限问题**
   ```bash
   # 确保数据目录权限正确
   sudo chown -R 1000:1000 ./data ./themes
   chmod -R 755 ./data ./themes
   ```

3. **容器无法启动**
   ```bash
   # 查看详细错误信息
   docker-compose logs filecodebox-go
   
   # 重建容器
   docker-compose down
   docker-compose up --build -d
   ```

4. **网络问题**
   ```bash
   # 检查网络配置
   docker network ls
   docker network inspect filecodebox_filecodebox-network
   ```

## 安全建议

1. **修改默认密码**: 启动后立即修改管理员密码
2. **使用 HTTPS**: 生产环境启用 SSL/TLS
3. **网络隔离**: 使用自定义网络限制容器间通信
4. **资源限制**: 设置适当的 CPU 和内存限制
5. **定期备份**: 建立自动备份策略
6. **日志监控**: 监控应用和访问日志

## 扩展部署

### 多实例部署

```yaml
# docker-compose.scale.yml
version: '3.8'
services:
  filecodebox-go:
    # ... 基础配置
    deploy:
      replicas: 3
```

```bash
# 启动多个实例
docker-compose -f docker-compose.scale.yml up -d --scale filecodebox-go=3
```

### 与其他服务集成

可以与 Redis、MySQL、MinIO 等服务集成，创建完整的服务栈。

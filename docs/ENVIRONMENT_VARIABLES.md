# FileCodeBox 环境变量配置

FileCodeBox 现在支持通过环境变量进行配置，便于容器化部署和配置管理。

## 支持的环境变量

### 🚀 基础配置

| 环境变量 | 类型 | 默认值 | 说明 |
|---------|------|--------|------|
| `PORT` | int | 12345 | HTTP 服务端口 |
| `HOST` | string | 0.0.0.0 | 绑定的 IP 地址 |
| `DATA_PATH` | string | ./data | 数据存储目录路径 |
| `PRODUCTION` | int | 0 | 生产模式 (0=调试模式, 1=生产模式) |

### 📁 上传配置

| 环境变量 | 类型 | 默认值 | 说明 |
|---------|------|--------|------|
| `OPEN_UPLOAD` | int | 1 | 启用上传功能 (0=禁用, 1=启用) |
| `UPLOAD_SIZE` | int64 | 10485760 | 最大上传文件大小 (字节) |

### 🗄️ 数据库配置

| 环境变量 | 类型 | 默认值 | 说明 |
|---------|------|--------|------|
| `DATABASE_TYPE` | string | sqlite | 数据库类型 (sqlite/mysql/postgres) |
| `DATABASE_HOST` | string | localhost | 数据库主机地址 |
| `DATABASE_PORT` | int | 3306 | 数据库端口 |
| `DATABASE_NAME` | string | filecodebox | 数据库名称 |
| `DATABASE_USER` | string | root | 数据库用户名 |
| `DATABASE_PASS` | string | "" | 数据库密码 |
| `DATABASE_SSL` | string | disable | SSL 模式 (disable/require/verify-full) |

## 🔄 兼容性

为了保持向后兼容，程序同时支持新旧两套环境变量命名：

### 新命名规范 (推荐)
- `DATABASE_TYPE`
- `DATABASE_HOST`
- `DATABASE_PORT` 等

### 旧命名规范 (兼容)
- `DB_TYPE`
- `DB_HOST`
- `DB_PORT` 等

**优先级**: 新命名规范的环境变量优先级高于旧命名规范。

## 📋 使用示例

### Docker 运行
```bash
docker run -d \
  --name filecodebox \
  -p 8080:8080 \
  -e PORT=8080 \
  -e HOST=0.0.0.0 \
  -e DATA_PATH=/app/data \
  -e DATABASE_TYPE=sqlite \
  -e UPLOAD_SIZE=52428800 \
  -e PRODUCTION=1 \
  -v ./data:/app/data \
  zy84338719/filecodebox:v1.3.1
```

### Docker Compose
```yaml
version: "3.8"
services:
  filecodebox:
    image: zy84338719/filecodebox:v1.3.1
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - HOST=0.0.0.0
      - DATA_PATH=/app/data
      - DATABASE_TYPE=sqlite
      - UPLOAD_SIZE=52428800
      - PRODUCTION=1
    volumes:
      - ./data:/app/data
```

### 系统服务
```bash
export PORT=8080
export HOST=127.0.0.1
export DATA_PATH=/var/lib/filecodebox
export DATABASE_TYPE=mysql
export DATABASE_HOST=localhost
export DATABASE_USER=filecodebox
export DATABASE_PASS=your_password
export DATABASE_NAME=filecodebox

./filecodebox
```

## 🔒 MySQL 配置示例

```bash
export DATABASE_TYPE=mysql
export DATABASE_HOST=mysql.example.com
export DATABASE_PORT=3306
export DATABASE_NAME=filecodebox
export DATABASE_USER=filecodebox_user
export DATABASE_PASS=secure_password
export DATABASE_SSL=disable
```

## 🐘 PostgreSQL 配置示例

```bash
export DATABASE_TYPE=postgres
export DATABASE_HOST=postgres.example.com
export DATABASE_PORT=5432
export DATABASE_NAME=filecodebox
export DATABASE_USER=filecodebox_user
export DATABASE_PASS=secure_password
export DATABASE_SSL=require
```

## 📝 注意事项

1. **数据目录**: 确保 `DATA_PATH` 指向的目录存在且有写权限
2. **端口冲突**: 确保 `PORT` 指定的端口未被占用
3. **数据库连接**: 使用外部数据库时，确保数据库服务可访问
4. **文件大小**: `UPLOAD_SIZE` 单位为字节，建议根据需求调整
5. **生产模式**: 生产环境建议设置 `PRODUCTION=1`

## 🔍 验证配置

启动程序后，可以通过日志输出验证配置是否生效：

```
INFO[2025-09-11T15:29:33+08:00] HTTP服务器启动在 127.0.0.1:8080
INFO[2025-09-11T15:29:33+08:00] 访问地址: http://127.0.0.1:8080
```

或访问 `/api/config` 端点查看当前配置。

# 数据库配置指南

FileCodeBox 支持三种数据库类型：SQLite、MySQL 和 PostgreSQL。

## 配置方式

### 1. 环境变量配置（推荐）

通过环境变量配置数据库连接信息：

```bash
# 数据库类型：sqlite, mysql, postgres
export DB_TYPE="mysql"

# 数据库连接信息
export DB_HOST="localhost"
export DB_PORT="3306"
export DB_NAME="filecodebox"
export DB_USER="your_username"
export DB_PASS="your_password"

# PostgreSQL 特有配置
export DB_SSL="disable"  # disable, require, verify-full
```

### 2. Docker Compose 配置

```yaml
version: '3.8'

services:
  app:
    image: filecodebox:latest
    environment:
      - DB_TYPE=mysql
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_NAME=filecodebox
      - DB_USER=root
      - DB_PASS=your_password
    depends_on:
      - mysql

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: your_password
      MYSQL_DATABASE: filecodebox
    volumes:
      - mysql_data:/var/lib/mysql

volumes:
  mysql_data:
```

## 数据库特定配置

### SQLite（默认）

- 不需要额外配置
- 数据库文件存储在 `./data/filecodebox.db`
- 适合单机部署和测试环境

```bash
export DB_TYPE="sqlite"
```

### MySQL

```bash
export DB_TYPE="mysql"
export DB_HOST="localhost"
export DB_PORT="3306"
export DB_NAME="filecodebox"
export DB_USER="root"
export DB_PASS="your_password"
```

**MySQL 要求：**
- MySQL 5.7+ 或 MariaDB 10.3+
- 支持 UTF8MB4 字符集
- 建议创建专用数据库和用户

**创建数据库示例：**
```sql
CREATE DATABASE filecodebox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'filecodebox'@'%' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON filecodebox.* TO 'filecodebox'@'%';
FLUSH PRIVILEGES;
```

### PostgreSQL

```bash
export DB_TYPE="postgres"
export DB_HOST="localhost"
export DB_PORT="5432"
export DB_NAME="filecodebox"
export DB_USER="postgres"
export DB_PASS="your_password"
export DB_SSL="disable"
```

**PostgreSQL 要求：**
- PostgreSQL 12+
- 支持 UUID 扩展（自动创建）

**创建数据库示例：**
```sql
CREATE DATABASE filecodebox;
CREATE USER filecodebox WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE filecodebox TO filecodebox;
```

## 数据库连接参数

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| DB_TYPE | sqlite | 数据库类型 |
| DB_HOST | localhost | 数据库主机 |
| DB_PORT | 3306 | 数据库端口 |
| DB_NAME | filecodebox | 数据库名称 |
| DB_USER | root | 数据库用户名 |
| DB_PASS | (空) | 数据库密码 |
| DB_SSL | disable | SSL模式（仅PostgreSQL） |

## 数据迁移

应用启动时会自动执行数据库迁移，创建必要的表结构：

- `file_codes` - 文件分享记录
- `upload_chunks` - 分片上传记录
- `key_values` - 配置存储
- `users` - 用户信息
- `user_sessions` - 用户会话

## 性能优化建议

### MySQL
```sql
-- 优化配置建议
SET GLOBAL innodb_buffer_pool_size = 256M;
SET GLOBAL max_connections = 200;
SET GLOBAL innodb_file_per_table = ON;
```

### PostgreSQL
```sql
-- 优化配置建议
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
SELECT pg_reload_conf();
```

## 故障排除

### 常见错误

1. **连接超时**
   - 检查数据库服务是否运行
   - 验证网络连接和防火墙设置

2. **认证失败**
   - 验证用户名和密码
   - 检查用户权限

3. **数据库不存在**
   - 确保数据库已创建
   - 检查数据库名称是否正确

### 日志排查

启用详细日志查看数据库连接问题：

```bash
# 设置日志级别
export LOG_LEVEL=debug
```

## 备份建议

### SQLite
```bash
# 备份
cp ./data/filecodebox.db ./backup/filecodebox_$(date +%Y%m%d).db

# 恢复
cp ./backup/filecodebox_20240101.db ./data/filecodebox.db
```

### MySQL
```bash
# 备份
mysqldump -u root -p filecodebox > backup_$(date +%Y%m%d).sql

# 恢复
mysql -u root -p filecodebox < backup_20240101.sql
```

### PostgreSQL
```bash
# 备份
pg_dump -U postgres filecodebox > backup_$(date +%Y%m%d).sql

# 恢复
psql -U postgres filecodebox < backup_20240101.sql
```

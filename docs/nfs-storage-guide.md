# NFS存储配置指南

FileCodeBox支持使用NFS（Network File System）作为存储后端，适用于需要高性能网络存储的企业环境。

## 功能特性

- **高性能网络存储**: 支持NFS v3/v4协议
- **自动挂载管理**: 可配置自动挂载和重连机制
- **灵活配置**: 支持多种挂载选项和超时设置
- **故障恢复**: 内置重试和错误恢复机制
- **权限管理**: 支持NFS权限和访问控制

## 前置要求

### NFS服务器端

1. **安装NFS服务器**：
```bash
# Ubuntu/Debian
sudo apt-get install nfs-kernel-server

# CentOS/RHEL
sudo yum install nfs-utils
```

2. **配置导出目录**：
```bash
# 编辑 /etc/exports
sudo nano /etc/exports

# 添加导出配置
/nfs/storage *(rw,sync,no_subtree_check,no_root_squash)
```

3. **启动NFS服务**：
```bash
sudo systemctl start nfs-kernel-server
sudo systemctl enable nfs-kernel-server
sudo exportfs -a
```

### NFS客户端（FileCodeBox运行环境）

1. **安装NFS客户端**：
```bash
# Ubuntu/Debian
sudo apt-get install nfs-common

# CentOS/RHEL
sudo yum install nfs-utils

# macOS
# 通常已预装，如需要可通过 brew install nfs-utils
```

2. **创建挂载点**：
```bash
sudo mkdir -p /mnt/nfs
```

## 配置方式

### 1. 环境变量配置

```bash
export NFS_SERVER="192.168.1.100"          # NFS服务器地址
export NFS_PATH="/nfs/storage"              # NFS导出路径
export NFS_MOUNT_POINT="/mnt/nfs"           # 本地挂载点
export NFS_VERSION="4"                      # NFS版本（3, 4, 4.1）
export NFS_OPTIONS="rw,sync,hard,intr"      # 挂载选项
export NFS_TIMEOUT="30"                     # 超时时间（秒）
export NFS_AUTO_MOUNT="1"                   # 自动挂载（0-禁用，1-启用）
export NFS_RETRY_COUNT="3"                  # 重试次数
export NFS_SUB_PATH="filebox_storage"       # 存储子路径
```

### 2. 配置文件

在应用配置中添加：
```json
{
  "file_storage": "nfs",
  "nfs_server": "192.168.1.100",
  "nfs_path": "/nfs/storage",
  "nfs_mount_point": "/mnt/nfs",
  "nfs_version": "4",
  "nfs_options": "rw,sync,hard,intr",
  "nfs_timeout": 30,
  "nfs_auto_mount": 1,
  "nfs_retry_count": 3,
  "nfs_sub_path": "filebox_storage"
}
```

### 3. 管理API配置

通过管理员API动态配置：

```bash
# 获取管理员token
TOKEN=$(curl -s -X POST http://localhost:12345/admin/login \
  -H "Content-Type: application/json" \
  -d '{"password": "your_admin_password"}' | \
  jq -r '.detail.token')

# 更新NFS配置
curl -X PUT http://localhost:12345/admin/storage \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "storage_type": "nfs",
    "config": {
      "server": "192.168.1.100",
      "nfs_path": "/nfs/storage",
      "mount_point": "/mnt/nfs",
      "version": "4",
      "options": "rw,sync,hard,intr",
      "timeout": 30,
      "auto_mount": true,
      "retry_count": 3,
      "sub_path": "filebox_storage"
    }
  }'

# 切换到NFS存储
curl -X POST http://localhost:12345/admin/storage/switch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"storage_type": "nfs"}'
```

## 配置参数详解

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `nfs_server` | NFS服务器IP地址或域名 | - | `192.168.1.100` |
| `nfs_path` | NFS服务器导出路径 | `/nfs/storage` | `/shared/filecodebox` |
| `nfs_mount_point` | 本地挂载点 | `/mnt/nfs` | `/mount/nfs_storage` |
| `nfs_version` | NFS协议版本 | `4` | `3`, `4`, `4.1` |
| `nfs_options` | 挂载选项 | `rw,sync,hard,intr` | `rw,sync,soft,intr` |
| `nfs_timeout` | 超时时间（秒） | `30` | `60` |
| `nfs_auto_mount` | 自动挂载 | `0` | `0`(禁用), `1`(启用) |
| `nfs_retry_count` | 重试次数 | `3` | `5` |
| `nfs_sub_path` | 存储子路径 | `filebox_storage` | `files` |

## 挂载选项说明

常用的NFS挂载选项：

- **`rw/ro`**: 读写/只读模式
- **`sync/async`**: 同步/异步写入
- **`hard/soft`**: 硬挂载/软挂载
  - `hard`: 网络中断时无限重试（推荐）
  - `soft`: 网络中断时返回错误
- **`intr`**: 允许中断NFS调用
- **`noatime`**: 不更新访问时间（提高性能）
- **`rsize/wsize`**: 读写缓冲区大小
- **`timeo`**: 超时时间（1/10秒）
- **`retrans`**: 重传次数

## 性能优化建议

### 1. 网络优化
```bash
# 增大读写缓冲区
mount -t nfs -o rsize=32768,wsize=32768,hard,intr server:/path /mnt/nfs
```

### 2. 缓存优化
```bash
# 启用属性缓存
mount -t nfs -o ac,acregmin=3,acregmax=60,acdirmin=30,acdirmax=60 server:/path /mnt/nfs
```

### 3. NFS v4 优化
```bash
# 使用NFS v4.1启用多路径
mount -t nfs -o vers=4.1,proto=tcp,hard,intr server:/path /mnt/nfs
```

## 安全配置

### 1. 网络安全
```bash
# 使用Kerberos认证
mount -t nfs -o sec=krb5 server:/path /mnt/nfs

# 限制客户端IP
# 在 /etc/exports 中：
/nfs/storage 192.168.1.0/24(rw,sync,no_subtree_check)
```

### 2. 文件权限
```bash
# 设置适当的权限
sudo chown -R filecodebox:filecodebox /mnt/nfs/filebox_storage
sudo chmod -R 755 /mnt/nfs/filebox_storage
```

## 故障排除

### 1. 挂载失败

**检查网络连接**：
```bash
ping nfs_server_ip
telnet nfs_server_ip 2049
```

**检查NFS服务**：
```bash
rpcinfo -p nfs_server_ip
showmount -e nfs_server_ip
```

**检查权限**：
```bash
# 在NFS服务器上
sudo exportfs -v
```

### 2. 性能问题

**检查网络延迟**：
```bash
ping -c 10 nfs_server_ip
```

**监控NFS统计**：
```bash
nfsstat -c  # 客户端统计
nfsstat -s  # 服务器统计
```

**检查挂载状态**：
```bash
mount | grep nfs
cat /proc/mounts | grep nfs
```

### 3. 常见错误

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| `Permission denied` | 权限配置错误 | 检查exports配置和文件权限 |
| `Stale file handle` | 文件句柄失效 | 重新挂载NFS |
| `Connection refused` | NFS服务未启动 | 启动NFS服务 |
| `No such file or directory` | 路径不存在 | 检查导出路径 |

## 测试脚本

使用提供的测试脚本验证NFS配置：

```bash
# 运行NFS存储测试
./test_nfs_storage.sh

# 自定义配置测试
NFS_SERVER=your_server ./test_nfs_storage.sh
```

## 监控和维护

### 1. 健康检查
- 定期检查挂载状态
- 监控NFS性能统计
- 检查网络连接质量

### 2. 自动化脚本
```bash
#!/bin/bash
# NFS健康检查脚本
if ! mountpoint -q /mnt/nfs; then
    echo "NFS not mounted, attempting to mount..."
    mount -t nfs server:/path /mnt/nfs
fi

# 测试读写
if ! touch /mnt/nfs/test_file 2>/dev/null; then
    echo "NFS write test failed"
    exit 1
fi
rm -f /mnt/nfs/test_file
```

### 3. 日志监控
```bash
# 查看NFS相关日志
tail -f /var/log/syslog | grep nfs
dmesg | grep -i nfs
```

## 最佳实践

1. **使用硬挂载**: 确保数据一致性
2. **配置适当超时**: 平衡性能和可靠性
3. **启用自动挂载**: 提高服务可用性
4. **定期备份**: NFS不是备份解决方案
5. **监控性能**: 及时发现和解决问题
6. **网络冗余**: 考虑使用多个NFS服务器
7. **安全加固**: 使用防火墙和认证机制

## 与其他存储的比较

| 特性 | NFS | 本地存储 | S3 | WebDAV |
|------|-----|----------|----|---------| 
| 性能 | 高 | 最高 | 中 | 中 |
| 扩展性 | 高 | 低 | 最高 | 中 |
| 可靠性 | 高 | 中 | 最高 | 中 |
| 复杂度 | 中 | 低 | 低 | 低 |
| 成本 | 中 | 低 | 高 | 低 |

NFS存储适合需要高性能网络存储且有专业运维团队的企业环境。

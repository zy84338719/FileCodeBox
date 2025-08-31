# FileCodeBox MCP Server 使用指南

## 概述

FileCodeBox MCP (Model Context Protocol) Server 为 AI 模型和 LLM 提供了与 FileCodeBox 文件分享系统的直接集成能力。通过 MCP 协议，AI 助手可以：

- 创建和管理文件分享
- 查看系统状态和统计信息  
- 管理用户和存储配置
- 执行系统维护任务

## 启用 MCP 服务器

### 环境变量配置

```bash
# 启用 MCP 服务器
export ENABLE_MCP_SERVER=true

# 设置 MCP 服务器端口（可选，默认 8081）
export MCP_PORT=8081
```

### 启动应用

```bash
./filecodebox-mcp
```

当启用 MCP 服务器时，你会看到类似以下的日志输出：

```
INFO[2025-01-01T12:00:00Z] 启动 MCP 服务器，监听端口: 8081
INFO[2025-01-01T12:00:00Z] 应用初始化完成
```

## MCP 工具列表

### 1. 文本分享工具
- **工具名**: `share_text`
- **功能**: 创建文本分享
- **参数**:
  - `text` (必填): 要分享的文本内容
  - `expire_value` (可选): 过期数值，默认 1
  - `expire_style` (可选): 过期类型 (minute/hour/day/count)，默认 day

### 2. 获取分享工具
- **工具名**: `get_share`
- **功能**: 根据分享代码获取分享信息
- **参数**:
  - `code` (必填): 分享代码

### 3. 列出分享工具
- **工具名**: `list_shares`
- **功能**: 分页列出所有分享记录
- **参数**:
  - `page` (可选): 页码，默认 1
  - `size` (可选): 每页大小，默认 10

### 4. 删除分享工具
- **工具名**: `delete_share`
- **功能**: 删除指定的分享记录
- **参数**:
  - `code` (必填): 要删除的分享代码

### 5. 系统状态工具
- **工具名**: `get_system_status`
- **功能**: 获取系统整体状态和统计信息

### 6. 存储信息工具
- **工具名**: `get_storage_info`
- **功能**: 获取存储配置和状态信息

### 7. 用户列表工具
- **工具名**: `list_users`
- **功能**: 列出系统用户（仅管理员）
- **参数**:
  - `page` (可选): 页码，默认 1
  - `size` (可选): 每页大小，默认 10

### 8. 清理过期文件工具
- **工具名**: `cleanup_expired`
- **功能**: 清理过期的分享文件
- **参数**:
  - `dry_run` (可选): 是否为试运行，默认 false

## MCP 资源列表

### 1. 系统配置资源
- **URI**: `config://system`
- **功能**: 获取系统配置信息

### 2. 系统状态资源
- **URI**: `status://system`
- **功能**: 获取系统运行状态

### 3. 系统统计资源
- **URI**: `stats://system`
- **功能**: 获取系统统计数据

### 4. 存储资源
- **URI**: `storage://info`
- **功能**: 获取存储配置和状态

## 使用示例

### 连接 MCP 服务器

```python
import json
import socket

# 连接到 MCP 服务器
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(('localhost', 8081))

# 发送初始化请求
init_request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {
            "roots": {
                "listChanged": True
            },
            "sampling": {}
        },
        "clientInfo": {
            "name": "test-client",
            "version": "1.0.0"
        }
    }
}

sock.send((json.dumps(init_request) + '\n').encode())
response = sock.recv(4096).decode()
print("初始化响应:", response)
```

### 创建文本分享

```python
share_request = {
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
        "name": "share_text",
        "arguments": {
            "text": "这是一个测试分享",
            "expire_value": 7,
            "expire_style": "day"
        }
    }
}

sock.send((json.dumps(share_request) + '\n').encode())
response = sock.recv(4096).decode()
print("分享响应:", response)
```

### 获取系统状态

```python
status_request = {
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
        "name": "get_system_status",
        "arguments": {}
    }
}

sock.send((json.dumps(status_request) + '\n').encode())
response = sock.recv(4096).decode()
print("系统状态:", response)
```

## 集成 AI 助手

### Claude Desktop 配置

在 Claude Desktop 的配置文件中添加：

```json
{
  "mcpServers": {
    "filecodebox": {
      "command": "nc",
      "args": ["localhost", "8081"],
      "env": {}
    }
  }
}
```

### VS Code + MCP 扩展

1. 安装 MCP 相关扩展
2. 在设置中配置 MCP 服务器：
   ```json
   {
     "mcp.servers": [
       {
         "name": "FileCodeBox",
         "transport": {
           "type": "tcp",
           "host": "localhost",
           "port": 8081
         }
       }
     ]
   }
   ```

## 错误处理

MCP 服务器遵循 JSON-RPC 2.0 错误规范：

- `-32700`: 解析错误
- `-32600`: 无效请求
- `-32601`: 方法未找到
- `-32602`: 无效参数
- `-32603`: 内部错误
- `-32000`: 工具执行错误

## 安全考虑

1. **网络安全**: MCP 服务器默认监听所有接口，建议在生产环境中配置防火墙
2. **认证**: 当前版本不包含认证机制，请确保网络环境安全
3. **权限**: 工具调用会使用系统配置的权限级别

## 调试和日志

MCP 服务器的日志会输出到主应用的日志系统中，包括：

- 连接建立和断开
- 工具调用和执行结果
- 错误和警告信息

使用以下命令查看详细日志：

```bash
./filecodebox-mcp 2>&1 | grep "MCP"
```

## 性能和限制

- **并发连接**: 支持多个客户端同时连接
- **消息大小**: 建议单个消息不超过 1MB
- **工具超时**: 工具执行超时时间为 30 秒
- **资源限制**: 遵循主应用的资源配置和限制

## 故障排除

### 常见问题

1. **连接被拒绝**
   - 检查 MCP 服务器是否已启用 (`ENABLE_MCP_SERVER=true`)
   - 验证端口是否正确配置
   - 检查防火墙设置

2. **工具调用失败**
   - 验证参数格式和类型
   - 检查系统权限和配置
   - 查看应用日志获取详细错误信息

3. **性能问题**
   - 监控内存和 CPU 使用情况
   - 考虑减少并发连接数
   - 优化数据库查询

### 诊断命令

```bash
# 检查 MCP 服务器是否在运行
netstat -ln | grep 8081

# 测试连接
telnet localhost 8081

# 查看进程
ps aux | grep filecodebox-mcp
```

## 更新和维护

MCP 服务器与主应用集成，更新主应用即可获得 MCP 功能的更新。建议定期：

1. 备份数据库和配置
2. 测试 MCP 功能
3. 监控系统性能
4. 更新客户端配置

## 贡献和反馈

如果遇到问题或有改进建议，请在项目仓库中提交 Issue 或 Pull Request。

---

*本文档版本: 1.0.0*  
*最后更新: 2025-08-31*

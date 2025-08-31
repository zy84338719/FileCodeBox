# FileCodeBox MCP Server 集成

## 概述

FileCodeBox 现已集成 **Model Context Protocol (MCP) Server** 功能！这允许 AI 模型和大语言模型 (LLM) 直接与 FileCodeBox 文件分享系统进行交互。

## 🎯 主要功能

### 核心能力
- **文件分享管理**: 创建、查看、删除文本和文件分享
- **系统监控**: 获取系统状态、统计信息、存储配置
- **用户管理**: 查看用户列表和信息（管理员权限）
- **维护操作**: 清理过期文件、测试存储连接

### MCP 工具清单
1. `share_text` - 创建文本分享
2. `get_share` - 获取分享信息
3. `list_shares` - 列出分享记录
4. `delete_share` - 删除分享
5. `get_system_status` - 系统状态
6. `get_storage_info` - 存储信息
7. `list_users` - 用户列表
8. `cleanup_expired` - 清理过期文件

### MCP 资源
1. `config://system` - 系统配置
2. `status://system` - 系统状态
3. `stats://system` - 统计数据
4. `storage://info` - 存储信息

## 🚀 快速开始

### 1. 启用 MCP 服务器

```bash
# 设置环境变量
export ENABLE_MCP_SERVER=true
export MCP_PORT=8081  # 可选，默认8081

# 启动应用
./filecodebox-mcp
```

### 2. 测试连接

```bash
# 使用提供的测试客户端
python3 test_mcp_client.py

# 或者手动测试连接
telnet localhost 8081
```

### 3. 集成 AI 助手

#### Claude Desktop
```json
{
  "mcpServers": {
    "filecodebox": {
      "command": "nc",
      "args": ["localhost", "8081"]
    }
  }
}
```

#### 自定义客户端
```python
import json
import socket

sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
sock.connect(('localhost', 8081))

# 发送初始化请求
init_request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {"roots": {"listChanged": True}},
        "clientInfo": {"name": "my-client", "version": "1.0.0"}
    }
}
```

## 📖 使用示例

### 创建文本分享
```python
# 通过 MCP 工具创建分享
{
    "method": "tools/call",
    "params": {
        "name": "share_text",
        "arguments": {
            "text": "Hello, World!",
            "expire_value": 7,
            "expire_style": "day"
        }
    }
}
```

### 获取系统状态
```python
# 获取完整的系统状态
{
    "method": "tools/call",
    "params": {
        "name": "get_system_status",
        "arguments": {}
    }
}
```

## 🛠️ 技术架构

### MCP 协议实现
- **协议版本**: 2024-11-05
- **传输方式**: TCP Socket (JSON-RPC 2.0)
- **并发支持**: 多客户端同时连接
- **错误处理**: 完整的 JSON-RPC 错误规范

### 集成架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AI/LLM 客户端  │ -> │   MCP 服务器     │ -> │  FileCodeBox    │
│   (Claude等)    │    │   (TCP:8081)    │    │   核心服务       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 代码结构
```
internal/mcp/
├── protocol.go     # MCP 协议定义
├── server.go       # MCP 服务器核心
└── filecodebox.go  # FileCodeBox 特定集成
```

## 🔧 配置选项

### 环境变量
| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `ENABLE_MCP_SERVER` | `false` | 是否启用 MCP 服务器 |
| `MCP_PORT` | `8081` | MCP 服务器监听端口 |

### 运行时配置
- **自动发现**: MCP 服务器会自动检测系统配置
- **权限继承**: 使用主应用的权限和配置
- **日志集成**: 日志输出到主应用日志系统

## 🔍 监控和调试

### 日志输出
```bash
# 查看 MCP 相关日志
./filecodebox-mcp 2>&1 | grep "MCP"

# 完整日志
./filecodebox-mcp --debug
```

### 健康检查
```bash
# 检查服务器状态
netstat -ln | grep 8081

# 测试基本连接
telnet localhost 8081
```

### 性能监控
- **连接数**: 支持多客户端并发
- **响应时间**: 通常 < 100ms
- **内存使用**: 最小化额外开销

## 🛡️ 安全考虑

### 网络安全
- **默认绑定**: `0.0.0.0:8081` (所有接口)
- **建议**: 生产环境使用防火墙限制访问
- **协议**: 目前为明文传输，适合内网使用

### 权限控制
- **服务权限**: 继承主应用权限级别
- **用户操作**: 遵循现有的用户系统规则
- **管理功能**: 需要管理员权限的功能会进行检查

## 📝 开发指南

### 添加新工具
1. 在 `filecodebox.go` 中定义工具
2. 实现处理函数
3. 注册到工具列表

```go
// 注册新工具
f.AddTool(Tool{
    Name:        "my_tool",
    Description: "我的自定义工具",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param": map[string]interface{}{
                "type": "string",
                "description": "参数说明",
            },
        },
    },
}, f.handleMyTool)
```

### 添加新资源
```go
// 注册新资源
f.AddResource(Resource{
    URI:         "my://resource",
    Name:        "我的资源",
    Description: "资源说明",
    MimeType:    "application/json",
}, f.readMyResource)
```

## 🤝 贡献

### 报告问题
- 在 GitHub Issues 中报告 MCP 相关问题
- 提供详细的错误信息和重现步骤
- 包含客户端和服务器的日志

### 提交改进
- Fork 项目并创建功能分支
- 确保新功能有适当的测试
- 更新相关文档

## 📚 相关文档

- [完整使用指南](docs/mcp-server-guide.md)
- [MCP 协议规范](https://modelcontextprotocol.io/)
- [FileCodeBox API 文档](docs/API-README.md)

## 🆘 故障排除

### 常见问题

**Q: MCP 服务器无法启动**
A: 检查环境变量 `ENABLE_MCP_SERVER=true` 是否设置，确认端口未被占用

**Q: 客户端连接被拒绝**  
A: 验证防火墙设置，确认服务器正在监听指定端口

**Q: 工具调用失败**
A: 检查参数格式，查看服务器日志获取详细错误信息

**Q: 性能问题**
A: 监控系统资源，考虑限制并发连接数

### 获取帮助
- 查看日志: `./filecodebox-mcp 2>&1 | grep "ERROR"`
- 测试连接: `python3 test_mcp_client.py`
- 联系支持: 在 GitHub 项目中提交 Issue

---

🎉 **恭喜！FileCodeBox 现在支持 MCP 协议，可以与各种 AI 助手无缝集成！**

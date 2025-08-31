# FileCodeBox MCP 管理界面使用指南

## 概述

FileCodeBox 现在支持通过管理界面配置和控制 MCP (Model Context Protocol) 服务器，无需再通过环境变量控制。

## 🎯 新功能特性

### 管理界面控制
- ✅ **Web界面配置**: 通过管理页面直接配置MCP服务器
- ✅ **实时控制**: 启动、停止、重启MCP服务器
- ✅ **状态监控**: 实时查看MCP服务器运行状态
- ✅ **动态配置**: 修改配置后立即生效

### 配置选项
- **启用/禁用**: 控制MCP服务器是否启用
- **端口配置**: 自定义MCP服务器监听端口
- **主机绑定**: 配置MCP服务器绑定的IP地址

## 🚀 使用方法

### 1. 启动应用

```bash
# 正常启动应用，不需要环境变量
./filecodebox
```

### 2. 访问管理界面

```
http://localhost:12345/admin/
```

### 3. 管理员登录

使用默认密码或配置的管理员密码登录。

### 4. MCP 配置管理

#### 4.1 查看MCP配置

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:12345/admin/mcp/config
```

响应示例：
```json
{
  "code": 200,
  "msg": "success",
  "detail": {
    "enable_mcp_server": 0,
    "mcp_port": "8081",
    "mcp_host": "0.0.0.0"
  }
}
```

#### 4.2 更新MCP配置

```bash
curl -X PUT \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -d '{
       "enable_mcp_server": 1,
       "mcp_port": "8081",
       "mcp_host": "0.0.0.0"
     }' \
     http://localhost:12345/admin/mcp/config
```

#### 4.3 查看MCP状态

```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:12345/admin/mcp/status
```

响应示例：
```json
{
  "code": 200,
  "msg": "success",
  "detail": {
    "running": true,
    "timestamp": "2025-08-31 15:30:00",
    "server_info": {
      "name": "FileCodeBox MCP Server",
      "version": "1.0.0"
    },
    "config": {
      "enabled": true,
      "port": "8081",
      "host": "0.0.0.0"
    }
  }
}
```

#### 4.4 控制MCP服务器

**启动服务器：**
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -d '{"action": "start"}' \
     http://localhost:12345/admin/mcp/control
```

**停止服务器：**
```bash
curl -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     -d '{"action": "stop"}' \
     http://localhost:12345/admin/mcp/control
```

**重启服务器：**
```bash
curl -X POST \
     -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:12345/admin/mcp/restart
```

## 📝 API 接口文档

### MCP 配置管理

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 获取MCP配置 | GET | `/admin/mcp/config` | 获取当前MCP配置 |
| 更新MCP配置 | PUT | `/admin/mcp/config` | 更新MCP配置并自动应用 |
| 获取MCP状态 | GET | `/admin/mcp/status` | 获取MCP服务器运行状态 |
| 控制MCP服务器 | POST | `/admin/mcp/control` | 启动或停止MCP服务器 |
| 重启MCP服务器 | POST | `/admin/mcp/restart` | 重启MCP服务器 |

### 请求参数

#### 更新MCP配置
```json
{
  "enable_mcp_server": 1,    // 0-禁用, 1-启用
  "mcp_port": "8081",        // 端口号（字符串）
  "mcp_host": "0.0.0.0"      // 绑定地址
}
```

#### 控制MCP服务器
```json
{
  "action": "start"  // "start" 或 "stop"
}
```

## 🔧 管理界面集成

### Web界面功能

1. **配置面板**
   - MCP服务器启用/禁用开关
   - 端口号输入框
   - 主机地址输入框
   - 保存配置按钮

2. **状态面板**
   - 服务器运行状态指示器
   - 实时状态更新
   - 最后更新时间

3. **控制面板**
   - 启动服务器按钮
   - 停止服务器按钮
   - 重启服务器按钮

### 前端实现示例

```javascript
// 获取MCP配置
async function getMCPConfig() {
  const response = await fetch('/admin/mcp/config', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return response.json();
}

// 更新MCP配置
async function updateMCPConfig(config) {
  const response = await fetch('/admin/mcp/config', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify(config)
  });
  return response.json();
}

// 控制MCP服务器
async function controlMCPServer(action) {
  const response = await fetch('/admin/mcp/control', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({ action })
  });
  return response.json();
}

// 获取MCP状态
async function getMCPStatus() {
  const response = await fetch('/admin/mcp/status', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  return response.json();
}
```

## 🔍 工作流程

### 配置更新流程
1. 用户在管理界面修改MCP配置
2. 前端调用 PUT `/admin/mcp/config` 接口
3. 后端更新数据库配置
4. 后端重新加载配置到内存
5. MCP管理器自动应用新配置
6. 如果启用状态改变，自动启动或停止服务器

### 状态监控流程
1. 前端定时调用 GET `/admin/mcp/status` 接口
2. 后端返回实时状态信息
3. 前端更新状态显示
4. 用户可以看到服务器的实时运行状态

## ⚠️ 注意事项

### 安全考虑
- 所有MCP管理接口都需要管理员权限
- JWT token必须有效
- 配置更改会立即生效，请谨慎操作

### 性能影响
- 频繁的启停操作可能影响性能
- 建议在业务空闲时进行配置更改
- 状态查询是轻量级操作，可以频繁调用

### 故障排除
- 如果MCP服务器启动失败，检查端口是否被占用
- 如果配置更新失败，检查参数格式是否正确
- 如果状态显示异常，尝试重启MCP服务器

## 🎉 迁移指南

### 从环境变量迁移

**旧方式（环境变量）：**
```bash
export ENABLE_MCP_SERVER=true
export MCP_PORT=8081
./filecodebox
```

**新方式（管理界面）：**
1. 启动应用：`./filecodebox`
2. 登录管理界面
3. 在MCP配置面板中：
   - 启用MCP服务器：开启
   - 设置端口：8081
   - 点击保存配置

### 配置持久化
- 所有配置都保存在数据库中
- 重启应用后配置自动生效
- 无需再设置环境变量

---

🎊 **恭喜！MCP服务器现在可以通过管理界面轻松配置和控制！** 🎊

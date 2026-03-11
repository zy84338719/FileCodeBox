# cmd/server/ - 服务入口

此目录存放服务启动相关代码。

## 目录结构

```
server/
├── main.go         # 程序入口
└── bootstrap/      # 初始化代码
```

## 文件说明

### main.go
- 程序入口点
- 调用 bootstrap 初始化
- 信号处理和优雅关闭

### bootstrap/
- 组件初始化（配置、日志、数据库等）
- 服务器创建和配置
- 资源清理

## 启动流程

1. 加载配置文件
2. 初始化日志
3. 初始化数据库连接
4. 初始化 Redis 连接
5. 创建 HTTP 服务器
6. 注册中间件和路由
7. 启动服务监听

## 运行方式

```bash
# 直接运行
go run cmd/server/main.go

# 编译后运行
go build -o server cmd/server/main.go
./server

# 指定配置文件
CONFIG_PATH=configs/config.prod.yaml ./server
```

# scripts/ - 脚本目录

此目录存放各类脚本文件。

## 现有脚本

### gen.sh - 代码生成脚本

统一的代码生成入口，支持 Hz HTTP 和 Kitex RPC。

```bash
# Hz HTTP 代码生成
./scripts/gen.sh hz-new idl/common.proto      # 初始化
./scripts/gen.sh hz-update idl/common.proto   # 更新

# Kitex RPC 代码生成
./scripts/gen.sh kitex idl/rpc/user.proto
```

**功能特性：**
- 自动检测工具是否安装
- 防止生成到旧的 biz/ 目录
- 自动执行 go mod tidy 和 go fmt

## 通过 Makefile 使用（推荐）

```bash
# HTTP 代码生成
make gen-http-new IDL=common.proto
make gen-http-update IDL=common.proto

# RPC 代码生成
make gen-rpc IDL=rpc/user.proto
make gen-rpc-all

# 安装工具
make tools-install
```

## 目录结构建议

```
scripts/
├── gen.sh          # 代码生成
├── bootstrap.sh    # 环境初始化
├── build/          # 构建相关
├── deploy/         # 部署相关
└── migration/      # 数据库迁移
```

## 注意

- 脚本应添加执行权限：`chmod +x script.sh`
- 使用相对路径时注意工作目录
- 敏感信息通过环境变量传递

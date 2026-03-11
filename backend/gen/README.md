# gen/ - 自动生成代码目录

此目录仅存放**自动生成**的代码，**禁止手动编写或修改**。

## 目录结构

```
gen/
├── http/           # Hz 生成的 HTTP 相关代码
│   ├── handler/    # 请求处理器骨架
│   ├── router/     # 路由注册代码
│   └── model/      # 请求/响应模型
└── rpc/            # Kitex 生成的 RPC 相关代码（如需要）
```

## 规则

- 所有代码由工具（hz、protoc、kitex）自动生成
- 不要手动编辑此目录下的文件
- 重新生成时会覆盖现有文件
- 业务逻辑应放在 `internal/app/` 中
- 数据访问层应放在 `internal/repo/` 中

## 相关命令

```bash
# 生成 HTTP 代码
make hz-new
make hz-update

# 生成 RPC 代码（如需要）
kitex -module github.com/zy84338719/fileCodeBox idl/xxx.thrift
```

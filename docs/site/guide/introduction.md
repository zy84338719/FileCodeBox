# 什么是 FileCodeBox

FileCodeBox（Go Edition）是一套可自托管的“文件快递柜”平台，目标是在内网或私有环境中提供 **安全、可控、体验良好** 的临时分享能力。

## 为什么选择 Go 版本？

- **性能与稳定性**：Go 的并发模型让大文件上传、分片合并更加顺滑。
- **可运维性**：单一二进制部署、内置健康检查、CLI 运维工具，降低上线成本。
- **现代体验**：2025 主题重新设计 Dashboard/Admin，新增 ARIA 语义、响应式布局、深色主题预留能力。
- **扩展生态**：提供统一的存储抽象层、配置管理中心、MCP（Model Context Protocol）工具集成。

## 核心场景

| 场景 | 描述 |
| --- | --- |
| 团队临时文件柜 | 共享设计稿、测试包、临时日志，支持自动过期、访问审计 |
| 远程协作 | 通过提取码和次数限制保证跨团队传输安全 |
| 自动化管道 | 利用 REST API / CLI 在 CI/CD 流程输出构建产物 |
| 一次性分享 | 分享文本、口令、密钥等敏感信息，下载一次自动失效 |

## 版本演进速览

- **v1.9.9（当前）**：
  - Dashboard/Admin UI 全面焕新，统一设计语言和表单交互
  - 文档体系重构，新增操作指南 & API 章节
  - 忘记密码流程改造，加入三步式进度与校验反馈
  - 管理后台静态资源访问策略升级，强化权限隔离
- **v1.9.x（历史）**：
  - ConfigManager 支持数据库动态配置与热更新
  - 多存储后端（local/S3/WebDAV/OneDrive）抽象与健康检查
  - Chunk Service 支持分片/秒传/断点续传
  - MCP Server 接入，服务内联调更便捷

> 查看更详细的架构与变更可参考仓库 `docs/` 目录下的专题文章，例如《PROJECT_STRUCTURE.md》、《CONFIG_REFACTOR_SUMMARY.md》。

## 架构概览

```
Client (Web UI / Admin / API / CLI)
    ↓
Routes (Gin)
    ↓
Handlers → Services → Repository → Database/Storage
    ↓
Storage Manager (local · s3 · webdav · onedrive)
```

- **Routes/Handlers**：负责 HTTP 协议解析、参数校验与错误处理。
- **Services**：实现核心业务流程，例如分享、分片上传、用户系统。
- **Repository**：对数据库访问进行封装，提供事务与统一错误处理。
- **Storage Manager**：抽象不同后端存储，实现可插拔扩展。
- **Config Manager**：集中管理配置，支持文件、环境变量、数据库三层覆盖。

接下来，建议前往 [快速开始](./getting-started.md) 体验从部署到使用的完整流程。

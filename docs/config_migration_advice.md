# FileCodeBox 配置协议重构迁移建议

1. config.yaml 迁移：
   - 所有平铺字段（如 notify_title、opacity、themes_select 等）全部归入对应功能模块（如 ui、base、transfer、user 等）。
   - 结构调整后，所有配置项都在一级模块下，便于维护和热重载。

2. Go struct 迁移：
   - ConfigManager 及所有子 struct 按 config.yaml 完全分层，字段命名、类型、json/yaml tag 保持一致。
   - AdminConfigRequest/Response 直接复用 ConfigManager，无需再单独定义冗余字段。

3. handler 层迁移：
   - /admin/config 响应直接返回 ConfigManager 对象，无需 hack、无需字段展开。
   - 更新配置时，直接反序列化为 ConfigManager 结构体。

4. 前端迁移：
   - 只需按模块对象解析（如 config.base.name、config.ui.notify_title），无需兼容处理。
   - 配置表单、展示、保存等全部按分层结构处理。

5. 兼容建议：
   - 迁移期间可保留旧字段一段时间，前后端同步切换。
   - 配置热重载、持久化、校验等逻辑建议全部基于新版分层结构实现。

6. 未来扩展：
   - 新增模块/字段时只需在 config.yaml、ConfigManager、前端 schema 同步添加即可。
   - 支持自动生成配置文档、前端表单 schema、API 文档等。

如需自动迁移脚本或批量转换工具，可进一步定制！

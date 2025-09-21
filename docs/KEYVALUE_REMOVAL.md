# KeyValue 表移除说明

背景
----
FileCodeBox 项目曾使用 `key_values` 表以键值对形式保存运行时以及启动时配置。为简化配置与提高可维护性，项目已切换到 YAML-first 配置（`config.yaml`），并移除对 `key_values` 表的运行时代码依赖。

重要变更
----
- `ConfigManager` 现在以 `config.yaml` 为主：Env > YAML > Defaults。配置以完整结构化的方式保存在 `config.yaml`。
- 运行时任意键值存储（`KeyValues`、`GetRuntimeKeyValue`、`SetRuntimeKeyValue` 等）已从代码中删除。请使用 `ConfigManager` 的结构化字段来保存配置信息。
- `sys_start` 已作为显式字段 `ConfigManager.SysStart` 持久化到 `config.yaml`（如果需要）。

迁移旧数据库数据
----
若你有旧的数据库（`data/filecodebox.db` 或其它），并希望将 `key_values` 表里的内容迁移到 `config.yaml`，我们提供了遗留脚本：

- `scripts/legacy/export_config_from_db.py`（Python，依赖 `pyyaml`）
- `scripts/legacy/export_config_from_db.go`（Go，可编译可运行）

示例（Python）：
```
python3 scripts/legacy/export_config_from_db.py data/filecodebox.db config.generated.yaml
```

生成的 `config.generated.yaml` 包含一个 `ui:` 块和基本的 `base`, `database`, `storage`, `user`, `mcp` 部分。请手动审阅并整合到你的 `config.yaml`，然后重启服务。

注意
----
- 这些脚本为迁移工具，仅作历史用途保留，并不建议在新部署中使用。
- 一旦确认迁移成功，可安全删除旧数据库中的 `key_values` 表（请在删除前备份数据库）。

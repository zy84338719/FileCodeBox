PR: Remove InitWithDB and stop using DB rows for configuration

Summary
-------
This PR removes the legacy initializer `InitWithDB()` from the configuration manager and changes the configuration source model to be YAML-first (`config.yaml`) with environment variable overrides. Database per-row configuration is no longer read or written by the application.

Key changes
-----------
- `InitWithDB()` removed from `internal/config/manager.go`.
- `SetDB(db *gorm.DB)` remains to allow injecting the DB connection if other subsystems need it, but it is not used for configuration loading.
- `InitDefaultDataInDB()`, `LoadFromDatabase()`, `saveToDatabase()`, and `Save()` no-op stubs were removed (or will be removed in follow-up commit) to avoid accidental DB-based config usage.
- Call sites using `InitWithDB()` are updated to call `SetDB(db)` instead.
- Updated docs with migration guidance and example `config.yaml` usage.

Migration guidance
------------------
1. Provide configuration via `config.yaml` at the repository root or set `CONFIG_PATH` to point to your YAML configuration file.
2. Environment variables take precedence for runtime overrides (`PORT`, `DATA_PATH`, etc.).
3. If you previously relied on DB rows for config, export them to YAML using the included script `scripts/export_config_from_db.go` and place the resulting file as `config.yaml`.

Why this change
----------------
Using `config.yaml` as the authoritative source simplifies deployment, reduces surprising runtime writes to the DB, and avoids configuration drift across instances. It also removes the complexity of managing multiple legacy DB formats.

Notes
-----
- This is a breaking change if your deployment relied on DB rows for live configuration. Ensure you export DB config and place it in `config.yaml` before upgrading.
- If you want help producing the `config.yaml` from your DB, run `go run scripts/export_config_from_db.go` or `python3 scripts/export_config_from_db.py`.

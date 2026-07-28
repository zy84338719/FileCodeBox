-- 回滚 baseline：按依赖反序删除所有表
DROP TABLE IF EXISTS notifies;
DROP TABLE IF EXISTS file_previews;
DROP TABLE IF EXISTS user_api_keys;
DROP TABLE IF EXISTS admin_operation_logs;
DROP TABLE IF EXISTS transfer_logs;
DROP TABLE IF EXISTS upload_chunks;
DROP TABLE IF EXISTS file_codes;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS schema_migrations;

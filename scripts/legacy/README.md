Deprecated scripts for KeyValue export/migration

These scripts were used to export configuration from the legacy `key_values` database
table into a YAML file. The `key_values` table has been removed from the active code
path; these scripts are kept here for historical migration purposes only and are
not used by the new YAML-first configuration system.

Files:
- export_config_from_db.py
- export_config_from_db.go

Do not run these scripts in production without understanding their behavior. They
are provided only as a convenience to migrate an existing database's key_values
into a `config.yaml` that follows the new structured format.

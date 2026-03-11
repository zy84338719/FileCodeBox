# configs/ - 配置文件目录

此目录存放配置文件（YAML/TOML/JSON），**仅存放配置文件，不含 Go 代码**。

## 文件说明

- `config.yaml` - 主配置文件

## 配置示例

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: mysql
  host: localhost
  port: 3306
  user: root
  password: ""
  db_name: mydb

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

log:
  level: info
  filename: ""
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: true

app:
  name: "My App"
  version: "1.0.0"
```

## 环境配置

建议通过环境变量覆盖敏感配置：

```bash
export CONFIG_PATH=configs/config.yaml
export DB_PASSWORD=secret
export REDIS_PASSWORD=secret
```

## 多环境配置

```
configs/
├── config.yaml         # 默认/开发配置
├── config.prod.yaml    # 生产配置
└── config.test.yaml    # 测试配置
```

## 注意

- Go 配置代码放在 `internal/conf/`
- 不要将敏感信息（真实密码）提交到版本控制

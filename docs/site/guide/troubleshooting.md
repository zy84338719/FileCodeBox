# 常见问题排查

本章节涵盖部署、上传、配置等常见问题，帮助你快速定位并恢复服务。

## 安装与启动

### 启动报错：端口已占用
- 检查是否有其他服务占用 12345 端口
- 修改 `config.yaml` 中的 `base.port`
- 或使用 `FILECODEBOX__BASE__PORT` 环境变量覆盖

### `go run` 提示缺少依赖
- 运行 `go mod tidy` 重新同步依赖
- 确保 Go 版本 ≥ 1.21
- 若在国内网络，可设置 `GOPROXY=https://goproxy.cn,direct`

### Docker 容器不断重启
- 使用 `docker logs filecodebox` 查看错误
- 确认挂载目录权限正确（容器内默认用户为 `app`）
- 若使用 SQLite，确保 `-v $(pwd)/data:/data` 已挂载

## 初始化与登录

### 无法访问 `/setup`
- 确认服务正在运行，访问日志是否返回 200
- 若已初始化，可删除 `data/init.lock`（谨慎操作）后重新启动

### 管理员密码忘记
- 在服务目录执行 `./filecodebox admin reset-password --username admin`
- 或直接操作数据库：`UPDATE users SET password='...' WHERE username='admin';`

## 上传与下载

### 上传 413 或 413 Request Entity Too Large
- 检查反向代理（Nginx）是否限制 `client_max_body_size`
- 调整 `transfer.upload.upload_size`

### 上传进度卡住
- 查看前端控制台是否有网络错误
- 检查对象存储 / WebDAV 的网络延迟
- 调整 `transfer.upload.chunk_size`，避免分片过小导致请求过多

### 下载失败 403
- 可能已超过下载次数或过期
- 若开启登录下载，确认当前用户已登录

## 存储

### S3 上传失败
- 校验 Endpoint、Region 是否与 Bucket 匹配
- 使用 `scripts/test_nfs_storage.sh` 或其他工具测试连通性

### WebDAV 连接超时
- 确认 WebDAV 服务是否开启 HTTPS
- 检查证书是否被信任
- 提高 `transfer.download.download_timeout`

## 配置热更新

### 后台修改配置无效
- 查看服务日志是否存在 `config sync` 相关错误
- 检查数据库 `key_value` 表是否成功写入
- 确认没有高优先级的环境变量覆盖

## API 调用

### 返回 401 Unauthorized
- 确认 Authorization 头格式：`Bearer <token>`
- Token 是否已过期，可重新登录获取

### Swagger 文档无法访问
- 默认访问 `/swagger/index.html`
- 若被反向代理拦截，添加相应放行规则

## 前端显示异常

- 清除浏览器缓存，确保加载最新主题资源
- 确认静态资源是否被反向代理缓存
- 检查 `themes/2025` 中是否有自定义修改

如仍无法解决，可在 Issues 中附带日志、配置、复现步骤求助。

# 开发数据库接入

开发库连接串只放在本地 `.env` 或密码管理器中，不写入 Git、日志或截图。

推荐数据库服务启用 TLS，并在 `pg_hba.conf` 中仅放行开发者固定出口 IP、指定数据库和指定用户。配置变更后 reload PostgreSQL，再用迁移命令验证：

```powershell
go run ./cmd/migrate up
```

若暂时使用非 TLS 连接，必须将范围限制在开发库，并在 `pg_hba.conf` 明确放行来源地址；不要使用全网段或 `0.0.0.0/0`。生产环境必须恢复 TLS 或使用受限内网。

常见错误：

- `server does not support SSL`：服务端未启用 TLS，连接串却要求 TLS。
- `no pg_hba.conf entry`：当前来源 IP、用户、数据库或加密模式未被访问规则允许。
- 迁移成功但应用失败：检查 API 与 Worker 是否使用了同一份连接配置。

# Debian 部署 Runbook

1. 建立 `pcplatform` 系统用户以及 `/opt/pcplatform`、`/etc/pcplatform`、`/var/lib/pcplatform`。
2. Windows 执行 `scripts/build-linux.ps1`，上传三个 Linux 可执行文件。
3. 创建仅 root 可读的 `/etc/pcplatform/pcplatform.env`。生产数据库必须使用 TLS 或受限内网连接。
4. 执行 `pcp-migrate up`，确认成功后替换 API/Worker。
5. 安装 `deploy/systemd` 下的 unit，执行 daemon-reload、enable 和 restart。
6. 检查 `/healthz`、`/readyz`、Worker 日志、发布队列和公开页面。

回滚应用时保留兼容迁移；数据库 schema 按 expand/backfill/switch/contract 管理，不直接回滚破坏性迁移。

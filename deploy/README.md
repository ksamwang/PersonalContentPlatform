# 部署

生产目标为 Debian Linux，保留两种方式：

- Docker Compose：在服务器执行 `docker compose -f deploy/compose/compose.yml up -d --build`。
- Windows 交叉编译：执行 `scripts/build-linux.ps1`，将 `bin/linux-amd64` 上传到 Debian，配合 systemd 与外部 PostgreSQL 运行。

生产环境必须使用独立强密码、TLS 数据库连接及 Secret 文件，并限制数据库来源地址。

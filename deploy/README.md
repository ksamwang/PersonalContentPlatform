# 部署

生产目标为 Debian Linux，保留两种方式：

- Docker Compose：在服务器执行 `docker compose -f deploy/compose/compose.yml up -d --build`。
- Windows 交叉编译：执行 `scripts/build-linux.ps1`，将 `bin/linux-amd64` 上传到 Debian，配合 systemd 与外部 PostgreSQL 运行。

生产环境必须使用独立强密码、TLS 数据库连接及 Secret 文件，并限制数据库来源地址。

Compose 启动前创建以下未纳入 Git 的文件：

- `deploy/secrets/db_password.txt`：仅包含容器 PostgreSQL 用户密码。
- `deploy/secrets/database_url.txt`：包含 API/Worker 使用的完整连接串，容器内数据库主机名为 `postgres`，生产连接应启用 TLS。
- 根目录 `.env`：应用非敏感配置；生产密钥也可以由受控的宿主环境或 Secret 管理器注入。

对象存储配置：

- 本地文件：`OBJECT_STORAGE_PROVIDER=filesystem`。
- 通用 S3：`OBJECT_STORAGE_PROVIDER=s3`。
- Cloudflare R2：`OBJECT_STORAGE_PROVIDER=r2`，同时配置 endpoint、bucket 和密钥。
- 阿里云 OSS：`OBJECT_STORAGE_PROVIDER=oss`，endpoint 填 OSS 地域地址，region、bucket 和密钥按对应账号配置。

Webhook 密钥不进入数据库。服务端环境中配置密钥变量，端点的 `secret_ref` 只记录变量名。

非容器部署使用 `deploy/systemd` 中的服务单元。上线、备份恢复和故障处理详见 `docs/runbooks`。

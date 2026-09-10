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

Webhook 密钥按项目约定明文保存在数据库中，备份和数据库运维权限应按包含密钥的数据管理。API 不返回密钥正文，旧 `secret_ref` 配置仍兼容。

公开站缓存默认保存 5 分钟。API 与 Worker 使用 `PUBLIC_WEB_ORIGIN` 通知公开站即时失效缓存；Docker Compose 已指向 `http://public-web:3000`。生产环境应在根目录 `.env` 为 API、Worker和 Public Web 设置相同的 `PUBLIC_REVALIDATE_TOKEN`。

Worker 同时负责发布、全文及向量索引、Inbox 智能处理、AI 建议和 Webhook 投递，生产环境不要只启动 API。

非容器部署使用 `deploy/systemd` 中的服务单元。上线、备份恢复和故障处理详见 `docs/runbooks`。

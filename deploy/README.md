# 宝塔非 Docker 部署

生产目标为 Debian Linux，当前正式部署路径不使用 Docker：

- API、Worker、迁移器：Windows 交叉编译为 Linux `amd64` 二进制，API 和 Worker 由宝塔分别守护。
- Studio、Public Web：在 Debian 上构建 Next.js standalone，由宝塔 Node 项目管理器运行。
- 域名、HTTPS、反向代理：由宝塔 Nginx 管理。
- PostgreSQL：使用服务器上已有实例，不随应用打包。

在 Windows 生成完整部署包：

```powershell
.\deploy\build.ps1
```

产物位于 `deploy/output/pcplatform-<commit>.zip`。解压后已经是可整体上传到 `/www/wwwroot/pcplatform` 的目录，包含：

- Studio 与 Public Web 的完整构建源码；
- 根目录 `package.json` 和 `package-lock.json`；
- 三个 Linux 可执行文件及 SHA-256 校验和；
- 已按 `loshin.org` 填写的三个生产配置文件；
- 前端一键构建脚本 `build-web.sh`；
- `data/assets` 目录。

部署包自动排除 Windows 的 `.next`、`node_modules`、`*.tsbuildinfo` 和开发说明文件。上传前只需要填写 `config/pcplatform.env` 中的数据库密码、`PASSWORD_PEPPER`、`PUBLIC_REVALIDATE_TOKEN`，并把同一个 Token 写入 `apps/public-web/.env.production`。更新已有服务器时保留服务器上的三个配置文件，不要被模板覆盖。

在 Linux 上传并准备二进制：

```bash
cd /www/wwwroot/pcplatform
chmod 755 bin/pcp-api bin/pcp-worker bin/pcp-migrate
chmod 640 config/pcplatform.env apps/studio-web/.env.production apps/public-web/.env.production
APP_ENV_FILE=/www/wwwroot/pcplatform/config/pcplatform.env ./bin/pcp-migrate up
```

在宝塔中新建两个独立的 Go 项目或进程守护项：

| 项目 | 工作目录 | 启动文件 | 环境变量 |
| --- | --- | --- | --- |
| API | `/www/wwwroot/pcplatform` | `./bin/pcp-api` | `APP_ENV_FILE=/www/wwwroot/pcplatform/config/pcplatform.env` |
| Worker | `/www/wwwroot/pcplatform` | `./bin/pcp-worker` | `APP_ENV_FILE=/www/wwwroot/pcplatform/config/pcplatform.env` |

两者都设置开机启动和异常自动重启。API 监听 `127.0.0.1:8080`，Worker 不配置公网端口。

前端源码已经包含在完整部署包中。前端必须在 Linux 构建，避免 Windows 的 Next.js/Sharp 原生依赖进入 Linux 产物：

```bash
bash build-web.sh
```

在宝塔中建立两个 Node 项目，共用项目目录，分别使用对应环境变量示例中的配置：

| 项目 | 启动命令 | 端口 |
| --- | --- | --- |
| Studio | `node apps/studio-web/.next/standalone/apps/studio-web/server.js` | 3001 |
| Public Web | `node apps/public-web/.next/standalone/apps/public-web/server.js` | 3000 |

宝塔创建两个网站并反向代理到对应本机端口。Studio 域名必须与 `WEBAUTHN_RP_ID`、`WEBAUTHN_RP_ORIGINS` 一致；API 仅监听 `127.0.0.1:8080`，不需要单独暴露公网端口。

生产数据库使用独立密码；远程连接应启用 TLS，同机或受限内网连接按服务器实际网络配置。根目录 `.env` 只用于本地开发，生产服务读取 `/www/wwwroot/pcplatform/config/pcplatform.env` 和宝塔 Node 项目环境变量。

对象存储 Provider 在 Studio 的 Workspace 设置中管理，可选 Filesystem、S3、Cloudflare R2 和阿里云 OSS。部署环境只保留 `OBJECT_STORAGE_BASE_PATH`，用于限制 Filesystem Profile 可访问的服务器目录。

Webhook 密钥按项目约定明文保存在数据库中，备份和数据库运维权限应按包含密钥的数据管理。API 不返回密钥正文，旧 `secret_ref` 配置仍兼容。

公开站缓存默认保存 5 分钟。API 与 Worker 使用 `PUBLIC_WEB_ORIGIN=http://127.0.0.1:3000` 通知公开站即时失效缓存；API、Worker 和 Public Web 必须使用相同的 `PUBLIC_REVALIDATE_TOKEN`。

Worker 同时负责发布、全文及向量索引、Inbox 智能处理、AI 建议和 Webhook 投递，生产环境不要只启动 API。

完整上线顺序、备份恢复和故障处理详见 `docs/runbooks`。`deploy/compose` 与根目录 Dockerfile 仅作为历史备用配置，不属于当前生产部署流程。

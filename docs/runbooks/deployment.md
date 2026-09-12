# 宝塔 Debian 部署 Runbook

1. 在 Windows 执行 `scripts/package-baota.ps1`，得到包含三个 Linux `amd64` 可执行文件的 ZIP 发布包。
2. 在服务器建立 `/www/wwwroot/pcplatform/bin`、`/www/wwwroot/pcplatform/config`、`/www/wwwroot/pcplatform/data/assets`，将发布包中的 `bin` 上传并增加执行权限。
3. 参考 `deploy/baota/pcplatform.env.example` 创建 `/www/wwwroot/pcplatform/config/pcplatform.env`。远程数据库使用 TLS；数据库与应用在同一受限服务器网络时可按实际情况配置。
4. 首次上线或包含迁移时，在宝塔终端执行 `APP_ENV_FILE=/www/wwwroot/pcplatform/config/pcplatform.env ./bin/pcp-migrate up`，确认成功后再替换 API 和 Worker。
5. 在宝塔中为 `pcp-api` 和 `pcp-worker` 建立两个独立的 Go 项目或进程守护项，工作目录使用 `/www/wwwroot/pcplatform`，并为两者设置同一个 `APP_ENV_FILE`。API 监听 `127.0.0.1:8080`，Worker 不开放端口。
6. 将 `deploy/baota/studio.env.example`、`public-web.env.example` 分别复制为两个应用的 `.env.production` 并填写实际值，然后在服务器项目目录执行 `npm ci && npm run build:web`。再把每个应用的 `.next/static` 复制到对应 standalone 目录。Next.js 包含平台相关依赖，因此不要将在 Windows 生成的 `.next` 或 `node_modules` 上传到 Linux。
7. 在宝塔添加两个 Node 项目，分别执行 `node apps/studio-web/.next/standalone/apps/studio-web/server.js` 和 `node apps/public-web/.next/standalone/apps/public-web/server.js`。启动前按两个环境变量示例配置项目变量，其中 Studio 使用 3001，Public Web 使用 3000。
8. 在宝塔创建公开站和 Studio 两个网站，分别反向代理到 `127.0.0.1:3000` 与 `127.0.0.1:3001`，申请 HTTPS 证书；API 保持只监听 `127.0.0.1:8080`。
9. 检查 API `/healthz`、`/readyz`、两个网页、登录、图片读取、发布缓存刷新以及 Worker 日志。

回滚应用时保留兼容迁移；数据库 schema 按 expand/backfill/switch/contract 管理，不直接回滚破坏性迁移。

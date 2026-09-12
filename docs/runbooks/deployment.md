# 宝塔 Debian 部署 Runbook

1. 在 Windows 执行 `deploy/build.ps1`，得到 `deploy/output/pcplatform-<commit>.zip` 完整部署包。
2. 将 ZIP 中的版本目录整体上传到 `/www/wwwroot/pcplatform`。包内已经包含前端源码、根包文件、三个 Linux `amd64` 可执行文件、配置文件和数据目录。
3. 填写 `config/pcplatform.env` 中的数据库密码、Pepper 和刷新 Token；将同一个刷新 Token 写入 `apps/public-web/.env.production`。远程数据库使用 TLS；数据库与应用在同一受限服务器网络时可按实际情况配置。
4. 首次上线或包含迁移时，在宝塔终端执行 `APP_ENV_FILE=/www/wwwroot/pcplatform/config/pcplatform.env ./bin/pcp-migrate up`，确认成功后再替换 API 和 Worker。
5. 在宝塔中为 `pcp-api` 和 `pcp-worker` 建立两个独立的 Go 项目或进程守护项，工作目录使用 `/www/wwwroot/pcplatform`，并为两者设置同一个 `APP_ENV_FILE`。API 监听 `127.0.0.1:8080`，Worker 不开放端口。
6. 在服务器项目目录执行 `bash build-web.sh`。脚本会安装依赖、构建两个前端并整理 standalone 静态资源。Next.js 包含平台相关依赖，因此部署包不会携带 Windows 生成的 `.next` 或 `node_modules`。
7. 在宝塔添加两个 Node 项目，分别执行 `node apps/studio-web/.next/standalone/apps/studio-web/server.js` 和 `node apps/public-web/.next/standalone/apps/public-web/server.js`。启动前按两个环境变量示例配置项目变量，其中 Studio 使用 3001，Public Web 使用 3000。
8. 在宝塔创建公开站和 Studio 两个网站，分别反向代理到 `127.0.0.1:3000` 与 `127.0.0.1:3001`，申请 HTTPS 证书；API 保持只监听 `127.0.0.1:8080`。
9. 检查 API `/healthz`、`/readyz`、两个网页、登录、图片读取、发布缓存刷新以及 Worker 日志。

回滚应用时保留兼容迁移；数据库 schema 按 expand/backfill/switch/contract 管理，不直接回滚破坏性迁移。

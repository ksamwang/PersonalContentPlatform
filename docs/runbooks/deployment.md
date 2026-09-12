# 宝塔 Docker 部署 Runbook

1. Windows 执行 `deploy/build.ps1`，生成 `deploy/output/pcplatform-docker-<commit>.zip`。
2. 将完整版本目录上传为 `/www/wwwroot/pcplatform`，填写根目录 `.env` 中的 PostgreSQL 密码、Pepper 和刷新 Token。
3. 执行 `docker compose --env-file .env -f deploy/compose/compose.yml config` 检查最终配置。
4. 执行 `docker compose --env-file .env -f deploy/compose/compose.yml up -d --build`。迁移器成功退出后 API 和 Worker 才会启动。
5. 宝塔将 `loshin.org` 反向代理到 `127.0.0.1:23000`，将 `studio.loshin.org` 反向代理到 `127.0.0.1:23001`，并为两个站点启用 HTTPS。
6. 检查 Compose 服务状态、迁移/API/Worker 日志、API `/healthz`、两个网页、登录、图片读取和发布缓存刷新。

更新应用时保留服务器 `.env` 和 Docker volumes，上传新版后再次执行 `up -d --build`。不要使用 `docker compose down -v`。

若要迁移已有外部 PostgreSQL 数据，先从原数据库导出，再恢复到 Compose PostgreSQL；不要同时让新旧数据库接受写入。

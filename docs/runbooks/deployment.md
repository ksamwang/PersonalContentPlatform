# 生产发布运维 Runbook

完整的首次部署、宝塔、Runner、`.env` 和手动备用部署说明见 [`deploy/README.md`](../../deploy/README.md)。本文只保留日常发布和故障处理速查。

## 正常发布

推送或合并到 `main` 后，在 GitHub Actions 中确认以下任务全部成功：

```text
(go + web) -> images -> deploy
```

实际部署入口是 `deploy/scripts/deploy.sh <full-git-sha>`。它会防止并发发布、备份数据库、拉取同一提交的三张镜像、执行迁移、更新容器并检查健康状态。健康检查失败时恢复上一应用镜像，但不会回滚数据库迁移。

## 发布后检查

```bash
cat /home/actions-runner/pcplatform-deploy/current-version
ls -lt /home/actions-runner/pcplatform-deploy/backups | head

docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}' \
  --filter name=pcplatform

curl -fsS http://127.0.0.1:23001/api/healthz
curl -I https://loshin.org/
curl -I https://studio.loshin.org/
curl -fsS https://studio.loshin.org/api/healthz
```

`current-version` 应等于本次 Actions 运行的完整提交 SHA。随后人工验证登录、图片读取、内容编辑、发布和公开页刷新。

## Runner 离线

```bash
cd /home/actions-runner
./svc.sh status
journalctl -u 'actions.runner.ksamwang-PersonalContentPlatform.pcplatform-production.service' \
  -n 100 --no-pager
./svc.sh start
```

GitHub Runners 页面应显示 `pcplatform-production` 为 `Online`，并具有 `self-hosted`、`Linux`、`X64` 和 `pcplatform-production` 标签。

## 查看生产日志

```bash
cd /home/actions-runner/_work/PersonalContentPlatform/PersonalContentPlatform
export PCP_ENV_FILE=/www/wwwroot/pcplatform/.env

docker compose --env-file "$PCP_ENV_FILE" \
  -f deploy/compose/compose.prod.yml ps
docker compose --env-file "$PCP_ENV_FILE" \
  -f deploy/compose/compose.prod.yml logs --tail=200 \
  postgres migrate api worker public-web studio-web
```

## 人工恢复上一应用版本

先确认目标完整 SHA 的三张 GHCR 镜像存在：

```bash
cd /home/actions-runner/_work/PersonalContentPlatform/PersonalContentPlatform
export PCP_ENV_FILE=/www/wwwroot/pcplatform/.env
export APP_VERSION=<previous-full-git-sha>

docker compose --env-file "$PCP_ENV_FILE" \
  -f deploy/compose/compose.prod.yml pull api worker public-web studio-web
docker compose --env-file "$PCP_ENV_FILE" \
  -f deploy/compose/compose.prod.yml up -d --no-build \
  api worker public-web studio-web
```

应用镜像回滚不会撤销数据库迁移。旧应用若不兼容新数据库，应停止反复重启并按已验证备份恢复。

## 关键路径

| 内容 | 路径 |
| --- | --- |
| 生产配置 | `/www/wwwroot/pcplatform/.env` |
| 当前成功版本 | `/home/actions-runner/pcplatform-deploy/current-version` |
| 发布前数据库备份 | `/home/actions-runner/pcplatform-deploy/backups` |
| 发布锁 | `/home/actions-runner/pcplatform-deploy/deploy.lock` |
| Runner | `/home/actions-runner` |

## 禁止操作

- 不要提交或覆盖服务器生产 `.env`。
- 不要执行 `docker compose down -v`，它会删除 PostgreSQL 和本地资产卷。
- 不要只更新 API 而长期停用 Worker。
- 不要把推送成功或镜像构建成功误认为生产部署成功，必须确认 `deploy` 任务和线上健康检查。

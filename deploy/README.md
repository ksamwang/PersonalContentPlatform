# Personal Content Platform 生产部署

本文档是当前生产环境的完整部署说明，适用于 Debian、宝塔面板和 Docker Compose。日常发布使用 GitHub Actions 与生产 Self-hosted Runner；Windows 完整部署包只作为首次部署或 CI/CD 故障时的备用方案。

## 1. 当前生产约定

| 项目 | 值 |
| --- | --- |
| Public Web | `https://loshin.org` |
| Studio | `https://studio.loshin.org` |
| Public Web 本机端口 | `127.0.0.1:23000` |
| Studio 本机端口 | `127.0.0.1:23001` |
| 生产配置 | `/www/wwwroot/pcplatform/.env` |
| Runner 用户 | `actionsrunner` |
| Runner 目录 | `/home/actions-runner` |
| Runner 标签 | `pcplatform-production` |
| 发布状态与备份 | `/home/actions-runner/pcplatform-deploy` |
| 自动发布分支 | `main` |

生产环境由 Docker Compose 运行 PostgreSQL + pgvector、一次性迁移器、API、Worker、Studio 和 Public Web。宝塔只负责 Nginx、域名、证书和反向代理，不运行应用进程，也不启动 Caddy。

Node.js 24 和 Go 编译环境都在构建镜像中，生产宿主机不需要安装 Node.js、npm 或 Go，只需要 Docker、Docker Compose、curl、tar 和正常可用的 HTTPS 网络。

## 2. 服务、镜像和数据

生产编排文件是 `deploy/compose/compose.prod.yml`。

| 服务 | 作用 | 镜像或数据 |
| --- | --- | --- |
| `postgres` | PostgreSQL 16 + pgvector 0.7.4 | `pgvector/pgvector:0.7.4-pg16`，数据卷 `postgres-data` |
| `migrate` | 执行数据库向上迁移后退出 | Backend 镜像 |
| `api` | Go API 与健康检查 | Backend 镜像，挂载 `asset-data` |
| `worker` | 发布、索引、AI、Inbox 和 Webhook 任务 | Backend 镜像，挂载 `asset-data` |
| `public-web` | 公开站点 | Public Web 镜像，只绑定 `127.0.0.1:23000` |
| `studio-web` | Studio 管理端 | Studio 镜像，只绑定 `127.0.0.1:23001` |

CI/CD 发布以下三张镜像，同时写入完整 Git SHA 标签和 `main` 标签：

```text
ghcr.io/ksamwang/personal-content-platform-backend
ghcr.io/ksamwang/personal-content-platform-studio
ghcr.io/ksamwang/personal-content-platform-public
```

生产部署始终使用完整的 40 位 Git SHA，不使用浮动的 `main` 标签，从而保证三个应用镜像来自同一次提交。

## 3. 首次准备服务器

### 3.1 安装并确认 Docker

在宝塔安装 Docker 管理器，或按 Docker 官方方式安装 Docker Engine。确认：

```bash
docker version
docker compose version
curl --version
```

端口 `23000` 和 `23001` 只绑定回环地址，不需要在防火墙或安全组中开放。公网只需要开放宝塔 Nginx 使用的 `80` 和 `443`。

### 3.2 创建生产配置

```bash
mkdir -p /www/wwwroot/pcplatform
cd /www/wwwroot/pcplatform
```

以 `deploy/compose/.env.example` 为模板创建 `/www/wwwroot/pcplatform/.env`，不要把生产 `.env` 提交到 Git：

```dotenv
COMPOSE_PROJECT_NAME=pcplatform
APP_VERSION=replace-with-full-git-sha
DOCKER_REGISTRY=docker.m.daocloud.io
GOPROXY=https://goproxy.cn,direct
NPM_REGISTRY=https://registry.npmmirror.com
PUBLIC_WEB_PORT=23000
STUDIO_WEB_PORT=23001

POSTGRES_DB=pcplatform
POSTGRES_USER=pcplatform
POSTGRES_PASSWORD=replace-with-url-safe-random-password

APP_ENV=production
DB_POOL_MAX=10
SESSION_COOKIE_NAME=pcp_session
SESSION_TTL=720h
PASSWORD_PEPPER=replace-with-a-long-random-value

WEBAUTHN_RP_ID=studio.loshin.org
WEBAUTHN_RP_ORIGINS=https://studio.loshin.org

WORKSPACE_SLUG=personal
PUBLIC_REVALIDATE_TOKEN=replace-with-a-random-value
```

随机值可以在服务器生成：

```bash
openssl rand -hex 24
openssl rand -hex 32
```

配置注意事项：

- `POSTGRES_PASSWORD` 使用 URL 安全字符，避免数据库连接串解析问题。
- `PASSWORD_PEPPER` 首次生成后固定保存，随意更换会影响已有密码登录。
- AI、Embedding、OSS/S3/R2 和 Webhook 等工作区参数在 Studio 设置页面维护，不写入生产 `.env`。
- Passkey 与域名绑定。更换 Studio 域名后需要修改两个 `WEBAUTHN_*` 值并重新注册 Passkey；密码和 TOTP 仍可用于登录。

限制配置文件权限，并让 Runner 用户可读：

```bash
chown root:actionsrunner /www/wwwroot/pcplatform/.env
chmod 640 /www/wwwroot/pcplatform/.env
```

## 4. 配置宝塔域名与 HTTPS

给 `loshin.org` 和 `studio.loshin.org` 配置指向生产服务器公网 IP 的 DNS A 记录。在宝塔中新建两个网站并配置反向代理：

| 宝塔网站 | 反向代理目标 |
| --- | --- |
| `loshin.org` | `http://127.0.0.1:23000` |
| `studio.loshin.org` | `http://127.0.0.1:23001` |

两个网站都申请 SSL 证书并开启 HTTP 跳转 HTTPS。Studio 网站增加：

```nginx
client_max_body_size 100m;
proxy_read_timeout 300s;
```

不要把 API 另行暴露到公网。浏览器请求 `/api/*` 时，由 Studio 或 Public Web 容器转发到 Docker 网络中的 `api:8080`。

## 5. 配置 GitHub Actions Runner

Runner 只负责生产部署，源码检查和镜像构建由 GitHub 托管 Runner 完成，因此生产服务器性能和宿主 Node.js 版本不会限制前端构建。

### 5.1 创建普通用户

GitHub Runner 禁止以 root 交互运行。若用户尚未创建：

```bash
useradd -m -s /bin/bash actionsrunner
usermod -aG docker actionsrunner
mkdir -p /home/actions-runner
chown -R actionsrunner:actionsrunner /home/actions-runner
```

进入 GitHub 仓库：

```text
Settings -> Actions -> Runners -> New self-hosted runner -> Linux -> x64
```

GitHub 页面会给出当前版本、下载地址、校验值和短期注册令牌。注册令牌不要写入仓库或本文档。

### 5.2 下载并注册

切换到普通用户，并按 GitHub 页面当时显示的版本和校验值执行：

```bash
su - actionsrunner
cd /home/actions-runner

curl -o actions-runner-linux-x64-<version>.tar.gz -L \
  https://github.com/actions/runner/releases/download/v<version>/actions-runner-linux-x64-<version>.tar.gz
echo "<sha256>  actions-runner-linux-x64-<version>.tar.gz" | sha256sum -c
tar xzf actions-runner-linux-x64-<version>.tar.gz

./config.sh \
  --url https://github.com/ksamwang/PersonalContentPlatform \
  --token <short-lived-registration-token> \
  --name pcplatform-production \
  --labels pcplatform-production \
  --work _work \
  --unattended

docker ps
```

如果 `docker ps` 仍提示权限不足，退出当前 shell 后重新登录再验证。

### 5.3 设置 Runner 常驻

使用 Runner 自带的服务安装器保持 Runner 在线。这只管理 Runner，应用本身仍由 Docker Compose 管理，不需要编写应用 systemd unit：

```bash
exit
cd /home/actions-runner
./svc.sh install actionsrunner
./svc.sh start
./svc.sh status
```

GitHub Runners 页面应显示：

```text
Name: pcplatform-production
Status: Online
Labels: self-hosted, Linux, X64, pcplatform-production
```

Runner Online 后，下一次推送到 `main` 会完成首次自动部署。也可以在 GitHub Actions 中重新运行最新一次 `main` 工作流。首次部署前必须确保 `/www/wwwroot/pcplatform/.env` 已填写且 `actionsrunner` 可读。

Runner 常用命令：

```bash
cd /home/actions-runner
./svc.sh status
./svc.sh stop
./svc.sh start
```

## 6. 自动发布流程

工作流文件是 `.github/workflows/ci.yml`。Pull Request 只执行检查；只有推送到 `main` 才构建镜像并部署生产环境。

```text
push / pull_request
        |
        +-- Go：test -> vet -> build
        |
        +-- Web：npm ci -> typecheck -> build
                    |
main push ----------+-- 三张 linux/amd64 镜像 -> GHCR
                                              |
                                              +-- 生产 Runner 部署
```

部署任务执行 `deploy/scripts/deploy.sh <full-git-sha>`，依次完成：

1. 使用文件锁阻止两个生产发布同时运行。
2. 启动或确认 PostgreSQL 正常。
3. 使用 `pg_dump -Fc` 备份当前数据库。
4. 拉取同一 Git SHA 的三张应用镜像。
5. 执行数据库向上迁移。
6. 更新 API、Worker 和两个 Web 容器。
7. 最多等待约 90 秒并检查 Public Web 与 API。
8. 成功后记录 `current-version`；失败时恢复上一版本应用镜像。

数据库迁移不会自动降级。自动回滚只恢复应用镜像，因此不兼容数据库变更必须采用可向后兼容的分阶段迁移。

工作流使用仓库自动生成的 `GITHUB_TOKEN` 登录 GHCR，不需要额外创建 SSH 密钥或 GitHub Secret。生产 Runner 主动连接 GitHub，GitHub 不需要 SSH 登录服务器。

## 7. 日常发布

正常更新只需要提交并推送代码：

```powershell
git push origin main
```

在 GitHub Actions 页面确认四个任务全部成功：

```text
go -> success
web -> success
images -> success
deploy -> success
```

不能仅凭代码已推送或镜像已构建判断上线完成；必须确认 `deploy` 成功并检查线上地址。

## 8. 发布后验证

### 8.1 版本、备份和容器

```bash
cat /home/actions-runner/pcplatform-deploy/current-version
ls -lt /home/actions-runner/pcplatform-deploy/backups | head

docker ps --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}' \
  --filter name=pcplatform
```

`current-version` 应等于本次 Actions 的完整提交 SHA。PostgreSQL、API、Studio 和 Public Web 应为 `healthy`，Worker 应为 `Up`。`migrate` 是一次性任务，成功退出后不会长期出现在 `docker ps` 中。

### 8.2 本机与公网

```bash
curl -I http://127.0.0.1:23000/
curl -I http://127.0.0.1:23001/
curl -fsS http://127.0.0.1:23001/api/healthz

curl -I https://loshin.org/
curl -I https://studio.loshin.org/
curl -fsS https://studio.loshin.org/api/healthz
```

Public Web 根路径按语言返回重定向是正常行为。最后人工验证登录、图片读取、内容编辑、发布和公开页面刷新。

## 9. 日志与故障定位

### Runner 离线

```bash
cd /home/actions-runner
./svc.sh status
journalctl -u 'actions.runner.ksamwang-PersonalContentPlatform.pcplatform-production.service' \
  -n 100 --no-pager
./svc.sh start
```

### 容器状态与日志

```bash
cd /home/actions-runner/_work/PersonalContentPlatform/PersonalContentPlatform
export PCP_ENV_FILE=/www/wwwroot/pcplatform/.env

docker compose --env-file "$PCP_ENV_FILE" \
  -f deploy/compose/compose.prod.yml ps
docker compose --env-file "$PCP_ENV_FILE" \
  -f deploy/compose/compose.prod.yml logs --tail=200 \
  postgres migrate api worker public-web studio-web
```

常见判断：

- `go` 或 `web` 失败：代码检查未通过，生产环境未更新。
- `images` 失败：镜像未完整生成，生产环境未更新。
- `deploy` 一直等待：检查 Runner 是否 Online、标签是否包含 `pcplatform-production`。
- `deploy` 拉取失败：检查 GHCR 登录步骤、服务器到 `ghcr.io` 的网络和磁盘空间。
- 本机端口正常但公网失败：检查宝塔反向代理、DNS 和 SSL。
- API 正常但异步功能不工作：检查 Worker 日志。

GitHub Actions 可能显示第三方 Action 的 Node.js Runtime 弃用警告；警告本身不等于失败，应以任务结论和实际错误为准。

## 10. 回滚

健康检查失败时，脚本会自动恢复到 `current-version` 记录的上一版本应用镜像，数据库迁移不会回滚。

人工恢复应用镜像前，先确认目标 SHA 的三张镜像都存在，然后执行：

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

若旧应用不兼容已经升级的数据库，不要反复重启，应修复应用或按已验证的数据库备份恢复。数据库恢复说明见 [`docs/runbooks/backup-restore.md`](../docs/runbooks/backup-restore.md)。

## 11. 数据备份

每次自动部署前都会创建 PostgreSQL custom-format 备份：

```text
/home/actions-runner/pcplatform-deploy/backups/pcplatform-<UTC时间>.dump
```

发布脚本不会自动删除旧备份，应在宝塔计划任务中按磁盘空间设置保留周期，并把重要备份复制到服务器之外。Filesystem 资产位于 Docker volume `pcplatform_asset-data`，数据库备份不包含这些文件；使用 OSS/S3/R2 时也要配置对象存储侧的备份或版本控制。

数据库恢复前要停止 API 和 Worker 写入，并先在隔离环境验证备份。

## 12. Windows 手动备用部署

CI/CD 不可用时，在 Windows 生成完整部署包：

```powershell
.\deploy\build.ps1
```

产物位于 `deploy/output/pcplatform-docker-<commit>.zip`，包含 Docker 构建所需的 Go/Next.js 源码、Dockerfile、Compose、迁移文件和依赖锁文件，并排除 `.next`、`node_modules`、本地数据和开发 `.env`。

自动发布模式不需要向服务器上传源码。手动备用模式也只上传该 ZIP，无需上传：

```text
.git
.github
.codex-tmp
node_modules
apps/*/.next
data
deploy/output 中的其他历史产物
```

将 ZIP 解压到临时版本目录。部署包自带的是占位 `.env`，必须先用生产配置替换它，然后从版本目录执行：

```bash
cp /www/wwwroot/pcplatform/.env ./.env
docker pull docker.m.daocloud.io/library/alpine:3.22
docker compose --env-file .env \
  -f deploy/compose/compose.yml config
docker compose --env-file .env \
  -f deploy/compose/compose.yml up -d --build
```

手动构建也在 Docker 内使用 Node.js 24 和 Go，不依赖宿主系统版本。更新时不要执行 `docker compose down -v`，其中 `-v` 会删除 PostgreSQL 和本地资产数据卷。

## 13. 修改域名或端口

修改 Studio 域名时同步更新 DNS、宝塔网站、`.env` 中的 `WEBAUTHN_RP_ID` 与 `WEBAUTHN_RP_ORIGINS`，重新发布后使用密码或 TOTP 登录并重新注册 Passkey。

修改 Public Web 域名时，还要修改 `deploy/compose/compose.prod.yml` 和 `deploy/compose/compose.yml` 中的 `PUBLIC_SITE_ORIGIN`，提交后重新发布。

修改本机端口时调整 `.env` 的 `PUBLIC_WEB_PORT`、`STUDIO_WEB_PORT`，并同步修改宝塔反向代理目标。

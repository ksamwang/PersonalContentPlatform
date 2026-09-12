# 宝塔 + Docker Compose 部署

生产环境由 Docker Compose 运行 PostgreSQL、迁移器、API、Worker、Studio 和 Public Web。Node.js 24 与 Go 编译环境都在镜像中，宿主操作系统不需要安装 Node.js 或 Go。宝塔继续负责 Nginx、域名和 HTTPS，不运行应用进程，也不启动 Caddy。

## 1. 在 Windows 生成完整部署包

```powershell
.\deploy\build.ps1
```

产物位于 `deploy/output/pcplatform-docker-<commit>.zip`。部署包包含 Docker 构建需要的全部 Go/Next.js 源码、Dockerfile、Compose、迁移文件和根依赖锁文件，自动排除 `.next`、`node_modules`、本地数据与开发环境文件。

## 2. 上传并填写配置

将 ZIP 内的版本目录整体上传为：

```text
/www/wwwroot/pcplatform
```

只需编辑根目录的 `.env`：

- `DOCKER_REGISTRY`：默认 `docker.m.daocloud.io`，用于服务器无法访问 Docker Hub 的环境。
- `GOPROXY`：默认 `https://goproxy.cn,direct`，用于下载 Go 依赖。
- `NPM_REGISTRY`：默认 `https://registry.npmmirror.com`，用于下载前端依赖。
- `POSTGRES_PASSWORD`：使用 URL 安全的随机密码，建议 `openssl rand -hex 24`。
- `PASSWORD_PEPPER`：首次部署生成后固定保存，建议 `openssl rand -hex 32`。
- `PUBLIC_REVALIDATE_TOKEN`：建议 `openssl rand -hex 32`。
- `WORKSPACE_SLUG`：默认 `personal`，如果已有 Workspace 使用其他 slug 再修改。

Passkey 已按以下生产域名配置：

```text
WEBAUTHN_RP_ID=studio.loshin.org
WEBAUTHN_RP_ORIGINS=https://studio.loshin.org
```

## 3. 构建并启动

在宝塔终端执行：

```bash
cd /www/wwwroot/pcplatform
docker pull docker.m.daocloud.io/library/alpine:3.22
docker compose --env-file .env -f deploy/compose/compose.yml config
docker compose --env-file .env -f deploy/compose/compose.yml up -d --build
```

第一条命令用于单独确认镜像仓库连通性。若它仍失败，应先查看完整错误和 `docker info` 中的 Registry Mirrors，不要连续重启相同构建。

首次启动会按顺序完成：PostgreSQL 健康检查 → 数据库迁移 → API/Worker → 两个前端。PostgreSQL 数据和本地资产分别保存在 Docker named volume 中，重新构建容器不会删除数据。

检查状态和日志：

```bash
docker compose --env-file .env -f deploy/compose/compose.yml ps
docker compose --env-file .env -f deploy/compose/compose.yml logs --tail=100 migrate api worker
```

## 4. 宝塔网站反向代理

Compose 只向宿主回环地址开放两个端口，不占用宝塔的 80/443：

| 宝塔网站 | 反向代理目标 |
| --- | --- |
| `loshin.org` | `http://127.0.0.1:23000` |
| `studio.loshin.org` | `http://127.0.0.1:23001` |

两个网站都申请 HTTPS，并开启 HTTP 跳转 HTTPS。Studio 网站建议在 Nginx 配置中增加：

```nginx
client_max_body_size 100m;
proxy_read_timeout 300s;
```

API 只在 Docker 网络内提供给两个前端，不需要公网域名或宿主机端口。

## 5. 后续更新

重新生成并上传新部署包，但保留服务器原来的 `.env`，然后执行：

```bash
docker compose --env-file .env -f deploy/compose/compose.yml up -d --build
```

不要执行 `docker compose down -v`，其中 `-v` 会删除 PostgreSQL 和本地资产卷。

Worker 负责发布、全文及向量索引、Inbox 智能处理、AI 建议和 Webhook 投递，生产环境不要单独停用 Worker。

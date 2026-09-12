# Personal Content Platform

围绕 Capture → Organize → Create → Connect → Publish → Discover → Reuse 构建的个人内容中枢。当前实现覆盖架构方案 Phase 0–5 的可运行基线，包括身份、内容、发布、资产、知识关系、AI 建议、导出、Webhook 与生产运维。

## 工程结构

```text
apps/                 Studio 与 Public 两个 Next.js 应用
cmd/                  API、Worker、迁移和容器健康检查入口
internal/<module>/    domain/application/ports/infrastructure/transport 分层模块
internal/platform/    配置、数据库、迁移、Outbox 等平台能力
deploy/               宝塔反向代理与 Docker Compose 部署配置
docs/                 架构决策和运维 Runbook
scripts/              Windows 验证及 Linux 交叉编译脚本
```

`internal/app/app.go` 仅作为 Composition Root 装配模块；业务规则不得写进入口或路由文件。

## 本地开发

需要 Go 1.26、Node.js 24、npm 和 PostgreSQL 17（推荐 pgvector 镜像）。

1. 复制 `.env.example` 为项目根目录的 `.env`，填写开发配置。程序默认自动加载该文件；也可通过 `APP_ENV_FILE` 指定其他路径，已有系统环境变量优先。
2. 执行 `go run ./cmd/migrate up` 初始化数据库。
3. 分别启动 API、Worker、Studio 和公开站点。

```powershell
go run ./cmd/api
go run ./cmd/worker
npm ci
npm run dev:studio
npm run dev:public
```

- API：`http://localhost:8080`
- Studio：`http://localhost:3001`
- Public Web：`http://localhost:3000`
- 存活/就绪检查：`GET /healthz`、`GET /readyz`

执行全部静态、测试、前端构建和 Linux 交叉编译门禁：

```powershell
.\scripts\verify.ps1
```

对 `.env` 指向的开发数据库执行会自动创建并清理隔离 Workspace 的 API 冒烟测试：

```powershell
$env:PCP_INTEGRATION_TEST='1'
$env:APP_ENV_FILE=(Resolve-Path '.env').Path
go test ./tests/smoke -run TestCorePlatformWorkflows -v -count=1
```

## 主要能力

- 邮箱密码登录使用 Argon2id，会话使用 HttpOnly Cookie；同时支持 WebAuthn Passkey。
- Workspace 设置可在 Studio 中管理站点、语言、登录、发布、Filesystem/S3/R2/阿里 OSS 与 OpenAI 兼容 Provider；部署级参数保留在 `.env`。
- Article、Note、Page 采用不可变 Revision，`zh-CN` 与 `en` 独立维护并发布到 Website/RSS。
- 资产存储可选 `filesystem`、`s3`、`r2`、`oss`，统一通过 Storage Port 使用。
- AI 接入 OpenAI-compatible API，输出先生成可审查的 Suggestion；用户接受后才写入草稿或知识关系。
- Inbox 智能处理与 AI 建议由 Worker 执行，Studio 会跟踪任务状态；API 和 Worker 必须同时运行。
- Workspace manifest 导出包含全部修订、草稿、资产 hash、知识关系和发布记录。
- Webhook 使用 Outbox 事件、HMAC-SHA256 签名、有限重试和投递审计。

Webhook 密钥按已确认方案明文保存在 Workspace 数据库配置中，API 只返回掩码，不返回密钥正文；旧数据中的 `secret_ref` 仍可兼容读取。可用 API：

```text
GET    /v1/workspaces/{workspaceID}/webhooks
POST   /v1/workspaces/{workspaceID}/webhooks
PATCH  /v1/workspaces/{workspaceID}/webhooks/{endpointID}
DELETE /v1/workspaces/{workspaceID}/webhooks/{endpointID}
GET    /v1/workspaces/{workspaceID}/exports/manifest.json
```

## 部署

生产环境使用 Docker Compose 运行 PostgreSQL + pgvector、API、Worker、Studio 和 Public Web，Node.js 24 固定在前端镜像内；宝塔负责域名、HTTPS 和反向代理。Windows 执行 `deploy/build.ps1` 可生成完整 Docker 部署包。具体步骤见 [deploy/README.md](deploy/README.md)，备份恢复与故障处理见 [docs/runbooks](docs/runbooks)。

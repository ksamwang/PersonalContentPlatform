# Workspace 设置

Studio 的“设置”按 Workspace 隔离。Owner 可以修改，Editor 和 Viewer 只能读取非敏感配置；每次修改都会写入独立审计记录。

## 保留在部署环境的配置

应用进程在数据库可用前需要以下配置，因此不由 Studio 管理：

- `APP_ENV`、`HTTP_ADDR`
- `DATABASE_URL` 或 `DATABASE_URL_FILE`、`DB_POOL_MAX`
- `SESSION_COOKIE_NAME`、`SESSION_TTL`（新 Workspace 的回退值和 Cookie 上限）
- `PASSWORD_PEPPER`
- `WEBAUTHN_RP_ID`、`WEBAUTHN_RP_ORIGINS`
- Next.js 服务端使用的 `API_ORIGIN`、`WORKSPACE_SLUG`

旧版对象存储、AI 和 Webhook 环境变量仍作为兼容回退读取，但不再出现在 `.env.example`；Workspace 保存对应设置后以数据库配置为准。

## 数据库中的配置

- 工作区、语言、时区、站点介绍、RSS、登录方式与默认发布渠道保存在 `workspaces.settings_json`。
- AI Provider 保存在 `ai_provider_configs`。
- 对象存储保存在 `storage_profiles`；每个 Upload Intent 和 Blob 固化实际使用的 Profile ID。
- Webhook 密钥保存在 `webhook_endpoints.secret_value`，旧 `secret_ref` 继续兼容。

根据项目决定，Provider 和 Webhook 密钥以明文存在数据库中。API 永不返回完整密钥，只返回掩码和“已配置”状态；审计记录不包含密钥正文。数据库备份、只读副本和运维账号应按包含密钥的高敏数据保护。

## 存储切换

保存新的活动 Storage Profile 后，新上传使用新 Provider；旧 Blob 继续通过自身的 `storage_profile_id` 读取原 Provider。切换配置不自动搬迁旧对象。

Filesystem Profile 只能指向部署时 `OBJECT_STORAGE_BASE_PATH` 确定的根目录或其子目录，不能通过 Studio 指定服务器上的任意路径。

## 连接测试

- Filesystem：在存储根目录创建并删除临时探针文件。
- S3/R2：执行 `HeadBucket`。
- 阿里 OSS：列出最多一个对象以验证 Endpoint、Bucket 和凭据。
- AI：向配置的 OpenAI 兼容接口发送最小生成请求；该操作可能产生极少量 Provider 用量。

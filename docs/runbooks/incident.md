# 故障处理 Runbook

- API 不可用：检查 `/readyz`、数据库连接池和最近迁移。
- Worker 积压：分别检查 Outbox 未派发事件和 `jobs` 中的 `pending/retry_wait/dead` 任务；修复 Provider 或网络配置后等待重试，不直接删除任务。
- 发布漂移：以 Publication/Attempt 为事实执行 reconciliation。
- 私密内容泄露风险：收紧 visibility、清理 CDN，并审计 Publication View。
- AI Provider 故障：停用 AI 路由；内容编辑与发布不得受影响。

API 访问日志包含 request ID、路径、状态码、响应大小和耗时，可用 request ID 串联反向代理与应用日志。日志不会记录请求正文和 Provider 密钥。

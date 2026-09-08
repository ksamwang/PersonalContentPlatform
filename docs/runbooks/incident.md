# 故障处理 Runbook

- API 不可用：检查 `/readyz`、数据库连接池和最近迁移。
- Worker 积压：检查 Outbox 未派发数量、最老事件和错误码；修复后由重试恢复，不删除事件。
- 发布漂移：以 Publication/Attempt 为事实执行 reconciliation。
- 私密内容泄露风险：收紧 visibility、清理 CDN，并审计 Publication View。
- AI Provider 故障：停用 AI 路由；内容编辑与发布不得受影响。

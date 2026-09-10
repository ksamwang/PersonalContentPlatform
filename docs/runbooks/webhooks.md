# Webhook Runbook

Webhook 端点由 Studio 会话认证后的 Workspace API 管理。`event_types` 为空表示接收所有 Outbox 事件；否则只投递列出的事件类型。

签名请求头：

- `X-PCP-Event-ID`：稳定事件 ID，可用于接收方去重。
- `X-PCP-Event-Type`：事件类型。
- `X-PCP-Timestamp`：Unix 秒时间戳。
- `X-PCP-Signature`：`sha256=<hex>`。

签名内容为 `<timestamp>.<原始请求体>`，使用端点保存在数据库中的 `secret_value` 执行 HMAC-SHA256；旧端点仍可通过 `secret_ref` 读取环境变量。接收方应比较常量时间签名、限制时间窗并按事件 ID 幂等处理。

Worker 对非 2xx、网络错误和缺失密钥执行有限重试，投递结果保存在 `webhook_attempts`。排障时先检查端点是否启用、密钥是否已配置、目标证书和最近响应摘要；不要把真实密钥写入日志或工单。

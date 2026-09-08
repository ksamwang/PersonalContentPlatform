# 备份与恢复 Runbook

- PostgreSQL 每日执行 custom-format 备份；成熟环境开启 WAL/PITR。
- 对象存储开启版本控制，避免删除仍被 Blob manifest 引用的对象。
- 每周下载 Workspace manifest，保存数据库之外的第二份内容清单。
- 备份加密并存放到不同故障域。

恢复时在隔离环境恢复数据库和对象，按 manifest 核对 SHA-256，然后验证登录、读取、搜索与发布。每季度至少演练一次并记录实际 RPO/RTO。

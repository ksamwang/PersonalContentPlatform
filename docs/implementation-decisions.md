# 实施决策

## 已确认范围

- 登录：邮箱密码和 Passkey。
- 开发环境：Windows；生产目标：Debian Linux。
- 部署：生产环境不使用 Docker 或 systemd unit；在 Windows 交叉编译 Go Linux 可执行文件，宝塔管理进程守护、域名、HTTPS、反向代理和 Next.js 前端。
- 对象存储：Filesystem（本地）、S3、Cloudflare R2、阿里云 OSS。
- 首版内容：Article、Note、Page；首版语言：`zh-CN`、`en`。
- 首版渠道：Website 与 RSS。
- 首版媒体：图片上传、封面、正文插图、缩略图和公开访问。
- 私密内容：服务端权限与加密，不做端到端加密。
- AI：OpenAI-compatible provider abstraction；高影响结果需人工确认。

## 工程约束

- 按领域模块拆分 `domain/application/ports/infrastructure/transport`，入口仅做装配。
- 每个架构 Phase 独立 Git 提交，全部完成并验证后统一推送。
- 真实凭据不得提交；开发数据库仅用于迁移与集成测试。
- 生产数据库必须启用 TLS 或通过受限网络访问。

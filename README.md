# Personal Content Platform

围绕 Capture → Organize → Create → Connect → Publish → Discover → Reuse 构建的个人内容中枢。

## 本地开发

1. 复制 `.env.example` 为 `.env` 并填写开发配置。
2. 执行 `go run ./cmd/migrate up` 初始化数据库。
3. 执行 `go run ./cmd/api` 启动 API。
4. 执行 `go run ./cmd/worker` 启动后台任务进程。

健康检查：`GET http://localhost:8080/healthz`。

```powershell
go test ./...
go run ./cmd/migrate up
go run ./cmd/api
go run ./cmd/worker
.\scripts\build-linux.ps1
```

Docker Compose 部署说明见 [deploy/README.md](deploy/README.md)。

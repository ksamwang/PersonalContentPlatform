$ErrorActionPreference = 'Stop'
$output = Join-Path $PSScriptRoot '..\bin\linux-amd64'
New-Item -ItemType Directory -Force -Path $output | Out-Null
$env:GOOS = 'linux'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'
go build -trimpath -ldflags '-s -w' -o (Join-Path $output 'pcp-api') ./cmd/api
go build -trimpath -ldflags '-s -w' -o (Join-Path $output 'pcp-worker') ./cmd/worker
go build -trimpath -ldflags '-s -w' -o (Join-Path $output 'pcp-migrate') ./cmd/migrate
Write-Host "Linux binaries written to $output"

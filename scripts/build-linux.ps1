param(
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '..\bin\linux-amd64')
)

$ErrorActionPreference = 'Stop'
$output = [System.IO.Path]::GetFullPath($OutputDirectory)
New-Item -ItemType Directory -Force -Path $output | Out-Null
$previousGOOS = $env:GOOS
$previousGOARCH = $env:GOARCH
$previousCGO = $env:CGO_ENABLED
try {
    $env:GOOS = 'linux'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    go build -trimpath -ldflags '-s -w' -o (Join-Path $output 'pcp-api') ./cmd/api
    go build -trimpath -ldflags '-s -w' -o (Join-Path $output 'pcp-worker') ./cmd/worker
    go build -trimpath -ldflags '-s -w' -o (Join-Path $output 'pcp-migrate') ./cmd/migrate
    Write-Host "Linux binaries written to $output"
} finally {
    $env:GOOS = $previousGOOS
    $env:GOARCH = $previousGOARCH
    $env:CGO_ENABLED = $previousCGO
}

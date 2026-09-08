$ErrorActionPreference = 'Stop'

function Invoke-Checked {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Description,
        [Parameter(Mandatory = $true)]
        [scriptblock]$Command
    )

    Write-Host "==> $Description"
    & $Command
    if ($LASTEXITCODE) {
        throw "$Description failed with exit code $LASTEXITCODE"
    }
}

$repositoryRoot = Resolve-Path (Join-Path $PSScriptRoot '..')
Push-Location $repositoryRoot
try {
    $unformatted = gofmt -l cmd internal
    if ($LASTEXITCODE) {
        throw "gofmt check failed with exit code $LASTEXITCODE"
    }
    if ($unformatted) {
        throw "gofmt required for: $($unformatted -join ', ')"
    }

    Invoke-Checked 'Go tests' { go test ./... }
    Invoke-Checked 'Go vet' { go vet ./... }
    Invoke-Checked 'Go command builds' { go build ./cmd/... }
    Invoke-Checked 'Frontend typecheck' { npm run typecheck }
    Invoke-Checked 'Frontend production builds' { npm run build:web }
    Invoke-Checked 'Linux cross build' { & (Join-Path $PSScriptRoot 'build-linux.ps1') }
    Invoke-Checked 'Git whitespace check' { git diff --check }
} finally {
    Pop-Location
}

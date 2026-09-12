param(
    [string]$Version = '',
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '..\deploy\output')
)

$ErrorActionPreference = 'Stop'
$projectRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))

if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = (git -C $projectRoot rev-parse --short HEAD).Trim()
}
if ($Version -notmatch '^[0-9A-Za-z._-]+$') {
    throw 'Version may only contain letters, numbers, dots, underscores, and hyphens.'
}

$resolvedOutput = [System.IO.Path]::GetFullPath($OutputDirectory)
$releaseRoot = [System.IO.Path]::GetFullPath((Join-Path $resolvedOutput "pcplatform-docker-$Version"))
$archivePath = "$releaseRoot.zip"
if (-not $releaseRoot.StartsWith($resolvedOutput + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw 'Resolved release directory is outside OutputDirectory.'
}

if (Test-Path -LiteralPath $releaseRoot) {
    Remove-Item -LiteralPath $releaseRoot -Recurse -Force
}
if (Test-Path -LiteralPath $archivePath) {
    Remove-Item -LiteralPath $archivePath -Force
}
New-Item -ItemType Directory -Force -Path $releaseRoot | Out-Null

function Copy-SourceTree {
    param([string]$RelativePath)

    $sourceRoot = Join-Path $projectRoot $RelativePath
    Get-ChildItem -LiteralPath $sourceRoot -Recurse -Force -File | ForEach-Object {
        $relativeFile = [System.IO.Path]::GetRelativePath($projectRoot, $_.FullName)
        $segments = $relativeFile -split '[\\/]'
        $excluded = ($segments | Where-Object { $_ -in @('.next', 'node_modules') }) -or
            $_.Extension -eq '.tsbuildinfo' -or
            $_.Name -in @('AGENTS.md', 'CLAUDE.md') -or
            $_.Name.StartsWith('.env')

        if (-not $excluded) {
            $targetPath = Join-Path $releaseRoot $relativeFile
            New-Item -ItemType Directory -Force -Path (Split-Path $targetPath) | Out-Null
            Copy-Item -LiteralPath $_.FullName -Destination $targetPath
        }
    }
}

Copy-SourceTree 'cmd'
Copy-SourceTree 'internal'
Copy-SourceTree 'apps'

@('go.mod', 'go.sum', 'package.json', 'package-lock.json', 'Dockerfile', 'Dockerfile.web', '.dockerignore') | ForEach-Object {
    Copy-Item -LiteralPath (Join-Path $projectRoot $_) -Destination $releaseRoot
}

New-Item -ItemType Directory -Force -Path (Join-Path $releaseRoot 'deploy\compose') | Out-Null
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\compose\compose.yml') -Destination (Join-Path $releaseRoot 'deploy\compose\compose.yml')
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\compose\.env.example') -Destination (Join-Path $releaseRoot '.env')
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\README.md') -Destination (Join-Path $releaseRoot 'README.md')
Set-Content -LiteralPath (Join-Path $releaseRoot 'VERSION') -Value $Version -Encoding ascii

Compress-Archive -LiteralPath $releaseRoot -DestinationPath $archivePath -CompressionLevel Optimal
Write-Host "Docker deployment package written to $archivePath"

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
$releaseRoot = [System.IO.Path]::GetFullPath((Join-Path $resolvedOutput "pcplatform-$Version"))
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

New-Item -ItemType Directory -Force -Path (Join-Path $releaseRoot 'bin') | Out-Null
& (Join-Path $PSScriptRoot 'build-linux.ps1') -OutputDirectory (Join-Path $releaseRoot 'bin')

function Copy-ApplicationSource {
    param([string]$Application)

    $sourceRoot = Join-Path $projectRoot "apps\$Application"
    $targetRoot = Join-Path $releaseRoot "apps\$Application"
    $excludedDirectories = @('.next', 'node_modules')
    $excludedFiles = @('AGENTS.md', 'CLAUDE.md')

    Get-ChildItem -LiteralPath $sourceRoot -Recurse -Force -File | ForEach-Object {
        $relativePath = [System.IO.Path]::GetRelativePath($sourceRoot, $_.FullName)
        $segments = $relativePath -split '[\\/]'
        $excluded = ($segments | Where-Object { $_ -in $excludedDirectories }) -or
            $_.Name -in $excludedFiles -or
            $_.Extension -eq '.tsbuildinfo' -or
            $_.Name.StartsWith('.env')

        if (-not $excluded) {
            $targetPath = Join-Path $targetRoot $relativePath
            New-Item -ItemType Directory -Force -Path (Split-Path $targetPath) | Out-Null
            Copy-Item -LiteralPath $_.FullName -Destination $targetPath
        }
    }
}

Copy-ApplicationSource 'studio-web'
Copy-ApplicationSource 'public-web'
Copy-Item -LiteralPath (Join-Path $projectRoot 'package.json') -Destination $releaseRoot
Copy-Item -LiteralPath (Join-Path $projectRoot 'package-lock.json') -Destination $releaseRoot

New-Item -ItemType Directory -Force -Path (Join-Path $releaseRoot 'config') | Out-Null
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\baota\pcplatform.env.example') -Destination (Join-Path $releaseRoot 'config\pcplatform.env')
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\baota\studio.env.example') -Destination (Join-Path $releaseRoot 'apps\studio-web\.env.production')
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\baota\public-web.env.example') -Destination (Join-Path $releaseRoot 'apps\public-web\.env.production')
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\baota\build-web.sh') -Destination (Join-Path $releaseRoot 'build-web.sh')
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\README.md') -Destination (Join-Path $releaseRoot 'README.md')
New-Item -ItemType Directory -Force -Path (Join-Path $releaseRoot 'data\assets') | Out-Null

Set-Content -LiteralPath (Join-Path $releaseRoot 'VERSION') -Value $Version -Encoding ascii
Get-ChildItem -LiteralPath (Join-Path $releaseRoot 'bin') -File |
    Get-FileHash -Algorithm SHA256 |
    ForEach-Object { "{0}  bin/{1}" -f $_.Hash.ToLowerInvariant(), $_.Path.Substring($_.Path.LastIndexOf('\') + 1) } |
    Set-Content -LiteralPath (Join-Path $releaseRoot 'SHA256SUMS') -Encoding ascii

Compress-Archive -LiteralPath $releaseRoot -DestinationPath $archivePath -CompressionLevel Optimal
Write-Host "Complete Baota deployment package written to $archivePath"

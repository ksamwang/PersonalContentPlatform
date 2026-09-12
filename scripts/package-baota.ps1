param(
    [string]$Version = '',
    [string]$OutputDirectory = (Join-Path $PSScriptRoot '..\bin\release')
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

Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\baota') -Destination (Join-Path $releaseRoot 'baota') -Recurse
Copy-Item -LiteralPath (Join-Path $projectRoot 'deploy\README.md') -Destination (Join-Path $releaseRoot 'README.md')

Set-Content -LiteralPath (Join-Path $releaseRoot 'VERSION') -Value $Version -Encoding ascii
Get-ChildItem -LiteralPath (Join-Path $releaseRoot 'bin') -File |
    Get-FileHash -Algorithm SHA256 |
    ForEach-Object { "{0}  bin/{1}" -f $_.Hash.ToLowerInvariant(), $_.Path.Substring($_.Path.LastIndexOf('\') + 1) } |
    Set-Content -LiteralPath (Join-Path $releaseRoot 'SHA256SUMS') -Encoding ascii

Compress-Archive -LiteralPath $releaseRoot -DestinationPath $archivePath -CompressionLevel Optimal
Write-Host "Baota release package written to $archivePath"

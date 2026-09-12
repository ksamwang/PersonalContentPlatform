param(
    [string]$Version = ''
)

$ErrorActionPreference = 'Stop'
& (Join-Path $PSScriptRoot '..\scripts\package-baota.ps1') -Version $Version -OutputDirectory (Join-Path $PSScriptRoot 'output')

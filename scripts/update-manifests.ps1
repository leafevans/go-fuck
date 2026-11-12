<#
.SYNOPSIS
Regenerate the distribution manifests for a released go-fuck version.

.DESCRIPTION
Downloads checksums.txt and the tag archive of a GitHub release, runs
tools/update-packaging, and verifies the packaging tests. By default the
regenerated files are left in the working tree for review; pass -Commit to
commit them (the commit is not pushed automatically).

.PARAMETER Version
Release version, with or without a leading "v" (for example 2.19.0).

.PARAMETER Commit
Commit the regenerated manifests with the usual release commit message.

.EXAMPLE
./scripts/update-manifests.ps1 -Version 2.19.0

.EXAMPLE
./scripts/update-manifests.ps1 -Version v2.19.0 -Commit
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$Version,

    [switch]$Commit
)

$ErrorActionPreference = "Stop"

$version = $Version.TrimStart("v")
if ($version -notmatch '^\d+\.\d+\.\d+$') {
    throw "invalid version '$Version': expected X.Y.Z"
}
if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "the GitHub CLI (gh) is required to download the release assets"
}
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "the Go toolchain is required to run tools/update-packaging"
}

$root = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$work = Join-Path ([System.IO.Path]::GetTempPath()) ("go-fuck-manifests-" + [guid]::NewGuid().ToString("N"))
$checksums = Join-Path $work "checksums.txt"
$archive = Join-Path $work "go-fuck-$version.tar.gz"

New-Item -ItemType Directory -Force -Path $work | Out-Null
Push-Location $root
try {
    Write-Host "downloading release assets for v$version"
    & gh release download "v$version" --pattern checksums.txt --dir $work --clobber
    if ($LASTEXITCODE -ne 0) {
        throw "failed to download checksums.txt for v$version"
    }

    $url = "https://github.com/leafevans/go-fuck/archive/refs/tags/v$version.tar.gz"
    Invoke-WebRequest -Uri $url -OutFile $archive -TimeoutSec 120
    if (-not (Test-Path -LiteralPath $archive)) {
        throw "failed to download $url"
    }

    & go run ./tools/update-packaging -version $version -checksums $checksums -aur-source $archive
    if ($LASTEXITCODE -ne 0) {
        throw "tools/update-packaging failed"
    }

    & go test ./packaging -count=1
    if ($LASTEXITCODE -ne 0) {
        throw "packaging tests failed"
    }

    $changed = & git status --porcelain -- Formula bucket packaging
    if (-not $changed) {
        Write-Host "manifests for v$version are already up to date"
        return
    }

    Write-Host "changed files:"
    $changed | ForEach-Object { Write-Host "  $_" }
    if ($Commit) {
        & git add -- Formula bucket packaging
        & git commit -m "chore: update distribution manifests for v$version"
        if ($LASTEXITCODE -ne 0) {
            throw "failed to commit the regenerated manifests"
        }
        Write-Host "committed the manifests for v$version (not pushed)"
    } else {
        Write-Host "review the diff and commit it, or rerun with -Commit"
    }
} finally {
    Pop-Location
    Remove-Item -Recurse -Force $work -ErrorAction SilentlyContinue
}

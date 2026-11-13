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

.PARAMETER PushAur
Publish the regenerated package files to the AUR with scripts/push-aur.sh
after the manifests are up to date or committed. Git Bash or WSL and an SSH
key registered with the AUR are required.

.PARAMETER AurAddress
IPv4 address to pass to scripts/push-aur.sh as --address. Use it when the
DNS-over-HTTPS lookup of aur.archlinux.org is blocked. Requires -PushAur.

.PARAMETER AurDryRun
Pass --dry-run to scripts/push-aur.sh: the AUR publish is only printed, not
executed. Requires -PushAur.

.EXAMPLE
./scripts/update-manifests.ps1 -Version 2.19.0

.EXAMPLE
./scripts/update-manifests.ps1 -Version v2.19.0 -Commit -PushAur

.EXAMPLE
./scripts/update-manifests.ps1 -Version 2.19.0 -PushAur -AurAddress 209.126.35.78
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$Version,

    [switch]$Commit,

    [switch]$PushAur,

    [string]$AurAddress,

    [switch]$AurDryRun
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

if ($AurAddress -and -not $PushAur) {
    throw "-AurAddress requires -PushAur"
}
if ($AurDryRun -and -not $PushAur) {
    throw "-AurDryRun requires -PushAur"
}
if ($AurAddress -and $AurAddress -notmatch '^[0-9]{1,3}(\.[0-9]{1,3}){3}$') {
    throw "invalid -AurAddress '$AurAddress': expected an IPv4 address"
}

# Find-GitBash returns the Git Bash executable, or $null when Git for Windows is
# not installed in one of the usual locations.
function Find-GitBash {
    $candidates = @(
        (Join-Path $env:ProgramFiles "Git\bin\bash.exe"),
        (Join-Path $env:LOCALAPPDATA "Programs\Git\bin\bash.exe"),
        "C:\Program Files (x86)\Git\bin\bash.exe"
    )
    foreach ($candidate in $candidates) {
        if ($candidate -and (Test-Path -LiteralPath $candidate)) {
            return $candidate
        }
    }
    return $null
}

# Convert-ToUnixPath converts a Windows path to the form that Git Bash (/c/...)
# or WSL (/mnt/c/...) uses.
function Convert-ToUnixPath {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path,
        [Parameter(Mandatory = $true)]
        [string]$Prefix
    )
    $drive = $Path.Substring(0, 1).ToLower()
    $rest = $Path.Substring(2).Replace("\", "/")
    return "$Prefix$drive$rest"
}

# Publish-AurPackage publishes the package files of the working tree with
# scripts/push-aur.sh, through Git Bash when it is installed and WSL otherwise.
function Publish-AurPackage {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Version,
        [string]$Address,
        [switch]$DryRun
    )
    $aurArguments = "--version $Version"
    if ($Address) {
        $aurArguments += " --address $Address"
    }
    if ($DryRun) {
        $aurArguments += " --dry-run"
    }
    $scriptRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
    $bash = Find-GitBash
    if ($bash) {
        $unixRoot = Convert-ToUnixPath -Path $scriptRoot -Prefix "/"
        Write-Host "publishing the AUR package with $bash"
        & $bash -lc "cd '$unixRoot' && bash scripts/push-aur.sh $aurArguments"
        if ($LASTEXITCODE -ne 0) {
            throw "scripts/push-aur.sh failed"
        }
        return
    }
    if (Get-Command wsl.exe -ErrorAction SilentlyContinue) {
        $unixRoot = Convert-ToUnixPath -Path $scriptRoot -Prefix "/mnt/"
        Write-Host "publishing the AUR package with wsl.exe"
        & wsl.exe -e bash -lc "cd '$unixRoot' && bash scripts/push-aur.sh $aurArguments"
        if ($LASTEXITCODE -ne 0) {
            throw "scripts/push-aur.sh failed"
        }
        return
    }
    throw "publishing to the AUR needs Git Bash or WSL: install Git for Windows or enable WSL"
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
    $downloaded = $false
    try {
        Invoke-WebRequest -Uri $url -OutFile $archive -TimeoutSec 120
        $downloaded = $true
    } catch {
        Write-Verbose "Invoke-WebRequest failed: $($_.Exception.Message)"
    }
    if (-not $downloaded) {
        # Windows PowerShell's Invoke-WebRequest sometimes fails with a closed
        # TLS connection; curl.exe ships with Windows 10+ and is more robust.
        $curl = Get-Command curl.exe -ErrorAction SilentlyContinue
        if (-not $curl) {
            throw "cannot download $url and curl.exe is not available"
        }
        & $curl.Source -fsSL --retry 3 --retry-delay 2 -o $archive $url
        if ($LASTEXITCODE -ne 0) {
            throw "failed to download $url"
        }
    }
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
        if ($PushAur) {
            Publish-AurPackage -Version $version -Address $AurAddress -DryRun:$AurDryRun
        }
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
    if ($PushAur) {
        Publish-AurPackage -Version $version -Address $AurAddress -DryRun:$AurDryRun
    }
} finally {
    Pop-Location
    Remove-Item -Recurse -Force $work -ErrorAction SilentlyContinue
}

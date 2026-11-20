<#
.SYNOPSIS
Updates the Windows and WSL go-fuck binaries from the GitHub releases.

.DESCRIPTION
Downloads the release archive and checksums.txt, verifies the SHA256 checksum
and replaces the installed go-fuck.exe with rollback on failure. Downloads have
bounded timeouts and up to three attempts. Unless -SkipWsl is used, the script
also runs update-go-fuck.sh in WSL, which has to be next to this script.

.EXAMPLE
./update-go-fuck.ps1
./update-go-fuck.ps1 -Version 2.11.0
./update-go-fuck.ps1 -SkipWsl
#>
[CmdletBinding()]
param(
    [string]$Version = "latest",
    [string]$InstallDir = (Join-Path $HOME ".local\bin"),
    [string]$WslDistro = "",
    [switch]$SkipWsl
)

$ErrorActionPreference = "Stop"
$repo = "leafevans/go-fuck"

function Invoke-WithRetry {
    param([scriptblock]$Operation, [string]$Description)
    for ($attempt = 1; $attempt -le 3; $attempt++) {
        try {
            return (& $Operation)
        } catch {
            if ($attempt -eq 3) {
                throw "update-go-fuck: $Description failed after 3 attempts: $($_.Exception.Message)"
            }
            Write-Warning "${Description} failed (attempt $attempt/3); retrying: $($_.Exception.Message)"
            Start-Sleep -Seconds 2
        }
    }
}

function Save-Download {
    param([string]$Uri, [string]$OutFile, [int]$TimeoutSec = 120)
    Invoke-WithRetry -Description "download $Uri" -Operation {
        try {
            Invoke-WebRequest -UseBasicParsing -Uri $Uri -OutFile $OutFile `
                -TimeoutSec $TimeoutSec -UserAgent "go-fuck-updater" -ErrorAction Stop
        } catch {
            Remove-Item -LiteralPath $OutFile -Force -ErrorAction SilentlyContinue
            throw
        }
    }
}

# Resolve-LatestVersion finds the newest release tag. The GitHub API is
# rate limited for unauthenticated requests, so fall back to the redirect of
# the releases/latest page, which points at /releases/tag/vX.Y.Z.
function Resolve-LatestVersion {
    try {
        return Invoke-WithRetry -Description "resolve the latest version" -Operation {
            $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" `
                -TimeoutSec 20 -UserAgent "go-fuck-updater" -ErrorAction Stop
            if ($release.tag_name -notmatch '^v?([0-9]+\.[0-9]+\.[0-9]+)$') {
                throw "invalid release tag"
            }
            return $Matches[1]
        }
    } catch {
        Write-Verbose "GitHub API request failed; resolving the version from the release redirect"
        try {
            return Invoke-WithRetry -Description "resolve the release redirect" -Operation {
                $request = [System.Net.HttpWebRequest]::Create("https://github.com/$repo/releases/latest")
                $request.AllowAutoRedirect = $false
                $request.Method = "HEAD"
                $request.Timeout = 20000
                $request.ReadWriteTimeout = 20000
                $request.UserAgent = "go-fuck-updater"
                $response = $request.GetResponse()
                try {
                    $location = $response.Headers["Location"]
                    if ($location -notmatch '^https://github\.com/leafevans/go-fuck/releases/tag/v([0-9]+\.[0-9]+\.[0-9]+)$') {
                        throw "invalid release redirect"
                    }
                    return $Matches[1]
                } finally {
                    $response.Close()
                }
            }
        } catch {
            Write-Verbose "release redirect failed: $($_.Exception.Message)"
        }
        throw "update-go-fuck: cannot resolve the latest version"
    }
}

if ($Version -eq "latest") {
    $Version = Resolve-LatestVersion
}
if ($Version -notmatch '^v?([0-9]+\.[0-9]+\.[0-9]+)$') {
    throw "update-go-fuck: invalid version: expected X.Y.Z or vX.Y.Z"
}
$Version = $Matches[1]

# A 32-bit PowerShell process can run on a 64-bit or ARM64 Windows host.
$machine = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
$arch = switch ($machine) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "update-go-fuck: unsupported architecture: $machine" }
}

$base = "go-fuck_${Version}_windows_$arch"
$asset = "$base.zip"
$url = "https://github.com/$repo/releases/download/v$Version"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("go-fuck-update-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp | Out-Null
$staged = $null

try {
    Write-Host "downloading go-fuck $Version (windows/$arch)"
    Save-Download -Uri "$url/$asset" -OutFile (Join-Path $tmp $asset)
    Save-Download -Uri "$url/checksums.txt" -OutFile (Join-Path $tmp "checksums.txt") -TimeoutSec 20

    $pattern = '^(\S+)\s+' + [regex]::Escape($asset) + '$'
    $lines = @(Select-String -LiteralPath (Join-Path $tmp "checksums.txt") -Pattern $pattern)
    if ($lines.Count -ne 1) { throw "expected exactly one SHA256 checksum for $asset" }
    $expected = $lines[0].Matches[0].Groups[1].Value
    if ($expected -notmatch '^[0-9a-fA-F]{64}$') { throw "invalid SHA256 checksum for $asset" }
    $actual = (Get-FileHash -LiteralPath (Join-Path $tmp $asset) -Algorithm SHA256).Hash
    if ($expected -ne $actual) { throw "checksum verification failed" }

    Expand-Archive -LiteralPath (Join-Path $tmp $asset) -DestinationPath (Join-Path $tmp "bin") -Force
    $source = Join-Path $tmp "bin\go-fuck.exe"
    if (-not (Test-Path -LiteralPath $source -PathType Leaf)) { throw "archive does not contain go-fuck.exe" }
    $target = Join-Path $InstallDir "go-fuck.exe"
    if (Test-Path -LiteralPath $target -PathType Container) { throw "installation target is a directory: $target" }
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $staged = Join-Path $InstallDir (".go-fuck-" + [guid]::NewGuid().ToString("N") + ".tmp")
    Copy-Item -LiteralPath $source -Destination $staged -Force
    $backup = $null
    if (Test-Path -LiteralPath $target) {
        # Use a unique backup name: a fixed name can collide with a backup of
        # an earlier update that is still locked by antivirus or a running
        # shell, which made the rename fail even after a retry.
        $backup = Join-Path $InstallDir ("go-fuck.exe.old." + [guid]::NewGuid().ToString("N"))
        Rename-Item -LiteralPath $target -NewName (Split-Path -Path $backup -Leaf) -Force
    }
    try {
        Move-Item -LiteralPath $staged -Destination $target -Force
        $staged = $null
    } catch {
        $installError = $_
        if ($backup) {
            try {
                Move-Item -LiteralPath $backup -Destination $target -Force
            } catch {
                throw "update-go-fuck: cannot restore the old binary; backup retained at ${backup}: $($_.Exception.Message)"
            }
        }
        throw $installError
    }
    if ($backup) {
        Remove-Item -LiteralPath $backup -Force -ErrorAction SilentlyContinue
    }
    # Clean up backups of earlier updates. A backup can still be locked; that
    # is harmless, and the next update tries again.
    Get-ChildItem -LiteralPath $InstallDir -Filter "go-fuck.exe.old.*" -File -ErrorAction SilentlyContinue |
        Remove-Item -Force -ErrorAction SilentlyContinue
    Write-Host "go-fuck $Version installed in $InstallDir"
}
finally {
    if ($staged) { Remove-Item -LiteralPath $staged -Force -ErrorAction SilentlyContinue }
    Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

if ($SkipWsl) { return }

$script = Join-Path $PSScriptRoot "update-go-fuck.sh"
if (-not (Test-Path -LiteralPath $script)) {
    Write-Warning "update-go-fuck.sh not found next to this script; skipping WSL"
    return
}

# Stream the script through stdin. This handles spaces and quotes in its path
# and avoids a shared /tmp filename that concurrent updates could overwrite.
$content = [System.IO.File]::ReadAllText($script)
$arguments = @("-e", "sh", "-c", "tr -d '\r' | sh -s -- $Version")
$previousOutputEncoding = $OutputEncoding
$previousInputEncoding = [Console]::InputEncoding
try {
    # A UTF-8 BOM becomes part of the shebang when sh reads from stdin.
    # Windows PowerShell 5.1 also takes the native stdin writer's encoding
    # from Console.InputEncoding, independently of OutputEncoding.
    $OutputEncoding = [System.Text.UTF8Encoding]::new($false)
    [Console]::InputEncoding = $OutputEncoding
    if ($WslDistro -ne "") {
        $content | & wsl.exe -d $WslDistro @arguments
    }
    else {
        $content | & wsl.exe @arguments
    }
    $wslExitCode = $LASTEXITCODE
} finally {
    $OutputEncoding = $previousOutputEncoding
    [Console]::InputEncoding = $previousInputEncoding
}
if ($wslExitCode -ne 0) { throw "the WSL update failed" }

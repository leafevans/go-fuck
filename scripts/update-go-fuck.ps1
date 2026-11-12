<#
.SYNOPSIS
Updates the Windows and WSL go-fuck binaries from the GitHub releases.

.DESCRIPTION
Downloads the release archive and checksums.txt, verifies the SHA256 checksum
and replaces the installed go-fuck.exe. Unless -SkipWsl is used, the script
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

# Resolve-LatestVersion finds the newest release tag. The GitHub API is
# rate limited for unauthenticated requests, so fall back to the redirect of
# the releases/latest page, which points at /releases/tag/vX.Y.Z.
function Resolve-LatestVersion {
    try {
        $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
        return $release.tag_name.TrimStart("v")
    } catch {
        Write-Verbose "GitHub API request failed; resolving the version from the release redirect"
        $location = $null
        try {
            $request = [System.Net.HttpWebRequest]::Create("https://github.com/$repo/releases/latest")
            $request.AllowAutoRedirect = $false
            $request.Method = "HEAD"
            $request.Timeout = 30000
            $request.UserAgent = "go-fuck-updater"
            $response = $request.GetResponse()
            try {
                $location = $response.Headers["Location"]
            } finally {
                $response.Close()
            }
        } catch {
            Write-Verbose "release redirect failed: $($_.Exception.Message)"
        }
        if ($location -match "/tag/v([^/]+)") {
            return $Matches[1]
        }
        throw "cannot resolve the latest go-fuck version"
    }
}

if ($Version -eq "latest") {
    $Version = Resolve-LatestVersion
}

$base = "go-fuck_${Version}_windows_amd64"
$url = "https://github.com/$repo/releases/download/v$Version"
$tmp = Join-Path $env:TEMP ("go-fuck-update-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
    Write-Host "downloading go-fuck $Version (windows/amd64)"
    Invoke-WebRequest "$url/$base.zip" -OutFile (Join-Path $tmp "$base.zip")
    Invoke-WebRequest "$url/checksums.txt" -OutFile (Join-Path $tmp "checksums.txt")

    $line = Select-String -Path (Join-Path $tmp "checksums.txt") -Pattern " $base.zip" | Select-Object -First 1
    if (-not $line) { throw "checksum for $base.zip not found" }
    $expected = ($line.Line -split "\s+")[0]
    $actual = (Get-FileHash (Join-Path $tmp "$base.zip") -Algorithm SHA256).Hash.ToLower()
    if ($expected -ne $actual) { throw "checksum verification failed" }

    Expand-Archive -LiteralPath (Join-Path $tmp "$base.zip") -DestinationPath (Join-Path $tmp "bin") -Force
    $target = Join-Path $InstallDir "go-fuck.exe"
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    $backup = $null
    if (Test-Path $target) {
        # Use a unique backup name: a fixed name can collide with a backup of
        # an earlier update that is still locked by antivirus or a running
        # shell, which made the rename fail even after a retry.
        $backup = Join-Path $InstallDir ("go-fuck.exe.old." + [guid]::NewGuid().ToString("N"))
        Rename-Item -Path $target -NewName (Split-Path -Path $backup -Leaf) -Force
    }
    Copy-Item (Join-Path $tmp "bin\go-fuck.exe") $target -Force
    if ($backup) {
        Remove-Item -LiteralPath $backup -Force -ErrorAction SilentlyContinue
    }
    # Clean up backups of earlier updates. A backup can still be locked; that
    # is harmless, and the next update tries again.
    Get-ChildItem -Path $InstallDir -Filter "go-fuck.exe.old*" -File -ErrorAction SilentlyContinue |
        Remove-Item -Force -ErrorAction SilentlyContinue
    Write-Host "go-fuck $Version installed in $InstallDir"
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

if ($SkipWsl) { return }

$script = Join-Path $PSScriptRoot "update-go-fuck.sh"
if (-not (Test-Path $script)) {
    Write-Warning "update-go-fuck.sh not found next to this script; skipping WSL"
    return
}

$unixScript = "/mnt/" + $script.Substring(0, 1).ToLower() + (($script.Substring(2)) -replace "\\", "/")
$unixVersion = if ($Version -eq "latest") { "latest" } else { $Version }
$command = "tr -d '\r' < '$unixScript' > /tmp/update-go-fuck.sh && bash /tmp/update-go-fuck.sh $unixVersion"
$arguments = @("-e", "bash", "-c", $command)
if ($WslDistro -ne "") {
    & wsl.exe -d $WslDistro @arguments
}
else {
    & wsl.exe @arguments
}
if ($LASTEXITCODE -ne 0) { throw "the WSL update failed" }

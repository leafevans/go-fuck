# Offline driver for updater_test.go. All downloads and WSL calls are mocked.
param([string]$Updater, [string]$InstallDir, [string]$Version)
$ErrorActionPreference = "Stop"
$fixture = $env:UPDATER_FIXTURES
$scenario = $env:UPDATER_SCENARIO
$global:updaterCounts = @{}
$env:PROCESSOR_ARCHITECTURE = "AMD64"
$env:PROCESSOR_ARCHITEW6432 = if ($scenario -eq "arm64") { "ARM64" } else { $null }

# Keep machine-local binaries out of the offline PATH resolution test.
function Get-Command {
    param([string]$Name, $CommandType, $ErrorAction)
    if ($Name -eq "go-fuck.exe" -and $CommandType -eq "Application") {
        if ($scenario -eq "path-shadowed") {
            [pscustomobject]@{ Source = Join-Path $fixture "earlier/go-fuck.exe" }
        } elseif ($scenario -eq "path-current") {
            [pscustomobject]@{ Source = (Resolve-Path -LiteralPath (Join-Path $InstallDir "go-fuck.exe")).ProviderPath }
        }
        return
    }
    Microsoft.PowerShell.CoreGet-Command @PSBoundParameters
}

function Start-Sleep { param([int]$Seconds) }

function Invoke-RestMethod {
    param([string]$Uri, [int]$TimeoutSec, [string]$UserAgent, $ErrorAction)
    if ($TimeoutSec -ne 20) { throw "missing API timeout" }
    return @{ tag_name = "v9.8.7" }
}

function Invoke-WebRequest {
    param([string]$Uri, [string]$OutFile, [int]$TimeoutSec, [string]$UserAgent,
        [switch]$UseBasicParsing, $ErrorAction)
    $kind = if ($Uri.EndsWith("/checksums.txt")) { "checksums" } elseif ($Uri.EndsWith(".tar.gz")) { "linux-archive" } else { "archive" }
    $expectedTimeout = if ($kind -eq "archive" -or $kind -eq "linux-archive") { 120 } else { 20 }
    if ($TimeoutSec -ne $expectedTimeout) { throw "missing download timeout" }
    Add-Content -LiteralPath (Join-Path $fixture "requests.log") -Value $kind
    $global:updaterCounts[$kind]++
    if (($scenario -eq "retry" -and $kind -eq "archive" -and $global:updaterCounts[$kind] -eq 1) -or
        ($scenario -eq "checksum-retry" -and $kind -eq "checksums" -and $global:updaterCounts[$kind] -eq 1) -or
        ($scenario -eq "download-failure" -and $kind -eq "archive")) {
        [System.IO.File]::WriteAllText($OutFile, "partial")
        throw "simulated transfer failure"
    }
    Microsoft.PowerShell.Management\Copy-Item -LiteralPath (Join-Path $fixture $kind) -Destination $OutFile -Force
}

function Copy-Item {
    param([string]$LiteralPath, [string]$Destination, [switch]$Force)
    if ($scenario -eq "stage-failure") { throw "simulated stage failure" }
    Microsoft.PowerShell.Management\Copy-Item -LiteralPath $LiteralPath -Destination $Destination -Force:$Force
}

function Move-Item {
    param([string]$LiteralPath, [string]$Destination, [switch]$Force)
    if ($scenario -eq "replace-failure" -and $LiteralPath.EndsWith(".tmp")) {
        throw "simulated replace failure"
    }
    Microsoft.PowerShell.Management\Move-Item -LiteralPath $LiteralPath -Destination $Destination -Force:$Force
}

if (-not $scenario.StartsWith("native-wsl")) {
    function wsl.exe {
        $content = @($input) -join "`n"
        Add-Content -LiteralPath (Join-Path $fixture "wsl.log") -Value (($args -join " ") + "`n" + $content)
        $arguments = $args -join " "
        if ($arguments.Contains("--exec uname -m")) {
            Write-Output "x86_64"
            $global:LASTEXITCODE = 0
        } elseif ($arguments.Contains("--exec wslpath -u")) {
            Write-Output "/mnt/c/three/verified/archive.tar.gz"
            $global:LASTEXITCODE = 0
        } elseif ($arguments.Contains("--exec sh -c") -and ($scenario -eq "wsl-timeout" -or $scenario -eq "wsl-fallback-checksum")) {
            $global:LASTEXITCODE = 1
        } else {
            $global:LASTEXITCODE = 0
        }
    }
}

try {
    $parameters = @{ Version = $Version; InstallDir = $InstallDir; SkipWsl = $true }
    if ($scenario.Contains("wsl")) {
        $parameters.SkipWsl = $false
        $parameters.WslDistro = "Test Distro"
    }
    # Exercise the encoding actually used by native pipelines, not just the
    # string seen by a function mock. Users may configure UTF-8 with a BOM.
    $OutputEncoding = [System.Text.UTF8Encoding]::new($true)
    $savedEncoding = $OutputEncoding
    [Console]::InputEncoding = [System.Text.UTF8Encoding]::new($true)
    $savedInputEncoding = [Console]::InputEncoding
    try {
        & $Updater @parameters
    } finally {
        if (-not [object]::ReferenceEquals($savedEncoding, $OutputEncoding)) { throw "OutputEncoding was not restored" }
        if ([Console]::InputEncoding.CodePage -ne $savedInputEncoding.CodePage -or
            [BitConverter]::ToString([Console]::InputEncoding.GetPreamble()) -ne [BitConverter]::ToString($savedInputEncoding.GetPreamble())) {
            throw "Console.InputEncoding was not restored"
        }
    }
} catch {
    [Console]::Error.WriteLine($_.Exception.Message)
    exit 1
}

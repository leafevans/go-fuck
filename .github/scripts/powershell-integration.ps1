# Checks the PowerShell specific behavior of go-fuck with a real PowerShell.
#
# The CI workflow runs this script once with Windows PowerShell 5.1 and once
# with PowerShell 7. It creates failing commands, asks go-fuck for a suggestion
# and checks both the suggestion and the result of running it.
param(
    [string]$Exe = "./go-fuck.exe"
)

$ErrorActionPreference = "Stop"

# The script changes into a temporary directory, so the executable is resolved
# to an absolute path first.
$Exe = (Resolve-Path -LiteralPath $Exe).Path

$root = Join-Path $env:TEMP ("go-fuck-powershell-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $root | Out-Null
$failures = 0

function Assert($name, $condition, $detail) {
    if ($condition) {
        Write-Host ("PASS " + $name)
    } else {
        Write-Host ("FAIL " + $name + ": " + $detail)
        $script:failures = $script:failures + 1
    }
}

function Get-Suggestion([string]$command) {
    $json = & $Exe --force-command $command --json 2>$null | Out-String
    $parsed = $json | ConvertFrom-Json
    if ($null -eq $parsed.suggestions -or $parsed.suggestions.Count -eq 0) {
        return ""
    }
    return $parsed.suggestions[0].command
}

Push-Location $root
try {
    $env:TF_SHELL = "powershell"
    $env:GOFUCK_SHELL = (Get-Process -Id $PID).Path
    $env:GOFUCK_WAIT_COMMAND = "20"

    # Removing a directory with children suggests -Recurse.
    New-Item -ItemType Directory -Path "nonempty" | Out-Null
    Set-Content -Path "nonempty\a.txt" -Value "x"
    $suggestion = Get-Suggestion ("rm " + (Join-Path $root "nonempty"))
    Assert "rm on a non-empty directory" ($suggestion -match "-Recurse") $suggestion
    if ($suggestion -ne "") {
        Invoke-Expression $suggestion | Out-Null
        Assert "rm -Recurse removes the directory" (-not (Test-Path -LiteralPath "nonempty")) "the directory still exists"
    }

    # Copying and moving to a missing directory creates it first.
    Set-Content -Path "file.txt" -Value "x"
    $suggestion = Get-Suggestion "cp file.txt nodir\file.txt"
    Assert "cp creates the missing directory" ($suggestion -match "mkdir -Force") $suggestion
    if ($suggestion -ne "") {
        Invoke-Expression $suggestion | Out-Null
        Assert "cp copies the file" (Test-Path -LiteralPath "nodir\file.txt") "the copy does not exist"
    }

    # Creating an existing directory adds -Force.
    New-Item -ItemType Directory -Path "existing" | Out-Null
    $suggestion = Get-Suggestion "mkdir existing"
    Assert "mkdir of an existing directory" ($suggestion -match "-Force") $suggestion
    if ($suggestion -ne "") {
        Invoke-Expression $suggestion | Out-Null
    }

    # An unquoted path with a space is quoted again.
    New-Item -ItemType Directory -Path (Join-Path $root "Program Files") | Out-Null
    Set-Content -LiteralPath (Join-Path $root "Program Files\marker.txt") -Value "x"
    $suggestion = Get-Suggestion ("cd " + $root + "\Program Files")
    Assert "cd quotes the path" ($suggestion -match "\x27") $suggestion
    if ($suggestion -ne "") {
        Invoke-Expression $suggestion | Out-Null
        Assert "cd enters the directory" (Test-Path -LiteralPath "marker.txt") "marker.txt is not in the current directory"
        Set-Location $root
    }

    # The version output uses the detected shell. cmd.exe merges the two streams
    # without turning the native stderr lines into PowerShell error records.
    $version = cmd /c "`"$Exe`" --version 2>&1"
    Assert "version mentions PowerShell" ($version -match "PowerShell") ($version -join " | ")
}
finally {
    Pop-Location
    Remove-Item -LiteralPath $root -Recurse -Force -ErrorAction SilentlyContinue
}

if ($failures -gt 0) {
    Write-Error ($failures.ToString() + " PowerShell integration checks failed")
    exit 1
}
Write-Host "All PowerShell integration checks passed."

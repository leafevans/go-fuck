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

    # Rerun initialization defines the alias without starting an interactive logger.
    # Exercise the installed alias with a real binary, without user input.
    $savedPath = $env:PATH
    try {
        $env:PATH = (Split-Path -Parent $Exe) + ';' + $savedPath
        $savedProfile = $PROFILE
        $PROFILE = Join-Path $root 'test profile.ps1'
        Set-Content -LiteralPath $PROFILE -Value 'profile sentinel'
        $profileHash = (Get-FileHash -LiteralPath $PROFILE).Hash
        Invoke-Expression ((& $Exe --alias --capture-mode=rerun) -join "`n")
        Set-Content -LiteralPath 'captured output' -Value 'not a git command' -Encoding ASCII
        $command = "git stats -- 'two words'"
        $reply = fuck --json --no-rerun --output-from 'captured output' --force-command $command | Out-String | ConvertFrom-Json
        Assert 'alias preserves quoted command' ($reply.command -ceq $command) $reply.command
        Assert 'alias preserves spaced input path' ($reply.output -match 'not a git command') $reply.output
        Invoke-Expression ((& $Exe --alias f --capture-mode=rerun) -join "`n")
        Set-Content -LiteralPath 'history-probe.txt' -Value 'short-alias-result'
        $timestamp = [DateTime]::Now
        Add-History -InputObject ([pscustomobject]@{
            Id = [long]1; CommandLine = 'Get-Content history-probe.tx'
            ExecutionStatus = [System.Management.Automation.Runspaces.PipelineState]::Completed
            StartExecutionTime = $timestamp; EndExecutionTime = $timestamp
        })
        $output = @(f --yes)
        Assert 'short alias executes only the corrected history command' ($output.Count -eq 1 -and $output[0] -ceq 'short-alias-result') ($output -join ' | ')
        Assert 'alias leaves profile unchanged' ((Get-FileHash -LiteralPath $PROFILE).Hash -eq $profileHash) $PROFILE
    } finally {
        $PROFILE = $savedProfile
        $env:PATH = $savedPath
    }

    # A real missing-source diagnostic must not pick between tied files.
    Set-Content -LiteralPath 'config.ta' -Value 'candidate A'
    Set-Content -LiteralPath 'config.tb' -Value 'candidate B'
    $hashA = (Get-FileHash -LiteralPath 'config.ta').Hash
    $hashB = (Get-FileHash -LiteralPath 'config.tb').Hash
    $renameCommand = 'Rename-Item config.tx renamed.txt'
    $diagnostic = $null
    try {
        Rename-Item config.tx renamed.txt -ErrorAction Stop
    } catch {
        $diagnostic = $_.Exception.Message
    }
    if (-not $diagnostic) { throw 'Expected a missing-source Rename-Item diagnostic' }
    Set-Content -LiteralPath 'rename.log' -Value $diagnostic -Encoding UTF8
    $reply = & $Exe --json --no-rerun --force-command $renameCommand --output-from 'rename.log' | Out-String | ConvertFrom-Json
    Assert 'tied rename candidates produce no suggestion' ($reply.command -ceq $renameCommand -and $reply.suggestions.Count -eq 0) ($reply | ConvertTo-Json -Depth 4)
    Assert 'preview preserves both candidates' (((Get-FileHash -LiteralPath 'config.ta').Hash -eq $hashA) -and ((Get-FileHash -LiteralPath 'config.tb').Hash -eq $hashB) -and -not (Test-Path -LiteralPath 'renamed.txt')) 'candidate contents or destination changed'
    Remove-Item -LiteralPath 'config.tb'
    $reply = & $Exe --json --no-rerun --force-command $renameCommand --output-from 'rename.log' | Out-String | ConvertFrom-Json
    Assert 'the same diagnostic suggests the unique candidate' ($reply.suggestions[0].command -ceq 'Rename-Item config.ta renamed.txt') ($reply | ConvertTo-Json -Depth 4)

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

    # A history command with embedded quotes has to reach go-fuck verbatim.
    # Windows PowerShell 5.1 mangles quoted arguments passed to native
    # executables, so the alias passes the history through TF_HISTORY. Use a
    # stub for go-fuck so this check does not need an interactive prompt.
    $stubDir = Join-Path $root "stub"
    New-Item -ItemType Directory -Path $stubDir | Out-Null
    $stub = Join-Path $stubDir "go-fuck.cmd"
    Set-Content -LiteralPath $stub -Value @"
@echo off
>> "%GOFUCK_ARGS_FILE%" echo ARGS:%*
>> "%GOFUCK_ARGS_FILE%" echo HISTORY:%TF_HISTORY%
"@
    function Get-History {
        param($Count)
        [pscustomobject]@{ CommandLine = ('git commit -m ' + [char]34 + 'hello world' + [char]34) }
    }
    $previousPath = $env:PATH
    $previousArgsFile = $env:GOFUCK_ARGS_FILE
    $env:PATH = $stubDir + ";" + $previousPath
    $env:GOFUCK_ARGS_FILE = Join-Path $root "stub-args.txt"
    # Match the documented profile initialization, which interpolates output
    # lines into one string and therefore requires statement terminators.
    Invoke-Expression "$(& $Exe --alias --capture-mode=rerun)"
    fuck
    $recorded = Get-Content -LiteralPath $env:GOFUCK_ARGS_FILE -Raw
    $quoted = 'git commit -m ' + [char]34 + 'hello world' + [char]34
    Assert "quoted history is passed verbatim" ($recorded -match [regex]::Escape($quoted)) $recorded
    Remove-Item -Path Function:Get-History -ErrorAction SilentlyContinue
    $env:PATH = $previousPath
    if ($null -eq $previousArgsFile) { Remove-Item Env:GOFUCK_ARGS_FILE -ErrorAction SilentlyContinue }
    else { $env:GOFUCK_ARGS_FILE = $previousArgsFile }

    # A command typed over several lines (a backtick continuation) is one
    # Get-History entry. The value has to reach go-fuck as a whole and the
    # multi line correction has to run afterwards.
    Set-Content -Path "multiline.txt" -Value "multiline-content"
    $env:TF_HISTORY = "Get-Content " + [char]96 + [Environment]::NewLine + "multiline.tx" + [Environment]::NewLine
    $json = & $Exe --json 2>$null | Out-String
    $parsed = $json | ConvertFrom-Json
    $multiSuggestion = if ($null -eq $parsed.suggestions -or $parsed.suggestions.Count -eq 0) { "" } else { $parsed.suggestions[0].command }
    Assert "multi line history is corrected" ($multiSuggestion -match "multiline\.txt") $multiSuggestion
    if ($multiSuggestion -ne "") {
        $multiOutput = Invoke-Expression $multiSuggestion | Out-String
        Assert "multi line correction runs" ($multiOutput -match "multiline-content") $multiOutput
    }
    Remove-Item Env:TF_HISTORY -ErrorAction SilentlyContinue

    # The version output uses the detected shell. cmd.exe merges the two streams
    # without turning the native stderr lines into PowerShell error records.
    $version = cmd /c "`"$Exe`" --version 2>&1"
    Assert "version mentions PowerShell" ($version -match "PowerShell") ($version -join " | ")
    $expectedShell = 'PowerShell ' + $PSVersionTable.PSVersion.ToString()
    $savedVersion = $env:GOFUCK_POWERSHELL_VERSION
    $savedCaller = $env:GOFUCK_SHELL
    $savedErrorAction = $ErrorActionPreference
    try {
        # Redirected version output is on stderr, even on success.
        $ErrorActionPreference = 'Continue'
        Remove-Item Env:GOFUCK_POWERSHELL_VERSION -ErrorAction SilentlyContinue
        $version = (& $Exe -v --debug 2>&1 | ForEach-Object { $_.ToString() }) -join "`n"
        Assert 'legacy caller path probes the current PowerShell' ($LASTEXITCODE -eq 0 -and $version.Contains($expectedShell)) $version

        Remove-Item Env:GOFUCK_SHELL -ErrorAction SilentlyContinue
        $version = (& $Exe -v 2>&1 | ForEach-Object { $_.ToString() }) -join "`n"
        Assert 'direct version works without caller metadata' ($LASTEXITCODE -eq 0 -and $version.Contains('PowerShell')) $version

        $env:PATH = (Split-Path -Parent $Exe) + ';' + $savedPath
        $version = (fuck -v --debug 2>&1 | ForEach-Object { $_.ToString() }) -join "`n"
        Assert 'alias reports the current PowerShell' ($LASTEXITCODE -eq 0 -and $version.Contains($expectedShell)) $version
    } finally {
        $ErrorActionPreference = $savedErrorAction
        $env:PATH = $savedPath
        $env:GOFUCK_POWERSHELL_VERSION = $savedVersion
        $env:GOFUCK_SHELL = $savedCaller
    }
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

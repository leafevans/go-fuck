package shells

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/instant"
	"github.com/leafevans/go-fuck/internal/logs"
)

// powershellAliasTemplate defines the alias function for PowerShell.
//
// Get-History only returns the previous command, so the alias passes it to
// go-fuck for the normal correction flow. Explicit arguments (for example -h
// or --version) must also work when the alias is the first command of a
// session, when there is no history yet, so go-fuck is invoked when either
// history or arguments are available.
const powershellAliasTemplate = `function %[1]s {
    $history = (Get-History -Count 1).CommandLine;
    if (-not [string]::IsNullOrWhiteSpace($history) -or $args.Count -gt 0) {
        $goFuckArgs = @($args);
        if (-not [string]::IsNullOrWhiteSpace($history)) { $goFuckArgs += $history; }
        $fuck = $(%[2]s @goFuckArgs);
        if (-not [string]::IsNullOrWhiteSpace($fuck)) {
            if ($fuck.StartsWith("echo")) { $fuck = $fuck.Substring(5); }
            else { iex "$fuck"; }
        }
    }
    [Console]::ResetColor()
}`

// powershellInstantInitTemplate starts the logged shell of the experimental
// instant mode.
//
// The code is evaluated by the PowerShell profile and replaces the current
// shell with a shell whose terminal output is written to the instant mode log.
// When the logger cannot be started (for example on a Windows version without
// ConPTY), the regular alias is installed instead of closing the terminal.
//
// The profile line evaluates the output with iex "$(...)", which joins the
// output lines with spaces, so every statement ends with a semicolon and the
// code is also valid as a single line.
const powershellInstantInitTemplate = `$env:GOFUCK_INSTANT_MODE = "True";
$env:GOFUCK_OUTPUT_LOG = %[2]s;
$env:GOFUCK_SHELL = (Get-Process -Id $PID).Path;
& %[1]s --shell-logger %[2]s;
$code = $LASTEXITCODE;
if (Test-Path -LiteralPath %[2]s) {
    Remove-Item -LiteralPath %[2]s -ErrorAction SilentlyContinue;
    exit $code;
};
Remove-Item Env:GOFUCK_INSTANT_MODE -ErrorAction SilentlyContinue;
Remove-Item Env:GOFUCK_OUTPUT_LOG -ErrorAction SilentlyContinue;
Remove-Item Env:GOFUCK_SHELL -ErrorAction SilentlyContinue;
iex "$(%[1]s --alias)"`

// powershellInstantMarkedTemplate defines the alias inside the logged shell.
//
// The prompt is wrapped so that every prompt starts with the invisible instant
// mode marker; the log reader uses the marker to split the terminal log into
// command calls. The marker is built from its character codes instead of being
// embedded in the generated code, because the generated code is captured
// through the console and the console encoding is not necessarily UTF-8.
//
// Like the initialization code this template is evaluated as a single line, so
// its statements and blocks are semicolon terminated.
const powershellInstantMarkedTemplate = `$global:__gofuck_mark = %[1]s;
if (-not $global:__gofuck_prompt_wrapped) {
    $global:__gofuck_original_prompt = $function:prompt;
    $global:__gofuck_prompt_wrapped = $true;
};
function global:prompt {
    $text = "";
    if ($null -ne $global:__gofuck_original_prompt) {
        $text = (& $global:__gofuck_original_prompt) -join "";
    };
    return $global:__gofuck_mark + $text;
};
%[2]s`

// Powershell implements the PowerShell specific behavior.
type Powershell struct {
	*Generic
}

// NewPowershell returns a PowerShell shell.
func NewPowershell() *Powershell {
	generic := newGeneric("powershell", "PowerShell")
	generic.versionFn = powershellVersion
	return &Powershell{Generic: generic}
}

// AppAlias returns the PowerShell alias function.
func (p *Powershell) AppAlias(aliasName string) string {
	return fmt.Sprintf(powershellAliasTemplate, aliasName, consts.ExecutableName())
}

// InstantModeAlias returns the alias of the experimental instant mode.
//
// In a normal PowerShell it returns the code that starts the logged shell.
// Inside the logged shell it returns the regular alias together with a prompt
// wrapper that marks the prompts of the shell. Instant mode is only available
// on Windows, because the shell logger needs ConPTY.
func (p *Powershell) InstantModeAlias(aliasName string) string {
	if runtime.GOOS != "windows" || !instant.Supported() {
		logs.Warn("Instant mode not supported by go-fuck")
		return p.AppAlias(aliasName)
	}
	if instantModeEnabled() {
		marked := powerShellStringExpression(consts.UserCommandMark)
		return fmt.Sprintf(powershellInstantMarkedTemplate, marked, p.AppAlias(aliasName))
	}
	logPath := quotePowerShellString(instantModeLogPath())
	return fmt.Sprintf(powershellInstantInitTemplate, consts.ExecutableName(), logPath)
}

// And joins commands with the PowerShell AND operator.
func (p *Powershell) And(commands ...string) string {
	joined := make([]string, len(commands))
	for i, command := range commands {
		joined[i] = "(" + command + ")"
	}
	return strings.Join(joined, " -and ")
}

// HowToConfigure describes how to configure the alias for PowerShell.
func (p *Powershell) HowToConfigure() *Configuration {
	content := "iex " + dquote + "$(" + consts.ExecutableName() + " --alias)" + dquote
	return &Configuration{
		Content:                   content,
		Path:                      "$profile",
		Reload:                    ". $profile",
		CanConfigureAutomatically: false,
	}
}

// powerShellStringExpression returns PowerShell code that evaluates to value.
//
// The string is built from character codes so that the generated code does not
// depend on the encoding of the console that captures it.
func powerShellStringExpression(value string) string {
	runes := []rune(value)
	parts := make([]string, 0, len(runes))
	for index := 0; index < len(runes); {
		end := index + 1
		for end < len(runes) && runes[end] == runes[index] {
			end++
		}
		part := fmt.Sprintf("[string]([char]0x%04X)", runes[index])
		if count := end - index; count > 1 {
			part = "(" + part + " * " + strconv.Itoa(count) + ")"
		}
		parts = append(parts, part)
		index = end
	}
	return strings.Join(parts, " + ")
}

// quotePowerShellString quotes value as a PowerShell single quoted string.
func quotePowerShellString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// powershellVersion returns the PowerShell version.
func powershellVersion() string {
	if output := runOutput("powershell.exe", "$PSVersionTable.PSVersion"); output != "" {
		lines := strings.Split(output, newline)
		return strings.Join(strings.Fields(lines[len(lines)-1]), ".")
	}
	fields := strings.Fields(runOutput("pwsh", "--version"))
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

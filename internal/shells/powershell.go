package shells

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/leafevans/go-fuck/v3/internal/consts"
	"github.com/leafevans/go-fuck/v3/internal/instant"
	"github.com/leafevans/go-fuck/v3/internal/logs"
)

// powershellAliasTemplate defines the alias function for PowerShell.
//
// Get-History only returns the previous command, so the alias passes it to
// go-fuck for the normal correction flow. The history is passed through the
// TF_HISTORY environment variable instead of a command line argument:
// Windows PowerShell 5.1 does not escape embedded quotes when it passes an
// argument to a native executable, so a command such as
// git commit -m "hello world" would reach go-fuck mangled. Environment values
// are transferred verbatim. Explicit arguments (for example -h or --version)
// must also work when the alias is the first command of a session, when there
// is no history yet, so go-fuck is invoked when either history or arguments
// are available. Without either, pass the empty history marker so the CLI
// can explain why there is nothing to correct.
const powershellAliasTemplate = `function %[1]s {
%[3]s    $env:TF_SHELL = "powershell";
    $env:TF_ALIAS = "%[1]s";
    $env:TF_ALIAS_VERSION = "%[4]s";
    try { $env:GOFUCK_SHELL = (Get-Process -Id $PID).Path } catch { };
    $history = @(
        Get-History -Count 8 |
            ForEach-Object { $_.CommandLine } |
            Where-Object {
                -not [string]::IsNullOrWhiteSpace($_) -and
                (($_.TrimStart() -split '\s+', 2)[0] -ne "%[1]s")
            } |
            Select-Object -Last 1
    );
    $history = [string]::Join([Environment]::NewLine, $history);
    $previousHistory = $env:TF_HISTORY;
    try {
        if (-not [string]::IsNullOrWhiteSpace($history)) { $env:TF_HISTORY = $history; }
        elseif ($args.Count -eq 0) { $env:TF_HISTORY = [string][char]1; }
        else { Remove-Item Env:TF_HISTORY -ErrorAction SilentlyContinue; }
        $fuck = ($(%[2]s @args) -join [Environment]::NewLine);
        if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($fuck)) {
            iex "$fuck";
        }
    } finally {
        if ($null -eq $previousHistory) { Remove-Item Env:TF_HISTORY -ErrorAction SilentlyContinue }
        else { $env:TF_HISTORY = $previousHistory }
    }
    [Console]::ResetColor()
}`

// powershellInstantInitTemplate starts the logged shell of the
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
$env:TF_ALIAS = %[3]s;
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
	return fmt.Sprintf(powershellAliasTemplate, aliasName, consts.ExecutableName(),
		powershellAliasGuard(aliasName, consts.ExecutableName()), consts.AliasVersion)
}

// InstantModeAlias returns the alias of the instant mode.
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
	return fmt.Sprintf(powershellInstantInitTemplate, consts.ExecutableName(), logPath,
		quotePowerShellString(aliasName))
}

// And joins commands so that the next command runs only after the previous one
// succeeded.
//
// PowerShell had no command separator with AND semantics before version 7, and
// the logical -and operator evaluates the truthiness of a command output
// instead of its exit code, so every command is guarded by $? instead.
func (p *Powershell) And(commands ...string) string {
	return powershellChain(commands, "if ($?) { %s }")
}

// Or joins commands so that the next command runs only after the previous one
// failed.
func (p *Powershell) Or(commands ...string) string {
	return powershellChain(commands, "if (-not $?) { %s }")
}

// powershellChain joins commands with the given guard template, nesting the
// guards from the last command to the first one.
func powershellChain(commands []string, template string) string {
	if len(commands) == 0 {
		return ""
	}
	joined := commands[len(commands)-1]
	for index := len(commands) - 2; index >= 0; index-- {
		joined = commands[index] + "; " + fmt.Sprintf(template, joined)
	}
	return joined
}

// HowToConfigure describes how to configure the alias for PowerShell.
func (p *Powershell) HowToConfigure() *Configuration {
	content := "iex " + dquote + "$(" + consts.ExecutableName() + " --alias)" + dquote
	configuration := &Configuration{
		Content:                   content,
		Path:                      "$profile",
		Reload:                    ". $profile",
		CanConfigureAutomatically: false,
	}
	// Instant mode needs ConPTY, so it is only available in Windows PowerShell.
	if runtime.GOOS == "windows" && instant.Supported() {
		configuration.InstantContent = "iex " + dquote + "$(" + consts.ExecutableName() +
			" --alias --enable-experimental-instant-mode)" + dquote
	}
	return configuration
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

// Quote returns a PowerShell escaped version of s.
//
// PowerShell does not concatenate adjacent string literals like POSIX shells
// do, so the shlex style '"'"' escape is a syntax error there. Single quoted
// strings are literal apart from the doubled single quote, which makes them the
// safe form for every value.
func (p *Powershell) Quote(s string) string {
	return quotePowerShellString(s)
}

// SplitCommand splits a command line into the arguments that PowerShell passes
// to the command.
//
// Backslash is an ordinary character in PowerShell (the backtick is the escape
// character), so a Windows path must not be treated as a POSIX escape
// sequence: "cat C:\Windows" has to keep the backslashes.
func (p *Powershell) SplitCommand(command string) []string {
	return splitPowerShell(command)
}

// splitPowerShell splits a command line using PowerShell quoting rules.
//
// Only the parts that matter for rule matching are implemented: argument
// separation, single and double quoted strings, doubled single quotes, and
// backtick escapes. Variables and subexpressions are kept as literal text.
func splitPowerShell(command string) []string {
	var (
		args    []string
		current strings.Builder
		inWord  bool
	)
	flush := func() {
		if inWord {
			args = append(args, current.String())
			current.Reset()
			inWord = false
		}
	}
	runes := []rune(command)
	for index := 0; index < len(runes); index++ {
		ch := runes[index]
		switch {
		case isArgumentSpace(ch):
			flush()
		case ch == rune(charSingleQuote):
			inWord = true
			for index++; index < len(runes); index++ {
				if runes[index] != rune(charSingleQuote) {
					current.WriteRune(runes[index])
					continue
				}
				// Two consecutive single quotes are one literal quote.
				if index+1 < len(runes) && runes[index+1] == rune(charSingleQuote) {
					current.WriteRune(rune(charSingleQuote))
					index++
					continue
				}
				break
			}
		case ch == rune(charDoubleQuote):
			inWord = true
			for index++; index < len(runes); index++ {
				next := runes[index]
				if next == rune(charBacktick) {
					if index+1 < len(runes) {
						current.WriteRune(runes[index+1])
						index++
					}
					continue
				}
				if next == rune(charDoubleQuote) {
					break
				}
				current.WriteRune(next)
			}
		case ch == rune(charBacktick):
			inWord = true
			if index+1 < len(runes) {
				current.WriteRune(runes[index+1])
				index++
			}
		default:
			inWord = true
			current.WriteRune(ch)
		}
	}
	flush()
	return args
}

// quotePowerShellString quotes value as a PowerShell single quoted string.
func quotePowerShellString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// powershellVersion trusts the calling alias rather than guessing an edition
// from PATH. Older aliases report only the executable path.
func powershellVersion() string {
	if version := strings.TrimSpace(env(consts.PowerShellVersionEnv)); version != "" {
		return version
	}
	executable := env(consts.ShellPathEnv)
	if !filepath.IsAbs(executable) {
		return ""
	}
	switch strings.ToLower(filepath.Base(executable)) {
	case "powershell.exe", "pwsh.exe", "pwsh":
		// Loading the user's profile can start another go-fuck shell logger.
		return runPowershellVersionProbe(executable, "-NoLogo", "-NoProfile", "-NonInteractive",
			"-Command", "$PSVersionTable.PSVersion.ToString()")
	default:
		return ""
	}
}

var runPowershellVersionProbe = func(name string, args ...string) string {
	return powershellProbeOutput(2*time.Second, name, args...)
}

// powershellProbeOutput runs a short-lived PowerShell probe with a timeout.
func powershellProbeOutput(timeout time.Duration, name string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	// A descendant may keep the output pipe open after PowerShell exits.
	cmd.WaitDelay = 250 * time.Millisecond
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

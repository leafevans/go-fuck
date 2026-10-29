package shells

import (
	"fmt"
	"strings"

	"github.com/leafevans/go-fuck/internal/consts"
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
	return fmt.Sprintf(powershellAliasTemplate, aliasName, consts.AppName)
}

// InstantModeAlias falls back to the regular alias because instant mode is not
// supported.
func (p *Powershell) InstantModeAlias(aliasName string) string {
	logs.Warn("Instant mode not supported by go-fuck")
	return p.AppAlias(aliasName)
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
	content := "iex " + dquote + "$(" + consts.AppName + " --alias)" + dquote
	return &Configuration{
		Content:                   content,
		Path:                      "$profile",
		Reload:                    ". $profile",
		CanConfigureAutomatically: false,
	}
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

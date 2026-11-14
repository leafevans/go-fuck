package shells

import (
	"fmt"
	"strings"
)

// informationalCommands lists the go-fuck flags that print information instead
// of a corrected command.
//
// The shell alias evaluates the standard output of go-fuck as the corrected
// command. For these flags the output is a rule list, a JSON report or a
// profile snippet, so evaluating it would execute every line as a command (and
// the rule list even contains names such as "java" or "history" that are real
// commands). The alias therefore runs go-fuck directly for them and leaves the
// output on the terminal.
var informationalCommands = []string{
	"-h", "--help",
	"-v", "--version",
	"--doctor",
	"--list-rules",
	"--json",
	"--setup",
	"--alias", "-a",
	"-l", "--shell-logger",
}

// informationalCommandPattern returns the flags joined for a case pattern.
func informationalCommandPattern() string {
	return strings.Join(informationalCommands, "|")
}

// informationalCommandList returns the flags as a space separated list.
func informationalCommandList() string {
	return strings.Join(informationalCommands, " ")
}

// posixAliasGuard returns the bash and zsh code that passes the informational
// commands through to go-fuck.
func posixAliasGuard(shellName, aliasName, executable string) string {
	return fmt.Sprintf("    for __gofuck_option in \"$@\"; do\n"+
		"        case $__gofuck_option in\n"+
		"            %[1]s)\n"+
		"                export TF_SHELL=%[2]s;\n"+
		"                export TF_ALIAS=%[3]s;\n"+
		"                %[4]s \"$@\";\n"+
		"                return;;\n"+
		"        esac\n"+
		"    done\n",
		informationalCommandPattern(), shellName, aliasName, executable)
}

// fishAliasGuard returns the fish code that passes the informational commands
// through to go-fuck.
func fishAliasGuard(aliasName, executable string) string {
	return fmt.Sprintf("  for __gofuck_option in $argv\n"+
		"    switch $__gofuck_option\n"+
		"      case %[1]s\n"+
		"        env TF_SHELL=fish TF_ALIAS=%[2]s %[3]s $argv\n"+
		"        return\n"+
		"    end\n"+
		"  end\n",
		informationalCommandList(), aliasName, executable)
}

// powershellAliasGuard returns the PowerShell code that passes the
// informational commands through to go-fuck.
func powershellAliasGuard(aliasName, executable string) string {
	quoted := make([]string, 0, len(informationalCommands))
	for _, flag := range informationalCommands {
		quoted = append(quoted, "'"+flag+"'")
	}
	return fmt.Sprintf("    $__gofuck_information = @(%[1]s)\n"+
		"    foreach ($__gofuck_option in $args) {\n"+
		"        if ($__gofuck_information -contains $__gofuck_option) {\n"+
		"            $env:TF_SHELL = \"powershell\";\n"+
		"            $env:TF_ALIAS = \"%[2]s\";\n"+
		"            $env:TF_ALIAS_VERSION = \"3\";\n"+
		"            try { $env:GOFUCK_SHELL = (Get-Process -Id $PID).Path } catch { };\n"+
		"            & %[3]s @args;\n"+
		"            return\n"+
		"        }\n"+
		"    }\n",
		strings.Join(quoted, ", "), aliasName, executable)
}

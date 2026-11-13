package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// powerShellUnknownParameter matches the error for a parameter that does not
// exist, for example: A parameter cannot be found that matches parameter name 'rf'.
var powerShellUnknownParameter = regexp.MustCompile(`A parameter cannot be found that matches parameter name '([^']*)'\.`)

// powerShellAmbiguousParameter matches the error for an abbreviated parameter
// that matches several parameters, for example: Parameter cannot be processed
// because the parameter name 'f' is ambiguous.
var powerShellAmbiguousParameter = regexp.MustCompile(`the parameter name '([^']*)' is ambiguous`)

// powerShellFlagFamilies maps a command family to the Unix style flags that it
// can translate. The keys are the PowerShell commands that accept the
// PowerShell parameters; the values map a Unix letter to a PowerShell
// parameter.
var powerShellFlagFamilies = []struct {
	commands []string
	flags    map[rune]string
}{
	{
		commands: []string{"rm", "del", "erase", "ri", "remove-item"},
		flags:    map[rune]string{'r': "-Recurse", 'f': "-Force"},
	},
	{
		commands: []string{"cp", "copy", "cpi", "copy-item"},
		flags:    map[rune]string{'r': "-Recurse", 'f': "-Force"},
	},
	{
		commands: []string{"mv", "move", "mi", "move-item"},
		flags:    map[rune]string{'f': "-Force"},
	},
}

// powerShellUnixFlagsMatch reports whether a Unix style flag was rejected by a
// PowerShell command.
func powerShellUnixFlagsMatch(command *types.Command) bool {
	_, ok := powerShellUnixFlagsReplacement(command)
	return ok
}

// powerShellUnixFlagsReplacement returns the parameters that replace the
// rejected flags and whether the command can be corrected.
func powerShellUnixFlagsReplacement(command *types.Command) (string, bool) {
	name, ok := powerShellRejectedParameter(command.Output)
	if !ok || name == "" {
		return "", false
	}
	parts := command.ScriptParts()
	if len(parts) == 0 || !utils.Contains(parts, "-"+name) {
		return "", false
	}
	flags := map[rune]string{}
	matched := false
	for _, family := range powerShellFlagFamilies {
		if !containsFold(family.commands, parts[0]) {
			continue
		}
		flags = family.flags
		matched = true
		break
	}
	if !matched {
		return "", false
	}
	parameters := make([]string, 0, len(name))
	for _, letter := range strings.ToLower(name) {
		parameter, ok := flags[letter]
		if !ok {
			// A flag without a PowerShell equivalent is left alone: a partial
			// translation would silently change the command.
			return "", false
		}
		parameters = append(parameters, parameter)
	}
	return strings.Join(parameters, " "), true
}

// powerShellRejectedParameter returns the parameter name from an unknown or
// ambiguous parameter error.
func powerShellRejectedParameter(commandOutput string) (string, bool) {
	if match := powerShellUnknownParameter.FindStringSubmatch(commandOutput); match != nil {
		return match[1], true
	}
	if match := powerShellAmbiguousParameter.FindStringSubmatch(commandOutput); match != nil {
		return match[1], true
	}
	return "", false
}

// containsFold reports whether values contains value, ignoring case.
func containsFold(values []string, value string) bool {
	for _, candidate := range values {
		if strings.EqualFold(candidate, value) {
			return true
		}
	}
	return false
}

// powerShellUnixFlagsNewCommand replaces the rejected Unix style flag with the
// equivalent PowerShell parameters.
func powerShellUnixFlagsNewCommand(command *types.Command) ([]string, error) {
	parameters, ok := powerShellUnixFlagsReplacement(command)
	if !ok {
		return nil, nil
	}
	name, _ := powerShellRejectedParameter(command.Output)
	replaced := utils.ReplaceArgument(command.Script, "-"+name, parameters)
	if replaced == command.Script {
		return nil, nil
	}
	return []string{replaced}, nil
}

func init() {
	register(types.NewRule(
		"powershell_unix_flags",
		powerShellUnixFlagsMatch,
		powerShellUnixFlagsNewCommand,
	))
}

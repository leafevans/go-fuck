package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// dotnetUnknownCommandToken matches the ASCII `dotnet-<command>` token that
	// the .NET CLI prints for a misspelled top level command, for example
	// `dotnet-buidl does not exist`. The surrounding message is localized, but
	// the token always contains the command the user typed.
	dotnetUnknownCommandToken = regexp.MustCompile(`dotnet-([A-Za-z0-9][A-Za-z0-9_.-]*)`)
	// dotnetUnknownCommandHelpLine matches one command of the "dotnet --help"
	// listing, for example `  build             Build a .NET project.`. The
	// command names are not localized.
	dotnetUnknownCommandHelpLine = regexp.MustCompile(`(?m)^  ([a-z][a-z0-9-]*) {2,}\S`)
)

// dotnetUnknownCommandFallbackCommands is used when "dotnet --help" cannot be
// inspected. It mirrors the SDK and bundled tool commands of the .NET SDK.
var dotnetUnknownCommandFallbackCommands = []string{
	"build", "build-server", "clean", "format", "help", "msbuild", "new",
	"nuget", "pack", "package", "publish", "reference", "restore", "run",
	"sdk", "solution", "store", "test", "tool", "vstest", "workload",
	"dev-certs", "fsi", "user-jwts", "user-secrets", "watch",
}

// dotnetUnknownCommandReadHelp runs "dotnet --help" and returns its output.
var dotnetUnknownCommandReadHelp = func() string {
	output, err := exec.Command("dotnet", "--help").CombinedOutput()
	if err != nil && len(output) == 0 {
		return ""
	}
	return string(output)
}

// dotnetUnknownCommandParseCommands extracts the command names from the
// "dotnet --help" output.
func dotnetUnknownCommandParseCommands(output string) []string {
	commands := []string{}
	seen := map[string]bool{}
	for _, line := range strings.Split(output, newline) {
		match := dotnetUnknownCommandHelpLine.FindStringSubmatch(line)
		if match == nil || seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		commands = append(commands, match[1])
	}
	return commands
}

// dotnetUnknownCommandCommands returns the commands listed by "dotnet --help".
var dotnetUnknownCommandCommands = utils.Memoize(func() []string {
	commands := dotnetUnknownCommandParseCommands(dotnetUnknownCommandReadHelp())
	if len(commands) == 0 {
		return append([]string{}, dotnetUnknownCommandFallbackCommands...)
	}
	return commands
})

// dotnetUnknownCommandBroken returns the misspelled top level command that the
// .NET CLI echoed in its output, or an empty string.
func dotnetUnknownCommandBroken(command *types.Command) string {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return ""
	}
	for _, match := range dotnetUnknownCommandToken.FindAllStringSubmatch(command.Output, -1) {
		for _, part := range parts[1:] {
			if part == match[1] {
				return match[1]
			}
		}
	}
	return ""
}

// dotnetUnknownCommandMatch reports whether the .NET CLI did not recognize a
// top level command and a similar built-in command exists.
func dotnetUnknownCommandMatch(command *types.Command) bool {
	return dotnetUnknownCommandClosest(command) != ""
}

// dotnetUnknownCommandClosest returns the closest built-in .NET CLI command.
func dotnetUnknownCommandClosest(command *types.Command) string {
	broken := dotnetUnknownCommandBroken(command)
	if broken == "" {
		return ""
	}
	return utils.GetClosest(broken, dotnetUnknownCommandCommands(), 0.6, false)
}

// dotnetUnknownCommandNewCommand replaces the misspelled command with the
// closest built-in .NET CLI command.
func dotnetUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	closest := dotnetUnknownCommandClosest(command)
	if closest == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, dotnetUnknownCommandBroken(command), closest)}, nil
}

func init() {
	register(types.NewRule(
		"dotnet_unknown_command",
		utils.ForApp([]string{"dotnet"}, 0, dotnetUnknownCommandMatch),
		dotnetUnknownCommandNewCommand,
	))
}

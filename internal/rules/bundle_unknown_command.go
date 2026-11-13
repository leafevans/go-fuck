package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// bundleUnknownCommandFallbackCommands lists the commands of Bundler 4. It is
// used when the local Bundler installation cannot be inspected.
var bundleUnknownCommandFallbackCommands = []string{
	"add", "binstubs", "cache", "check", "clean", "config", "console",
	"doctor", "env", "exec", "fund", "gem", "help", "info", "init",
	"install", "issue", "licenses", "list", "lock", "open", "outdated",
	"package", "pack", "platform", "plugin", "pristine", "remove", "show",
	"update", "version",
}

// bundleUnknownCommandRegexp matches the command that Bundler did not find,
// for example: Could not find command "zzz".
var bundleUnknownCommandRegexp = regexp.MustCompile(`Could not find command "([^"]*)"\.`)

// bundleUnknownCommandReferenceRegexp matches a backticked command reference
// of the help output, for example "bundle install(1)" or the alias
// "bundle package".
var bundleUnknownCommandReferenceRegexp = regexp.MustCompile(`\x60bundle ([a-zA-Z0-9_-]+)(?:\([0-9]\))?\x60`)

// bundleUnknownCommandParseCommands parses the command names of the
// "bundle help" output.
//
// When no system man page viewer is available, Bundler prints the raw ronn
// source of the bundle(1) manual page. Every command is listed as a backticked
// "bundle <command>(1)" reference, while aliases are listed as backticked
// "bundle <alias>" references next to the command they alias, for example
// "(aliases: bundle package, bundle pack)". The result is deduplicated and
// keeps the order of the help output.
func bundleUnknownCommandParseCommands(output string) []string {
	commands := []string{}
	seen := make(map[string]bool)
	for _, match := range bundleUnknownCommandReferenceRegexp.FindAllStringSubmatch(output, -1) {
		name := match[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		commands = append(commands, name)
	}
	return commands
}

// bundleUnknownCommandHelp returns the commands reported by "bundle help".
var bundleUnknownCommandHelp = utils.Memoize(func() []string {
	output, err := exec.Command("bundle", "help").Output()
	if err != nil {
		return append([]string{}, bundleUnknownCommandFallbackCommands...)
	}
	commands := bundleUnknownCommandParseCommands(string(output))
	if len(commands) == 0 {
		// A system man page viewer renders the help output and hides the
		// backticked command references, so fall back to the static list.
		return append([]string{}, bundleUnknownCommandFallbackCommands...)
	}
	return commands
})

// bundleUnknownCommandMatch reports whether Bundler did not recognize the
// command.
func bundleUnknownCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Could not find command ") &&
		bundleUnknownCommandRegexp.MatchString(command.Output)
}

// bundleUnknownCommandNewCommand suggests the closest valid Bundler commands.
func bundleUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	match := bundleUnknownCommandRegexp.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	closest := utils.GetCloseMatches(match[1], bundleUnknownCommandHelp(), 0, 0.6)
	return utils.ReplaceArguments(command.Script, match[1], closest), nil
}

func init() {
	register(types.NewRule(
		"bundle_unknown_command",
		utils.ForApp([]string{"bundle"}, 0, bundleUnknownCommandMatch),
		bundleUnknownCommandNewCommand,
	))
}

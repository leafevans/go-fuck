package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// cobraUnknownCommand matches the error line printed by command line tools that
// use the Cobra framework, for example `unknown command "prr" for "gh"`.
var cobraUnknownCommand = newRegexp(`unknown command "([^"]*)" for "([^"]*)"`)

// cobraSuggestions extracts the broken command and the "Did you mean this?"
// suggestions from a Cobra error message.
//
// Cobra prints the suggestions as tab indented lines directly after the
// separator, followed by the usage text, for example:
//
//	unknown command "prr" for "gh"
//
//	Did you mean this?
//		org
//		pr
//	Usage:  gh <command> <subcommand> [flags]
func cobraSuggestions(commandOutput string) (broken string, suggestions []string) {
	match := cobraUnknownCommand.FindStringSubmatch(commandOutput)
	if match == nil {
		return "", nil
	}
	broken = match[1]

	inSuggestions := false
	for line := range strings.SplitSeq(commandOutput, newline) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Did you mean this?" {
			inSuggestions = true
			continue
		}
		if !inSuggestions {
			continue
		}
		if strings.HasPrefix(line, "	") {
			suggestions = append(suggestions, trimmed)
			continue
		}
		if trimmed == "" {
			continue
		}
		break
	}
	return broken, suggestions
}

// cobraUnknownCommandMatcher returns the match function of one Cobra
// application. The application name is matched as a prefix of the command that
// Cobra reports, so nested commands such as "gh repo" are covered too.
func cobraUnknownCommandMatcher(app string) func(*types.Command) bool {
	return func(command *types.Command) bool {
		output := command.Output
		return strings.Contains(output, `unknown command "`) &&
			strings.Contains(output, ` for "`+app) &&
			strings.Contains(output, "Did you mean this?")
	}
}

// cobraUnknownCommandNewCommand replaces the unknown command with every
// suggestion printed by Cobra.
func cobraUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	broken, suggestions := cobraSuggestions(command.Output)
	if broken == "" {
		return nil, nil
	}
	return utils.ReplaceArgumentsSorted(command.Script, broken, suggestions), nil
}

var cobraSubcommandDiagnostic = newRegexp(`^(?:Error: |error: )?unknown command "([^"]+)" for "([^"]+)"$`)

// registerScopedCobraRule uses the reported parent path, rather than replacing
// another occurrence of the typo elsewhere in the invocation.
func registerScopedCobraRule(name, root string, apps []string, options func(string) commandOption) {
	correct := func(command *types.Command) []string {
		if !strings.Contains(command.Output, "Did you mean") || !strings.Contains(command.Output, "unknown command") {
			return nil
		}
		var path []string
		var broken string
		var suggestions []string
		inSuggestions := false
		for line := range strings.SplitSeq(command.Output, "\n") {
			line = strings.TrimSuffix(line, "\r")
			if match := cobraSubcommandDiagnostic.FindStringSubmatch(line); match != nil {
				if broken != "" {
					return nil
				}
				path = strings.Fields(match[2])
				if len(path) == 0 || path[0] != root {
					return nil
				}
				broken = match[1]
				continue
			}
			if broken == "" || strings.TrimSpace(line) == "" {
				continue
			}
			if !inSuggestions {
				if line != "Did you mean this?" && line != "Did you mean one of these?" {
					break
				}
				inSuggestions = true
				continue
			}
			hint := strings.TrimSpace(line)
			if !strings.HasPrefix(line, "\t") || !cliSubcommandName.MatchString(hint) {
				break
			}
			suggestions = append(suggestions, hint)
		}
		if len(suggestions) == 0 || !cliSubcommandName.MatchString(broken) {
			return nil
		}
		return cliSubcommandCorrections(command, apps, path[1:], broken, suggestions, options)
	}
	register(types.NewRule(name, func(command *types.Command) bool { return len(correct(command)) != 0 },
		func(command *types.Command) ([]string, error) { return correct(command), nil }))
}

// newCobraUnknownCommandRule registers the standard unknown command rule of a
// Cobra application. ruleName is the rule name and app the executable name that
// appears in the error. apps optionally lists additional executable names used
// to decide which command failed.
func newCobraUnknownCommandRule(ruleName, app string, apps ...string) {
	if len(apps) == 0 {
		apps = []string{app}
	}
	register(types.NewRule(
		ruleName,
		utils.ForApp(apps, 0, cobraUnknownCommandMatcher(app)),
		cobraUnknownCommandNewCommand,
	))
}

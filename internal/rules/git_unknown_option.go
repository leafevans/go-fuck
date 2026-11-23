package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// gitUnknownOption matches the error line for an unknown long option, for
// example: error: unknown option `set-upstrem'.
var gitUnknownOption = newRegexp("error: unknown option `([^']*)'")

// gitUsageOption matches a long option in the usage list of git. Boolean
// options are printed as --[no-]option, so the optional prefix is skipped.
var gitUsageOption = newRegexp(`(?m)^\s+(?:-[A-Za-z],\s+)?--(?:\[no-\])?([A-Za-z0-9][A-Za-z0-9-]*)`)

// gitUnknownOptionBroken returns the long option git rejected, or an empty
// string when the output does not name one.
func gitUnknownOptionBroken(commandOutput string) string {
	if match := gitUnknownOption.FindStringSubmatch(commandOutput); match != nil {
		return match[1]
	}
	return ""
}

// gitUnknownOptionCandidates returns the unique long options printed in the
// usage list of git, in their first appearance order.
func gitUnknownOptionCandidates(commandOutput string) []string {
	seen := map[string]bool{}
	options := []string{}
	for _, match := range gitUsageOption.FindAllStringSubmatch(commandOutput, -1) {
		option := "--" + match[1]
		if seen[option] {
			continue
		}
		seen[option] = true
		options = append(options, option)
	}
	return options
}

// gitUnknownOptionMatch reports whether git rejected a long option that is part
// of the command line and printed the usage list to correct it.
func gitUnknownOptionMatch(command *types.Command) bool {
	_, _, _, ok := gitUnknownOptionCorrection(command)
	return ok
}

func gitUnknownOptionCorrection(command *types.Command) (start, end int, option string, ok bool) {
	broken := gitUnknownOptionBroken(command.Output)
	if broken == "" {
		return 0, 0, "", false
	}
	broken = "--" + broken
	start, end, ok = specific.GitOptionSpan(command.Script, broken, shells.Current())
	if !ok {
		return 0, 0, "", false
	}
	option = utils.GetClosest(broken, gitUnknownOptionCandidates(command.Output), 0.8, false)
	return start, end, option, option != "" && option != broken
}

// gitUnknownOptionNewCommand replaces the rejected option with the closest
// valid option from the usage list.
func gitUnknownOptionNewCommand(command *types.Command) ([]string, error) {
	start, end, option, ok := gitUnknownOptionCorrection(command)
	if !ok {
		return nil, nil
	}
	return []string{command.Script[:start] + option + command.Script[end:]}, nil
}

func init() {
	register(types.NewRule(
		"git_unknown_option",
		specific.GitSupportMatch(gitUnknownOptionMatch),
		specific.GitSupportNew(gitUnknownOptionNewCommand),
	))
}

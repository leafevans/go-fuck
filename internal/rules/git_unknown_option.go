package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitUnknownOption matches the error line for an unknown long option, for
// example: error: unknown option `set-upstrem'.
var gitUnknownOption = regexp.MustCompile("error: unknown option `([^']*)'")

// gitUsageOption matches a long option in the usage list of git. Boolean
// options are printed as --[no-]option, so the optional prefix is skipped.
var gitUsageOption = regexp.MustCompile(`(?m)^\s+(?:-[A-Za-z],\s+)?--(?:\[no-\])?([A-Za-z0-9][A-Za-z0-9-]*)`)

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
	broken := gitUnknownOptionBroken(command.Output)
	if broken == "" {
		return false
	}
	return utils.Contains(command.ScriptParts(), "--"+broken) &&
		len(gitUnknownOptionCandidates(command.Output)) > 0
}

// gitUnknownOptionNewCommand replaces the rejected option with the closest
// valid option from the usage list.
func gitUnknownOptionNewCommand(command *types.Command) ([]string, error) {
	broken := gitUnknownOptionBroken(command.Output)
	closest := utils.GetClosest("--"+broken, gitUnknownOptionCandidates(command.Output), 0.8, false)
	if closest == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, "--"+broken, closest)}, nil
}

func init() {
	register(types.NewRule(
		"git_unknown_option",
		specific.GitSupportMatch(gitUnknownOptionMatch),
		specific.GitSupportNew(gitUnknownOptionNewCommand),
	))
}

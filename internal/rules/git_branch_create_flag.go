package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

var gitBranchCreateFlagError = newRegexp("(?m)^error: unknown switch `([bc])'\r?$")

func gitBranchCreateFlagCorrection(command *types.Command) (start, end int, replacement string, ok bool) {
	match := gitBranchCreateFlagError.FindStringSubmatch(command.Output)
	if match == nil {
		return 0, 0, "", false
	}
	broken := "-" + match[1]
	subcommand, replacement := "switch", "-c"
	if broken == "-c" {
		subcommand, replacement = "checkout", "-b"
	}
	shell := shells.Current()
	seenSubcommand, seenFlag, seenBranch := false, false, false
	for _, word := range specific.GitCommandPathWords(command.Script, shell) {
		switch {
		case !seenSubcommand:
			if word != subcommand {
				return 0, 0, "", false
			}
			seenSubcommand = true
		case !seenFlag:
			// Only skip known valueless options, so we never edit an option value.
			if word == "-q" || word == "--quiet" {
				continue
			}
			if word != broken {
				return 0, 0, "", false
			}
			seenFlag = true
		case !seenBranch:
			if word == "" || strings.HasPrefix(word, "-") {
				return 0, 0, "", false
			}
			seenBranch = true
		case word == "--" || word == replacement:
			return 0, 0, "", false
		}
	}
	if !seenBranch {
		return 0, 0, "", false
	}
	start, end, ok = specific.GitOptionSpan(command.Script, broken, shell)
	return start, end, replacement, ok
}

func gitBranchCreateFlagMatch(command *types.Command) bool {
	_, _, _, ok := gitBranchCreateFlagCorrection(command)
	return ok
}

func gitBranchCreateFlagNewCommand(command *types.Command) ([]string, error) {
	start, end, replacement, ok := gitBranchCreateFlagCorrection(command)
	if !ok {
		return nil, nil
	}
	return []string{command.Script[:start] + replacement + command.Script[end:]}, nil
}

func init() {
	register(types.NewRule("git_branch_create_flag",
		specific.GitSupportMatch(gitBranchCreateFlagMatch),
		specific.GitSupportNew(gitBranchCreateFlagNewCommand)))
}

package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// gitFlagAfterFilenameFlagPattern matches a flag that was put after a
// filename.
var gitFlagAfterFilenameFlagPattern = newRegexp(
	"fatal: bad flag '(.*?)' used after filename")

// gitFlagAfterFilenameOptionPattern matches an option that git requires before
// non-option arguments.
var gitFlagAfterFilenameOptionPattern = newRegexp(
	"fatal: option '(.*?)' must come before non-option arguments")

// gitFlagAfterFilenameBadFlag returns the flag that git rejected, or an empty
// string.
func gitFlagAfterFilenameBadFlag(commandOutput string) string {
	if match := gitFlagAfterFilenameFlagPattern.FindStringSubmatch(commandOutput); match != nil {
		return match[1]
	}
	if match := gitFlagAfterFilenameOptionPattern.FindStringSubmatch(commandOutput); match != nil {
		return match[1]
	}
	return ""
}

// gitFlagAfterFilenameMatch reports whether only the flag position is wrong.
func gitFlagAfterFilenameMatch(command *types.Command) bool {
	corrected, _ := gitFlagAfterFilenameNewCommand(command)
	return len(corrected) != 0
}

// gitFlagAfterFilenameNewCommand moves the flag before the filename.
func gitFlagAfterFilenameNewCommand(command *types.Command) ([]string, error) {
	badFlag := gitFlagAfterFilenameBadFlag(command.Output)
	// These switches do not take a separate value. Moving an arbitrary option
	// could detach it from its value and change the meaning of the command.
	if !gitDiffStandaloneOption(badFlag) {
		return nil, nil
	}

	start, end, ok := specific.GitOptionSpan(command.Script, badFlag, shells.Current())
	if !ok {
		return nil, nil
	}
	var subcommandEnd int
	hasOperand := false
	for span, word := range specific.GitCommandWords(command.Script, shells.Current()) {
		if span[0] >= start {
			break
		}
		if subcommandEnd == 0 {
			subcommandEnd = span[1]
		} else if !strings.HasPrefix(word, "-") {
			hasOperand = true
		}
	}
	if !hasOperand {
		return nil, nil
	}
	beforeFlag := strings.TrimRight(command.Script[subcommandEnd:start], " \t")
	return []string{command.Script[:subcommandEnd] + " " + command.Script[start:end] + beforeFlag + command.Script[end:]}, nil
}

func init() {
	register(types.NewRule("git_flag_after_filename",
		specific.GitSupportMatch(gitFlagAfterFilenameMatch),
		specific.GitSupportNew(gitFlagAfterFilenameNewCommand)))
}

package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitFlagAfterFilenameFlagPattern matches a flag that was put after a
// filename.
var gitFlagAfterFilenameFlagPattern = regexp.MustCompile(
	"fatal: bad flag '(.*?)' used after filename")

// gitFlagAfterFilenameOptionPattern matches an option that git requires before
// non-option arguments.
var gitFlagAfterFilenameOptionPattern = regexp.MustCompile(
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
	return gitFlagAfterFilenameBadFlag(command.Output) != ""
}

// gitFlagAfterFilenameNewCommand moves the flag before the filename.
func gitFlagAfterFilenameNewCommand(command *types.Command) ([]string, error) {
	badFlag := gitFlagAfterFilenameBadFlag(command.Output)
	if badFlag == "" {
		return nil, nil
	}

	parts := append([]string{}, command.ScriptParts()...)
	badFlagIndex := utils.Index(parts, badFlag)
	if badFlagIndex < 0 {
		return nil, nil
	}

	filenameIndex := -1
	for index := badFlagIndex - 1; index >= 0; index-- {
		if len(parts[index]) > 0 && parts[index][0] != '-' {
			filenameIndex = index
			break
		}
	}
	if filenameIndex < 0 {
		return nil, nil
	}

	parts[badFlagIndex], parts[filenameIndex] = parts[filenameIndex], parts[badFlagIndex]
	return []string{strings.Join(parts, " ")}, nil
}

func init() {
	register(types.NewRule("git_flag_after_filename",
		specific.GitSupportMatch(gitFlagAfterFilenameMatch),
		specific.GitSupportNew(gitFlagAfterFilenameNewCommand)))
}

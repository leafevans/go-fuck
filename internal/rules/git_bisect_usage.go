package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitBisectUsageBroken matches the mistyped bisect subcommand.
var gitBisectUsageBroken = regexp.MustCompile("git bisect ([^ $]*).*")

// gitBisectUsageSubcommands matches the list of subcommands in the usage line.
var gitBisectUsageSubcommands = regexp.MustCompile(`usage: git bisect \[([^\]]+)\]`)

// gitBisectUsageMatch reports whether git printed the bisect usage because of
// a mistyped subcommand.
func gitBisectUsageMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "bisect") &&
		strings.Contains(command.Output, "usage: git bisect")
}

// gitBisectUsageNewCommand replaces the subcommand with its closest matches.
func gitBisectUsageNewCommand(command *types.Command) ([]string, error) {
	brokenMatch := gitBisectUsageBroken.FindStringSubmatch(command.Script)
	usageMatch := gitBisectUsageSubcommands.FindStringSubmatch(command.Output)
	if brokenMatch == nil || usageMatch == nil {
		return nil, nil
	}
	subcommands := strings.Split(usageMatch[1], "|")
	return utils.ReplaceCommand(command, brokenMatch[1], subcommands), nil
}

func init() {
	register(types.NewRule("git_bisect_usage",
		specific.GitSupportMatch(gitBisectUsageMatch),
		specific.GitSupportNew(gitBisectUsageNewCommand)))
}
